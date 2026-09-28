# Memex Search for Obsidian

Search your vault by meaning from inside Obsidian, using a
[memex](../README.md) semantic index. Desktop only: the plugin runs the
`memex` CLI and reads its `search --json` output.

## Requirements

- The `memex` binary, built from this repo (`go build -o ~/go/bin/memex ./cmd`).
- A memex repo for the vault, indexed at least once:

  ```sh
  memex repo init my-vault --directory /path/to/vault
  memex index my-vault
  ```

  The repo's directory must be the vault's folder, so result paths match
  the vault's files.

## Build

From the repo root:

```sh
just plugin-build   # type-check, then bundle to main.js
just plugin-test    # unit tests for the memex runner
```

Or in this folder: `npm install`, then `npm run build`, `npm test`, or
`npm run dev` to rebuild on change.

## Install

From the repo root, build the plugin and symlink it into a vault's plugin
folder, then enable **Memex Search** under *Settings → Community plugins*:

```sh
just plugin-install /path/to/vault
# or set the vault once:
export OBSIDIAN_VAULT=/path/to/vault
just plugin-install
```

The files are symlinked, so later runs of `just plugin-install` (or
`just plugin-build`) update the plugin in place; reload Obsidian to pick
up changes.

In the plugin settings, set **Repo** to the memex repo name. If Obsidian
can't find `memex` on its `PATH` (common when launched from a desktop
menu), set **memex executable** to its full path.

## Use

Run **Memex Search: Search by meaning** from the command palette, or click
the ribbon icon. Results update as you type (after a short pause); press
↵ to open a note at the matching section, or ctrl/⌘ ↵ to open it in a new
tab.
