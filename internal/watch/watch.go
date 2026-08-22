// Package watch triggers incremental re-indexing when files in the vault
// change, debounced so a burst of saves (or a sync client writing several
// files at once) doesn't trigger a re-index per event.
package watch

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/index"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/fsnotify/fsnotify"
)

// Debounce is how long to wait after the last filesystem event before
// running an index pass. Chosen to comfortably absorb editor autosave
// bursts and multi-file sync writes without feeling laggy.
const Debounce = 2 * time.Second

// Run watches opts.VaultRoot recursively and re-indexes on change until ctx
// is cancelled.
func Run(ctx context.Context, st *store.Store, emb *embed.Client, opts index.Options, log *slog.Logger) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	if err := addRecursive(w, opts.VaultRoot); err != nil {
		return err
	}

	log.Info("watching vault for changes", "path", opts.VaultRoot)

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	pending := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			if !relevant(ev) {
				continue
			}
			pending = true
			timer.Reset(Debounce)

			// New directories (e.g. a note moved into a fresh subfolder)
			// need to be watched too.
			if ev.Op&fsnotify.Create != 0 {
				_ = w.Add(ev.Name) // best-effort; fails silently for files
			}

		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			log.Warn("watcher error", "err", err)

		case <-timer.C:
			if !pending {
				continue
			}
			pending = false
			log.Info("re-indexing after change burst")
			stats, err := index.Run(ctx, st, emb, opts, log)
			if err != nil {
				log.Error("index run failed", "err", err)
				continue
			}
			log.Info("index pass complete",
				"changed", stats.FilesChanged,
				"skipped", stats.FilesSkipped,
				"deleted", stats.FilesDeleted,
				"chunks_written", stats.ChunksWritten,
			)
		}
	}
}

func relevant(ev fsnotify.Event) bool {
	if !strings.EqualFold(filepath.Ext(ev.Name), ".md") {
		// Still relevant if it's a directory create, so we can watch it.
		return ev.Op&fsnotify.Create != 0
	}
	return ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0
}

func addRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".obsidian" || base == ".trash" || strings.HasPrefix(base, ".") && path != root {
				return filepath.SkipDir
			}
			return w.Add(path)
		}
		return nil
	})
}
