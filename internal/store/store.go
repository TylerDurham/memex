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

// File is a row in the files table: one indexed document.
type File struct {
	Path        string // vault-relative path
	Application string // provider that indexed it, e.g. "obsidian"
	URI         string // URI that opens the document in its application, if any
	Title       string // from frontmatter
	Description string // from frontmatter
	ContentHash string
	ModTime     int64
}

// Result is a single search hit.
type Result struct {
	Chunk
	Application string  // from the chunk's file record
	URI         string  // from the chunk's file record
	Title       string  // from the chunk's file record
	Description string  // from the chunk's file record
	Score       float32 // cosine similarity, 1.0 = identical
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
			title         TEXT,
			description   TEXT,
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

// ReplaceFile atomically swaps all chunks for f.Path with newChunks and
// upserts the file record. Runs in a transaction so a crash mid-index never
// leaves stale + fresh chunks for the same file coexisting.
func (s *Store) ReplaceFile(ctx context.Context, f File, newChunks []Chunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE file_path = ?`, f.Path); err != nil {
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
			f.Path, c.Heading, c.HeadingPath, c.Content, c.StartLine,
			f.ContentHash, f.ModTime, encodeVector(c.Embedding),
		); err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO files (file_path, application, uri, title, description, content_hash, mod_time, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(file_path) DO UPDATE SET
			application  = excluded.application,
			uri          = excluded.uri,
			title        = excluded.title,
			description  = excluded.description,
			content_hash = excluded.content_hash,
			mod_time     = excluded.mod_time,
			indexed_at   = excluded.indexed_at
	`, f.Path, f.Application, f.URI, f.Title, f.Description, f.ContentHash, f.ModTime); err != nil {
		return fmt.Errorf("upsert file record: %w", err)
	}

	return tx.Commit()
}

// UpdateFileInfo sets a file record's application, URI, title and
// description without touching its chunks, for files whose embedded content
// is unchanged (e.g. only the frontmatter was edited). It writes only when a
// value actually differs, and does nothing for unknown paths.
func (s *Store) UpdateFileInfo(ctx context.Context, f File) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE files SET application = ?1, uri = ?2, title = ?3, description = ?4
		WHERE file_path = ?5 AND (application IS NOT ?1 OR uri IS NOT ?2
			OR title IS NOT ?3 OR description IS NOT ?4)
	`, f.Application, f.URI, f.Title, f.Description, f.Path)
	if err != nil {
		return fmt.Errorf("update file info: %w", err)
	}
	return nil
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
	results, err := s.scoreAll(ctx, query, minScore)
	if err != nil {
		return nil, err
	}
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// SearchFiles is Search grouped by file: it returns the topK most similar
// files, each represented by its best-scoring chunk, so one long note can't
// crowd every other note out of the results.
func (s *Store) SearchFiles(ctx context.Context, query []float32, topK int, minScore float32) ([]Result, error) {
	results, err := s.scoreAll(ctx, query, minScore)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var out []Result
	for _, r := range results {
		if seen[r.FilePath] {
			continue
		}
		seen[r.FilePath] = true
		out = append(out, r)
		if topK > 0 && len(out) == topK {
			break
		}
	}
	return out, nil
}

// scoreAll scores every chunk against query and returns those at or above
// minScore, best first.
func (s *Store) scoreAll(ctx context.Context, query []float32, minScore float32) ([]Result, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.file_path, c.heading, c.heading_path, c.content, c.start_line,
		       c.content_hash, c.mod_time, c.embedding,
		       COALESCE(f.application, ''), COALESCE(f.uri, ''),
		       COALESCE(f.title, ''), COALESCE(f.description, '')
		FROM chunks c
		LEFT JOIN files f ON f.file_path = c.file_path
	`)
	if err != nil {
		return nil, fmt.Errorf("query chunks: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		var blob []byte
		if err := rows.Scan(&r.ID, &r.FilePath, &r.Heading, &r.HeadingPath,
			&r.Content, &r.StartLine, &r.ContentHash, &r.ModTime, &blob,
			&r.Application, &r.URI, &r.Title, &r.Description); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		r.Embedding = decodeVector(blob)

		r.Score = cosineSimilarity(query, r.Embedding)
		if r.Score >= minScore {
			results = append(results, r)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
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
