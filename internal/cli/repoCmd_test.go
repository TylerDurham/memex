package cli

import (
	"strings"
	"testing"
)

func TestRepoRegistersInit(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"repo", "init"})
	if err != nil {
		t.Fatalf("Find repo init: %v", err)
	}
	if cmd != repoInitCmd {
		t.Errorf("repo init resolves to %q, want repoInitCmd", cmd.CommandPath())
	}
}

func TestRepoShowsHelp(t *testing.T) {
	out, _, err := runCmd(t, "repo")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	if !strings.Contains(out, "Work with a memex repository") || !strings.Contains(out, "init") {
		t.Errorf("repo output isn't its help listing init:\n%s", out)
	}
}
