package cli

import (
	"fmt"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/embed"
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/spf13/cobra"
)

// repoOptions are the flags shared by commands that work on an indexed
// repository: where it lives, which index it maps to, and how to embed.
type repoOptions struct {
	directory string
	name      string
	ollamaURL string
	model     string
}

func (o *repoOptions) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.directory, "directory", "d", ".", "The repository (e.g. Obsidian vault) directory.")
	cmd.Flags().StringVarP(&o.name, "name", "n", "", "The index name. Defaults to the directory's base name.")
	cmd.Flags().StringVar(&o.ollamaURL, "ollama-url", config.OllamaURL(), "Ollama server URL ($"+config.EnvOllamaURL+").")
	cmd.Flags().StringVar(&o.model, "model", config.OllamaModel(), "Embedding model ($"+config.EnvOllamaModel+").")
}

// resolve makes directory absolute and fills in the default name.
func (o *repoOptions) resolve() error {
	dir, err := filepath.Abs(o.directory)
	if err != nil {
		return fmt.Errorf("could not resolve directory %q: %w", o.directory, err)
	}
	o.directory = dir
	if o.name == "" {
		o.name = filepath.Base(dir)
	}
	return nil
}

func (o *repoOptions) embedder() *embed.Ollama {
	return embed.NewOllama(o.ollamaURL, o.model)
}
