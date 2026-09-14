package config

import (
	"os"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempDir() string {
	tmp, _ := os.MkdirTemp("/tmp", "memex*")

	return tmp
}

type property = func() string

func testDir(t *testing.T, expected string, callback property) {
	actual := callback()
	t.Logf("Checking dir: %q", expected)
	assert.NotEmpty(t, actual, "dir is empty")
	require.DirExists(t, actual, "dir does not exist")
	assert.Equal(t, expected, actual)
	assert.Equal(t, unsafe.StringData(expected), unsafe.StringData(actual), "strings have been copied")
}

func Test_Config_Init(t *testing.T) {
	expected := ConfigDir()
	testDir(t, expected, ConfigDir)
	testDir(t, expected+"/logs", LogDir)

	// Set MEMEX_CONFIG_DIR and ensure it works
	tmp := tempDir()
	os.Setenv(EnvConfigDir, tmp)
	reset()
	testDir(t, tmp, ConfigDir)
	testDir(t, tmp+"/logs", LogDir)
}
