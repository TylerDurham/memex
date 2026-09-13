package store

import (
	"log"
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/globals"
)

func cleanup(t *testing.T, app globals.App) {
	t.Logf("removing repo at %q", app.Config.ConfigDirectory())
	os.RemoveAll(app.Config.ConfigDirectory())
}

func setup(t *testing.T) (globals.App)  {
	tmp, _ := os.MkdirTemp("/tmp", "memex-test-*")
	os.Setenv("MEMEX_CONFIG_DIR", tmp)
	app, err := globals.InitApp()

	if err != nil {
		log.Fatalf("setup failed: %v", err)
		os.Exit(1)
	}

	t.Logf("created repo at %q", tmp)

	return app
}

func Test_Store_Init(t *testing.T) {

	app := setup(t)

	defer cleanup(t, app)

	defer os.RemoveAll(app.Config.ConfigDirectory())
	t.Logf("%s", app.Config.ConfigDirectory())

	store, err := Init(app, "foo-test")

	if err != nil {
		log.Fatalf("Init() failed: %+v", err)
		return
	}

	defer store.Close()

	
	
}



