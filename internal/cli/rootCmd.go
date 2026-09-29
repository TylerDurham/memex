package cli

import (
	"os"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/config"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/spf13/cobra"
)

// verbose is set by the persistent -v/--verbose flag and is available to every subcommand.
var verbose bool
var formatJSON bool

// configDirFlag is set by the persistent --config-dir flag. Use configDir() to read the
// effective config directory rather than this variable.
var configDirFlag string

// configDir returns the config directory, preferring --config-dir, then
// $MEMEX_CONFIG_DIR, then the default from config.Dir.
func configDir() (string, error) {
	if configDirFlag != "" {
		return configDirFlag, nil
	}
	return config.Dir()
}

// LogFile is the name of the log file in the config directory.
const LogFile = "memex.log"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memex",
	Short: "A simple semantic indexing system.",
	Long:  `// TODO: Replace with longer description.`,
	// Log to <config dir>/memex.log as well as the console. A log file that can't be
	// opened isn't worth failing the command over, so it's only a warning.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, err := configDir()
		if err != nil {
			return err
		}
		if err := logger.OpenFile(filepath.Join(cfgDir, LogFile)); err != nil {
			logger.Warn("warning: could not open log file: %v\n", err)
		}
		return nil
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	logger.Close()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.memex.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVar(&configDirFlag, "config-dir", "", "config directory (default is $"+config.EnvConfigDir+", else $XDG_CONFIG_HOME/memex or ~/.config/memex, or %AppData%\\memex on Windows)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
