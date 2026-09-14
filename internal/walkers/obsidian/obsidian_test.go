package obsidian

import (
	"os"
	"path/filepath"
	"testing"
)

func getWD() string {
	wd, _ := os.Getwd()
	return wd
}

// //go:embed testdata
// var testdataFS embed.FS
func Test_Obsidian_Walk(t *testing.T) {
	wd := filepath.Join(getWD(), "./testdata/")
	// logger.LogLevel.Set(slog.LevelDebug)
	docs, err := Walk(wd)

	if err != nil {
		t.Fatalf("could not walk %q: %+v", wd, err)
	}

	t.Logf("%+v", docs)
}
