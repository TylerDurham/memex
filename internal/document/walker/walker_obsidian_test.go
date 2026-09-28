package walker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getCWD gets the working directory starting at the project root.
func getCWD() string {
	wd, _ := os.Getwd()
	return wd
}

func Test_Obsidian_Walk(t *testing.T) {
	repoPath := filepath.Join(getCWD(), "./testdata/obsidian/")
	t.Logf("loading test files from '%q'", repoPath)
	walker, _ := NewWalker("obsidian", document.Unspecified)

	docs, err := walker.Walk(repoPath)
	if err != nil {
		t.Fatalf("could not walk %q: %+v", repoPath, err)
	}

	logger.Debug("%+v", docs)
	assert.NotNilf(t, docs, "nothing returned")
	assert.GreaterOrEqual(t, len(docs), 1, "no docs walked")

	for _, d := range docs {
		require.FileExists(t, d.DocPath)
		info, _ := os.Stat(d.DocPath)
		relPath, _ := filepath.Rel(repoPath, d.DocPath)
		assert.Equal(t, filepath.Ext(d.DocPath), d.Extension, "Extension")
		assert.Equal(t, info.ModTime(), d.ModTime, "ModTime")
		assert.Equal(t, info.Size(), d.Size, "Size")
		assert.Equal(t, relPath, d.RelPath, "RelPath")
		assert.Equal(t, walker.handler.AppName(), d.Application, "Application")

		// TODO: This URL check needs to be 'stronger'
		// Ensure URI starts with app name. Blank line means the app doesn't 
		// support document url/uri schemes
		assert.True(t, strings.HasPrefix(d.URI, walker.handler.AppName()), "URL")
	}
}

func Test_Obsidian_Walk_WithContent(t *testing.T) {
	repoPath := filepath.Join(getCWD(), "./testdata/obsidian/")
	walker, err := NewWalker("obsidian", document.IncludeProperties|document.IncludeChunks)
	require.NoError(t, err)

	docs, err := walker.Walk(repoPath)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	for _, d := range docs {
		assert.NotEmptyf(t, d.Chunks, "%s: no chunks", d.RelPath)
		for _, c := range d.Chunks {
			assert.NotEmptyf(t, c.Text, "%s: empty chunk", d.RelPath)
			assert.LessOrEqualf(t, c.StartLine, c.EndLine, "%s: line range", d.RelPath)
		}
	}

	// Every gear spec note carries a title in its frontmatter.
	for _, d := range docs {
		if strings.HasSuffix(d.RelPath, "-specs.md") {
			assert.NotEmptyf(t, d.Properties["title"], "%s: missing title", d.RelPath)
		}
	}
}

// writeFiles creates each path (relative to root) with a small note.
func writeFiles(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		full := filepath.Join(root, p)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o750))
		require.NoError(t, os.WriteFile(full, []byte("# Note\n\nBody.\n"), 0o644))
	}
}

func relPaths(docs []document.IndexDocument) []string {
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.RelPath
	}
	return out
}

func Test_Walk_SkipsHidden(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root,
		"note.md",
		"sub/visible.md",
		"sub/.hidden-note.md",
		".obsidian/plugins/some-plugin/README.md",
		".trash/deleted.md",
		".opencode/node_modules/pkg/README.md",
		"sub/.git/notes.md",
	)

	w, err := NewWalker("obsidian", document.Unspecified)
	require.NoError(t, err)
	docs, err := w.Walk(root)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"note.md", filepath.Join("sub", "visible.md")}, relPaths(docs))
}

func Test_Walk_HiddenRootIsWalked(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".notes")
	writeFiles(t, root, "note.md", ".obsidian/README.md")

	w, err := NewWalker("obsidian", document.Unspecified)
	require.NoError(t, err)
	docs, err := w.Walk(root)
	require.NoError(t, err)

	assert.Equal(t, []string{"note.md"}, relPaths(docs))
}
