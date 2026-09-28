package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/spf13/cobra"
)

type repoRemoveOptions struct {
	yes bool
}

func newRepoRemoveCmd() *cobra.Command {
	var opts = repoRemoveOptions{}

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		Short:   "Remove a repo's settings and index.",
		Long: "Remove a repo: delete its settings and its index. The repository " +
			"directory itself (e.g. the Obsidian vault) is not touched.",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			out := cmd.OutOrStdout()
			dir := config.RepoDir(config.ConfigDir(), name)

			if !opts.yes {
				prompt := fmt.Sprintf("Remove repo %q and its index (%s)?", name, dir)
				// A repo with no config is still removable; there's just no
				// directory to mention.
				if repo, err := config.LoadRepo(name); err == nil {
					prompt += fmt.Sprintf("\n%s is not touched.", repo.Directory)
				} else if !errors.Is(err, config.ErrRepoNotFound) {
					return err
				}

				ok, err := confirm(cmd.InOrStdin(), out, prompt)
				if err != nil {
					return err
				}
				if !ok {
					return errors.New("not removed")
				}
			}

			if err := config.RemoveRepo(name); err != nil {
				return err
			}
			fmt.Fprintf(out, "removed repo %q\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "Don't ask for confirmation.")

	return cmd
}

// confirm prints prompt and reads one line from in. Only "y" or "yes"
// (any case) confirms; anything else, including no input, declines.
func confirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	fmt.Fprintf(out, "%s [y/N] ", prompt)

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	if errors.Is(err, io.EOF) && line == "" {
		fmt.Fprintln(out) // keep the next output off the prompt line
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	}
	return false, nil
}
