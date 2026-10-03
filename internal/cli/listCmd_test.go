package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TylerDurham/memex/internal/repo"
)

// initRepos creates a repo named after each name in configDir, each with its own directory.
func initRepos(t *testing.T, configDir string, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := repo.InitWithConfigDir(configDir, repo.InitOptions{Directory: t.TempDir(), Name: name, Application: "obsidian"}); err != nil {
			t.Fatalf("Init(%q): %v", name, err)
		}
	}
}

func TestListNoRepos(t *testing.T) {
	configDir := t.TempDir()

	out, _, err := runCmd(t, "list", "--config-dir", configDir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// "no repos found" is a warning on the console log, not list output.
	if out != "" {
		t.Errorf("list output = %q, want none", out)
	}
	if log := readLogFile(t, configDir); !strings.Contains(log, `"level":"WARN","msg":"no repos found"`) {
		t.Errorf("log file has no no-repos warning:\n%s", log)
	}
}

func TestListNoReposJSON(t *testing.T) {
	out, _, err := runCmd(t, "list", "--config-dir", t.TempDir(), "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("list --json output = %q, want []", out)
	}
}

func TestListTable(t *testing.T) {
	configDir := t.TempDir()
	initRepos(t, configDir, "zeta", "alpha")

	// The ls alias runs the same command.
	for _, cmd := range []string{"list", "ls"} {
		out, _, err := runCmd(t, cmd, "--config-dir", configDir)
		if err != nil {
			t.Fatalf("%s: %v", cmd, err)
		}
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) != 3 {
			t.Fatalf("%s output has %d lines, want header and 2 repos:\n%s", cmd, len(lines), out)
		}
		if fields := strings.Fields(lines[0]); strings.Join(fields, " ") != "NAME APPLICATION DIRECTORY CONFIG DB" {
			t.Errorf("%s header = %q", cmd, lines[0])
		}
		for i, name := range []string{"alpha", "zeta"} {
			fields := strings.Fields(lines[i+1])
			if len(fields) != 5 || fields[0] != name || fields[1] != "obsidian" {
				t.Errorf("%s row %d = %q, want repo %s", cmd, i+1, lines[i+1], name)
			}
		}
	}
}

func TestListJSON(t *testing.T) {
	configDir := t.TempDir()
	initRepos(t, configDir, "zeta", "alpha")

	out, _, err := runCmd(t, "list", "--config-dir", configDir, "-j")
	if err != nil {
		t.Fatalf("list -j: %v", err)
	}
	var repos []repo.RepoConfigInfo
	if err := json.Unmarshal([]byte(out), &repos); err != nil {
		t.Fatalf("list -j output isn't a JSON array of repos: %v\n%s", err, out)
	}
	if len(repos) != 2 || repos[0].Name != "alpha" || repos[1].Name != "zeta" {
		t.Fatalf("list -j = %+v, want repos alpha and zeta", repos)
	}
	want := filepath.Join(configDir, repo.ReposDir, "alpha", repo.ConfigFile)
	if repos[0].ConfigFile != want {
		t.Errorf("alpha config = %q, want %q", repos[0].ConfigFile, want)
	}
}

func TestListSkipsBrokenRepo(t *testing.T) {
	configDir := t.TempDir()
	initRepos(t, configDir, "good", "broken")
	path := filepath.Join(configDir, repo.ReposDir, "broken", repo.ConfigFile)
	if err := os.WriteFile(path, []byte("name: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, _, err := runCmd(t, "list", "--config-dir", configDir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// The console handler writes to the real stderr, so check the log file, which
	// gets the same records.
	if log := readLogFile(t, configDir); !strings.Contains(log, `"level":"WARN","msg":"could not load repo","repo":"broken"`) {
		t.Errorf("log file has no warning about repo broken:\n%s", log)
	}
	if !strings.Contains(out, "good") || strings.Contains(out, "broken") {
		t.Errorf("list output should show good but not broken:\n%s", out)
	}
}

func TestListRejectsArgs(t *testing.T) {
	if _, _, err := runCmd(t, "list", "--config-dir", t.TempDir(), "extra"); err == nil {
		t.Error("list with an argument succeeded, want error")
	}
}
