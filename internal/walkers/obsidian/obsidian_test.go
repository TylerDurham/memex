package obsidian

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getWD() string {
	wd, _ := os.Getwd()
	return wd
}

func Test_Obsidian_Walk(t *testing.T) {
	wd := filepath.Join(getWD(), "./testdata/")
	// logger.LogLevel.Set(slog.LevelDebug)
	docs, err := Walk(wd)

	if err != nil {
		t.Fatalf("could not walk %q: %+v", wd, err)
	}

	assert.NotNilf(t, docs, "nothing returned")
	assert .GreaterOrEqual(t, len(docs), 1, "no docs walked")
	for _, d := range docs {
		t.Logf("%+v", d)
		require.FileExists(t, d.AbsPath)
		s, _ := os.Stat(d.AbsPath)
		assert.Equal(t, s.Size(), d.Size)
		assert.Equal(t, s.ModTime(), d.ModTime)
	}
}
