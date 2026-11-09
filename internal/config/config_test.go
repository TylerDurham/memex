package config_test

import (
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/config"
	"github.com/TylerDurham/memex/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	testutil.IsolateConfig(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("could not load config: %v", err)
	}
	path := cfg.AbsPath
	if err != nil {
		t.Fatalf("could not determine MemexConfigDir: %v", err)
	}

	assert.Equal(t, path, os.Getenv(config.EnvMemexConfigHome))
}
