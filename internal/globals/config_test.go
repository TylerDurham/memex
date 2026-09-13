package globals

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The *FromEnv tests below call os.Setenv(EnvConfigDir, ...) and never
// restore the original value, so the override leaks into every test that
// runs after them. That's safe here only because every assertion computes
// its "expected" value via GetConfigDir/GetLogDir at test time rather than
// hardcoding the default — but it does mean Test_Config_GetConfigDir (the
// one test that assumes no override is set) must keep running before any
// *FromEnv test.

// Test_Config_GetConfigDir verifies that GetConfigDir falls back to
// ~/.config/<AppName> when EnvConfigDir is not set.
func Test_Config_GetConfigDir(t *testing.T) {
	var (
		home     string
		actual   string
		expected string
		err      error
	)

	home, err = os.UserHomeDir()
	assert.Nil(t, err, "Could not get user home directory")

	expected = filepath.Join(home, ".config", AppName)
	actual = GetConfigDir()
	assert.Equal(t, expected, actual)

}

// Test_Config_GetConfigDirFromEnv verifies that GetConfigDir returns the
// value of EnvConfigDir directly when that environment variable is set,
// overriding the default ~/.config/<AppName> path.
func Test_Config_GetConfigDirFromEnv(t *testing.T) {

	var (
		expected string
		actual   string
		err      error
	)

	expected, err = os.MkdirTemp("/tmp", "memex-test-*")
	assert.Nilf(t, err, "could not create temp directory: %+v", err)

	// Set the env variable, call api, and assert
	err = os.Setenv(EnvConfigDir, expected)
	assert.Nilf(t, err, "could not set env[%s]", EnvConfigDir)
	actual = GetConfigDir()
	assert.Equal(t, expected, actual)

}

// Test_Config_GetLogDir verifies that GetLogDir returns a "logs"
// subdirectory of whatever GetConfigDir currently resolves to.
func Test_Config_GetLogDir(t *testing.T) {
	var expected = GetConfigDir() + "/logs"
	var actual = GetLogDir()
	assert.Equal(t, expected, actual)
}

// Test_Config_GetLogDirFromEnv verifies that GetLogDir tracks GetConfigDir
// when EnvConfigDir is set, i.e. the log dir moves along with the config dir
// override rather than staying pinned to the default.
func Test_Config_GetLogDirFromEnv(t *testing.T) {

	var expected, err = os.MkdirTemp("/tmp", "memex-test-*")
	assert.Nilf(t, err, "could not create temp directory: %+v", err)

	// Set the env variable, call api, and assert
	err = os.Setenv(EnvConfigDir, expected)
	assert.Nilf(t, err, "could not set env[%s]", EnvConfigDir)

	var actual = GetLogDir()
	assert.Equal(t, expected+"/logs", actual)

}

// Test_Config_NewConfig verifies that the ConfigInfo returned by NewConfig
// reports the same paths GetConfigDir/GetLogDir compute directly, and that
// the log directory it's responsible for creating actually exists on disk
// afterward.
func Test_Config_NewConfig(t *testing.T) {
	var (
		cfg      ConfigInfo // ConfigInfo instance
		err      error      // error
		expected string     // expected value
		actual   string     // actual value
	)

	cfg, err = NewConfig()
	assert.Nil(t, err, "factory func returned error")

	expected = GetConfigDir()
	actual = cfg.ConfigDirectory()
	assert.Equal(t, expected, actual)

	expected = GetLogDir()
	actual = cfg.LogDirectory()
	assert.Equal(t, expected, actual)

	require.DirExists(t, actual)
}

// Test_Config_NewConfigFromEnv repeats Test_Config_NewConfig with
// EnvConfigDir set, confirming NewConfig respects the override end-to-end
// (both the reported paths and the directory it creates on disk).
func Test_Config_NewConfigFromEnv(t *testing.T) {
	var (
		cfg      ConfigInfo // ConfigInfo instance
		err      error      // error
		expected string     // expected value
		actual   string     // actual value
	)

	expected, err = os.MkdirTemp("/tmp", "memex-test-*")
	assert.Nilf(t, err, "could not create temp directory: %+v", err)

	// Set the env variable, call api, and assert
	err = os.Setenv(EnvConfigDir, expected)
	assert.Nilf(t, err, "could not set env[%s]", EnvConfigDir)

	cfg, err = NewConfig()
	assert.Nil(t, err, "factory func returned error")

	// Ensure config dir matches
	expected = GetConfigDir()
	actual = cfg.ConfigDirectory()
	assert.Equal(t, expected, actual)

	// Ensure log dir matches
	expected = GetLogDir()
	actual = cfg.LogDirectory()
	assert.Equal(t, expected, actual)

	require.DirExists(t, actual)

}
