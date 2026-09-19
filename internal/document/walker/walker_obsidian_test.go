package walker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getCWD() string {
	wd, _ := os.Getwd()
	return wd
}

func Test_Obsidian_Walk(t *testing.T) {
	tData := filepath.Join(getCWD(), "./testdata/obsidian/")
	t.Logf("loading test files from '%q'", tData)
	w, _ := NewWalker("obsidian", document.Unspecified)
	docs, err := w.Walk(tData)

	if err != nil {
		t.Fatalf("could not walk %q: %+v", tData, err)
	}

	logger.Debug("%+v", docs)
	assert.NotNilf(t, docs, "nothing returned")
	assert.GreaterOrEqual(t, len(docs), 1, "no docs walked")

	for _, d := range docs {
		require.FileExists(t, d.DocPath)
		info, _ := os.Stat(d.DocPath)
		assert.Equal(t, info.Size(), d.Size)
		assert.Equal(t, info.ModTime(), d.ModTime)
	}
}
