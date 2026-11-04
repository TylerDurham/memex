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
	path, err := config.MemexConfigDir()
	if err != nil {
		t.Fatalf("could not determine MemexConfigDir: %v", err)
	}

	assert.Equal(t, path, os.Getenv(config.EnvMemexConfigDir))
}
