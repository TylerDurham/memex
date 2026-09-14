package store

import (
	"log"
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/globals/config"
)

func cleanup(t *testing.T) {
	t.Logf("removing repo at %q", config.ConfigDir())
	os.RemoveAll(config.ConfigDir())
}

func setup(t *testing.T) {
	tmp, _ := os.MkdirTemp("/tmp", "memex-test-*")
	os.Setenv("MEMEX_CONFIG_DIR", tmp)

	t.Logf("created repo at %q", tmp)

}

func Test_Store_Init(t *testing.T) {

	setup(t)

	defer cleanup(t)

	defer os.RemoveAll(config.ConfigDir())
	t.Logf("%s", config.ConfigDir())

	store, err := Init("foo-test")

	if err != nil {
		log.Fatalf("Init() failed: %+v", err)
		return
	}

	defer store.Close()

}
