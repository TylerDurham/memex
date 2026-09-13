// Package store persists chunks and their embeddings, and provides
// similarity search over them.
//
// Deliberately not using a vector-search SQLite extension (sqlite-vec/vss)
// for the KNN step: that requires loading a *second* native extension on
// top of the driver itself. Instead embeddings are stored as raw float32
// blobs and search is brute-force cosine similarity done in Go. For a
// vault-sized corpus (low thousands of chunks, ~768 dims) this is well
// under 50ms.
//
// Uses mattn/go-sqlite3, which is cgo-based (needs gcc at build time, same
// as any other native-extension-capable SQLite driver). If a fully static,
// cgo-free binary matters more to you than sqlite-vec-readiness, swap this
// import for modernc.org/sqlite (pure Go, drop-in — same database/sql
// interface, just change the driver import and DSN string) and cross-
// compilation gets simpler at the cost of losing the option to load native
// SQLite extensions later.
package store

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	_ "github.com/mattn/go-sqlite3" // cgo driver; see package doc for the pure-Go alternative
)

type Store struct {
	db *sql.DB
}

// Chunk mirrors chunker.Chunk plus the fields needed for persistence and
// change detection.
type Chunk struct {
	ID          int64
	FilePath    string // vault-relative path
	Heading     string
	HeadingPath string
	Content     string
	StartLine   int
	ContentHash string
	ModTime     int64
	Embedding   []float32
}

// Result is a single search hit.
type Result struct {
	Chunk
	Score float32 // cosine similarity, 1.0 = identical
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")

	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Vault indexing is a single-writer workload; a small pool avoids
	// SQLITE_BUSY under WAL without needing an explicit mutex.
	db.SetMaxOpenConns(1)

	s := &Store{db: db}

	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS chunks (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path     TEXT NOT NULL,
			heading       TEXT,
			heading_path  TEXT,
			content       TEXT NOT NULL,
			start_line    INTEGER,
			content_hash  TEXT NOT NULL,
			mod_time      INTEGER NOT NULL,
			embedding     BLOB NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_chunks_file_path ON chunks(file_path);

		CREATE TABLE IF NOT EXISTS files (
			file_path     TEXT PRIMARY KEY,
			application   TEXT,
			uri           TEXT,
			content_hash  TEXT NOT NULL,
			mod_time      INTEGER NOT NULL,
			indexed_at    INTEGER NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// FileHash returns the last-indexed content hash for a file, or "" if the
// file has never been indexed. Used by the indexer to skip unchanged files.
func (s *Store) FileHash(ctx context.Context, relPath string) (string, error) {
	var hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT content_hash FROM files WHERE file_path = ?`, relPath,
	).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query file hash: %w", err)
	}
	return hash, nil
}

// ReplaceFile atomically swaps all chunks for relPath with newChunks and
// records the new content hash. Runs in a transaction so a crash mid-index
// never leaves stale + fresh chunks for the same file coexisting.
func (s *Store) ReplaceFile(ctx context.Context, relPath, contentHash string, modTime int64, newChunks []Chunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE file_path = ?`, relPath); err != nil {
		return fmt.Errorf("delete old chunks: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO chunks (file_path, heading, heading_path, content, start_line, content_hash, mod_time, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, c := range newChunks {
		if _, err := stmt.ExecContext(ctx,
			relPath, c.Heading, c.HeadingPath, c.Content, c.StartLine,
			contentHash, modTime, encodeVector(c.Embedding),
		); err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO files (file_path, content_hash, mod_time, indexed_at)
		VALUES (?, ?, ?, unixepoch())
		ON CONFLICT(file_path) DO UPDATE SET
			content_hash = excluded.content_hash,
			mod_time     = excluded.mod_time,
			indexed_at   = excluded.indexed_at
	`, relPath, contentHash, modTime); err != nil {
		return fmt.Errorf("upsert file record: %w", err)
	}

	return tx.Commit()
}

// RemoveFile deletes all chunks and file record for a path that no longer
// exists in the vault (e.g. the note was deleted or renamed).
func (s *Store) RemoveFile(ctx context.Context, relPath string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE file_path = ?`, relPath); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE file_path = ?`, relPath); err != nil {
		return err
	}
	return tx.Commit()
}

// KnownFiles returns all file paths currently tracked in the index, used by
// the indexer to detect deletions (files present in the DB but no longer on
// disk).
func (s *Store) KnownFiles(ctx context.Context) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT file_path FROM files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out[p] = true
	}
	return out, rows.Err()
}

// Search embeds nothing itself — callers pass an already-embedded query
// vector — and returns the topK most similar chunks by cosine similarity,
// filtered to those scoring at or above minScore.
func (s *Store) Search(ctx context.Context, query []float32, topK int, minScore float32) ([]Result, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, file_path, heading, heading_path, content, start_line, content_hash, mod_time, embedding
		FROM chunks
	`)
	if err != nil {
		return nil, fmt.Errorf("query chunks: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var c Chunk
		var blob []byte
		if err := rows.Scan(&c.ID, &c.FilePath, &c.Heading, &c.HeadingPath,
			&c.Content, &c.StartLine, &c.ContentHash, &c.ModTime, &blob); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		c.Embedding = decodeVector(blob)

		score := cosineSimilarity(query, c.Embedding)
		if score >= minScore {
			results = append(results, Result{Chunk: c, Score: score})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// Count returns the total number of indexed chunks and files, for `status`.
func (s *Store) Count(ctx context.Context) (chunks, files int, err error) {
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks`).Scan(&chunks); err != nil {
		return 0, 0, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM files`).Scan(&files); err != nil {
		return 0, 0, err
	}
	return chunks, files, nil
}

// --- vector encoding helpers ---

func encodeVector(v []float32) []byte {
	buf := make([]byte, 4*len(v))
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

func decodeVector(b []byte) []float32 {
	n := len(b) / 4
	v := make([]float32, n)
	for i := 0; i < n; i++ {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, magA, magB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		magA += float64(a[i]) * float64(a[i])
		magB += float64(b[i]) * float64(b[i])
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(magA) * math.Sqrt(magB)))
}
