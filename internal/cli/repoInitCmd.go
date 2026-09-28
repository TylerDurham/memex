package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/document/walker"
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/store"
	"github.com/spf13/cobra"
)

type initOptions struct {
	directory string
	app       string
	ollamaURL string
	model     string
}

func newInitCmd() *cobra.Command {

	var opts = initOptions{}

	cmd := &cobra.Command{
		Use:     "init <name>",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Short:   "Initialize a new memex repo, or update an existing one's settings.",
		Long: "Initialize a new memex repo: save its settings to " +
			filepath.Join("<config dir>", config.ReposDirName, "<name>", config.RepoConfigFile) +
			" and create its index. Other commands then only need the repo name.\n\n" +
			"Running init on an existing repo updates only the flags given.",
		Example: "  memex repo init tech-kasten --directory ~/vaults/Tech-Kasten\n" +
			"  memex repo init tech-kasten --ollama-url https://ollama.example.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			logger.Debug("config: ", "ConfigDirectory", config.ConfigDir())

			repo, err := config.LoadRepo(name)
			isNew := errors.Is(err, config.ErrRepoNotFound)
			switch {
			case isNew:
				if !cmd.Flags().Changed("directory") {
					return fmt.Errorf("--directory is required for a new repo")
				}
				repo = &config.Repo{Name: name}
			case err != nil:
				return err
			}

			flags := cmd.Flags()
			oldModel := repo.Model
			if isNew || flags.Changed("directory") {
				repo.Directory = opts.directory
			}
			if isNew || flags.Changed("app") {
				repo.App = opts.app
			}
			if isNew || flags.Changed("ollama-url") {
				repo.OllamaURL = opts.ollamaURL
			}
			if isNew || flags.Changed("model") {
				repo.Model = opts.model
			}

			if err := validateRepo(repo); err != nil {
				return err
			}
			logger.Debug("repo: ", "name", repo.Name, "directory", repo.Directory,
				"app", repo.App, "ollamaURL", repo.OllamaURL, "model", repo.Model)

			if err := repo.Save(); err != nil {
				return err
			}

			store, err := store.Init(repo.Name)
			if err != nil {
				return err
			}
			store.Close()

			out := cmd.OutOrStdout()
			verb := "updated"
			if isNew {
				verb = "initialized"
			}
			fmt.Fprintf(out, "%s repo %q\n  directory:  %s\n  app:        %s\n  ollama url: %s\n  model:      %s\n",
				verb, repo.Name, repo.Directory, repo.App, repo.OllamaURL, repo.Model)
			if !isNew && oldModel != repo.Model {
				fmt.Fprintln(out, "model changed: run 'memex index' to re-embed documents indexed with a different model")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.directory, "directory", "d", "", "The repository (e.g. Obsidian vault) directory. Required for a new repo.")
	cmd.Flags().StringVarP(&opts.app, "app", "a", "obsidian", "The application the repository belongs to.")
	cmd.Flags().StringVar(&opts.ollamaURL, "ollama-url", config.OllamaURL(), "Ollama server URL ($"+config.EnvOllamaURL+").")
	cmd.Flags().StringVar(&opts.model, "model", config.OllamaModel(), "Embedding model ($"+config.EnvOllamaModel+").")

	return cmd

}

// validateRepo checks the repo's settings and makes Directory absolute.
func validateRepo(repo *config.Repo) error {
	dir, err := filepath.Abs(repo.Directory)
	if err != nil {
		return fmt.Errorf("could not resolve directory %q: %w", repo.Directory, err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}
	repo.Directory = dir

	if _, ok := walker.Registry()[repo.App]; !ok {
		return fmt.Errorf("app %q not supported", repo.App)
	}
	if repo.OllamaURL == "" || repo.Model == "" {
		return errors.New("--ollama-url and --model must not be empty")
	}
	return nil
}
