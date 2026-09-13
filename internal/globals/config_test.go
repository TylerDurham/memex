package globals

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Config_GetConfigDir tests the config dir api
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

// Test_Config_GetConfigDirFromEnv test the confi dir api when the
//
//	config dir ENV variable is set
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

// Test_Config_GetLogDir tests that the logs directory is within the config directory
func Test_Config_GetLogDir(t *testing.T) {
	var expected = GetConfigDir() + "/logs"
	var actual = GetLogDir()
	assert.Equal(t, expected, actual)
}

// Test_Config_GetLogDir tests that the logs directory is within the config directory when the
//
//	config dir ENV is set
func Test_Config_GetLogDirFromEnv(t *testing.T) {

	var expected, err = os.MkdirTemp("/tmp", "memex-test-*")
	assert.Nilf(t, err, "could not create temp directory: %+v", err)

	// Set the env variable, call api, and assert
	err = os.Setenv(EnvConfigDir, expected)
	assert.Nilf(t, err, "could not set env[%s]", EnvConfigDir)

	var actual = GetLogDir()
	assert.Equal(t, expected+"/logs", actual)

}

// Test_Config_NewConfig Tests that all config apis "line up" and don't provide conflicting values
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

// Test_Config_NewConfigFromEnv Tests that all config apis "line up" and don't provide conflicting values when setting the
//
//	config dir ENV variable
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
