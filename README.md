# memex

A local-first semantic search CLI for Obsidian vaults (or any directory of
markdown). Embeds note chunks via [Ollama](https://ollama.com), stores them
in SQLite, and searches by cosine similarity. No cloud API, no API key.

Named for the machine in Vannevar Bush's ["As We May
Think"](https://www.theatlantic.com/magazine/archive/1945/07/as-we-may-think/303881/)
(1945): a personal store of documents traversed by *associative trails*
rather than by index terms. Which is more or less what asking a vault a
question in your own words turns out to be.

Designed to be shelled out to from a thin Obsidian plugin, the way
[QMD](https://github.com/tobi/qmd) does — but pointed at Ollama instead of a
bundled model, so it fits alongside anything else already running on your
network.

## Build

```bash
go mod tidy
go build -o memex .
```

Requires `gcc` on your PATH (the SQLite driver, `mattn/go-sqlite3`, is
cgo-based). If you'd rather have a fully static, cgo-free binary and don't
need to load native SQLite extensions later, swap the import in
`internal/store/store.go` for `modernc.org/sqlite` — same `database/sql`
interface, one-line change, documented in that file's package comment.

## Usage

```bash
# Index a vault (only re-embeds files whose content actually changed)
memex index --vault ~/notes

# Force full re-index, e.g. after switching embedding models
memex index --vault ~/notes --force

# See what would change without embedding or writing anything
memex index --vault ~/notes --dry-run

# Search (flags first — see note below)
memex search --vault ~/notes "runway pressure"
memex search --vault ~/notes --json --limit 5 "runway pressure"

# Continuously re-index on file change (debounced 2s)
memex watch --vault ~/notes

# Check index size
memex status --vault ~/notes
```

By default it talks to Ollama at `https://ollama.snork.co` using the
`nomic-embed-text` model. Override either with flags — note they must come
*before* the search query, since Go's flag parser stops at the first
positional argument:

```bash
memex search --vault ~/notes \
  --ollama-url http://erebor.snork.co:11434 \
  --model nomic-embed-text \
  "..."
```

or via the environment, which is probably what you want in an Obsidian
plugin's shell-out call so the endpoint isn't hardcoded per-machine:

| Variable | Sets |
| --- | --- |
| `MEMEX_OLLAMA_URL` | Ollama base URL |
| `OLLAMA_URL` | Ollama base URL |
| `OLLAMA_HOST` | Ollama base URL (Ollama's own convention) |
| `MEMEX_MODEL` | embedding model name |

Flags beat the environment, and among the three URL variables the first one
set wins, in the order listed. A bare `host:port` (how `OLLAMA_HOST` is
usually written) gets an `http://` scheme filled in automatically.

Ollama needs the model pulled first: `ollama pull nomic-embed-text`.

## How it's structured

```
main.go                  CLI entry, subcommand dispatch, flag parsing
internal/vault/          walks the vault, finds .md files, content hashing
internal/chunker/        splits notes into heading-delimited chunks
internal/embed/          Ollama /api/embed client (batched, truncating)
internal/store/          SQLite persistence + brute-force cosine search
internal/index/          orchestrates walk -> hash-diff -> chunk -> embed -> store
internal/watch/          fsnotify-driven incremental re-indexing
```

Each package is usable on its own — `chunker.Split` and `store.Store` in
particular have no CLI dependency, so you can drive them from a test, a
different frontend, or a long-running daemon instead of the one-shot CLI.

### Where the index lives

`.obsidian/memex.db` inside the vault if `.obsidian/` exists, otherwise
`.memex.db` at the vault root. Add it to `.obsidian/.gitignore` (or your
vault's `.gitignore`) if the vault is version-controlled — it's a derived
artifact, not source of truth.

### Why brute-force cosine instead of a vector index extension

`sqlite-vec`/`sqlite-vss` give you a real ANN index, which matters at
hundreds of thousands of vectors. A single Obsidian vault is realistically
low thousands of chunks — brute-force cosine over that in Go is comfortably
under 50ms, and it avoids the extra native-extension-loading step on top of
the cgo SQLite driver itself. If your vault gets big enough that this stops
being true, that's the point to swap `store.Search` for a real ANN index.

## Wiring into Obsidian

The plugin side is intentionally thin: shell out to this binary with
`--json`, parse stdout, render results in a modal or sidebar pane. Something
like:

```ts
const { stdout } = await execFile("memex", [
  "search",
  "--vault", vaultPath,
  "--json",
  "--limit", "10",
  query, // must come last — flags are parsed up to the first positional
]);
const results = JSON.parse(stdout);
```

Run `memex watch --vault <path>` as a background process (or via a
systemd user unit / launchd agent) so the index stays current without the
plugin needing to trigger indexing itself.

## Known rough edges / next steps

- `search` embeds the query on every call — fine for interactive use,
  wasteful if you're scripting many queries in a loop (batch the queries
  through one process instead).
- No embedding model dimension check — if you switch models with different
  vector sizes without `--force`, `cosineSimilarity` will silently score
  mismatched-length vectors as 0 rather than erroring loudly. Worth adding
  a dimension guard if you expect to swap models often.
- `watch` re-runs a full `index.Run` pass on every debounced batch, which
  re-walks the whole vault even though only a few files changed. Fine at
  vault scale, but if this becomes a bottleneck, having fsnotify pass the
  specific changed paths through to a more targeted index function would
  cut redundant hashing.
