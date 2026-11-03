package config_test

import (
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	path, err := config.MemexConfigDir()
	if err != nil {
		t.Fatalf("could not determine MemexConfigDir: %v", err)
	}

	t.Logf("config dir retrieved: '%s'", path)

	assert.Equal(t, path, os.Getenv(config.EnvMemexConfigDir))
	
}
