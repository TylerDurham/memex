// Runs the memex CLI and parses its `search --json` output. Kept free of
// Obsidian imports so it can be tested with plain Node.

import { execFile } from "node:child_process";

/** One result from `memex search --json`. Mirrors jsonResult in searchCmd.go. */
export interface MemexResult {
	score: number;
	title: string; // frontmatter title, or file name if none
	description?: string;
	path: string; // repo-relative
	line: number; // 1-based line where the section starts
	heading?: string;
	heading_path?: string;
	application?: string;
	uri?: string;
	content: string;
}

export interface SearchOptions {
	binary: string; // memex executable, a name on PATH or a full path
	repo: string;
	query: string;
	top: number;
	minScore: number;
	signal?: AbortSignal; // aborting kills the memex process
	timeoutMs?: number;
}

/** A failure reported by memex, or in running it, with a user-facing message. */
export class MemexError extends Error {
	override name = "MemexError";
}

const DEFAULT_TIMEOUT_MS = 30_000;
const MAX_OUTPUT_BYTES = 16 * 1024 * 1024;

/** Arguments for a search. The query follows "--" so text starting with "-" isn't read as a flag. */
export function searchArgs(opts: SearchOptions): string[] {
	return [
		"search", opts.repo, "--json",
		"-k", String(opts.top),
		"--min-score", String(opts.minScore),
		"--", opts.query,
	];
}

/** Runs a search. Rejects with MemexError on failure, or with an AbortError if opts.signal fires. */
export function search(opts: SearchOptions): Promise<MemexResult[]> {
	return new Promise((resolve, reject) => {
		execFile(
			opts.binary,
			searchArgs(opts),
			{
				signal: opts.signal,
				timeout: opts.timeoutMs ?? DEFAULT_TIMEOUT_MS,
				maxBuffer: MAX_OUTPUT_BYTES,
			},
			(err, stdout, stderr) => {
				if (err) {
					if (err.name === "AbortError") {
						reject(err);
					} else if ((err as NodeJS.ErrnoException).code === "ENOENT") {
						reject(new MemexError(
							`memex not found at "${opts.binary}". Set its full path in the plugin settings.`));
					} else if (err.killed) {
						reject(new MemexError("memex search timed out. Is the Ollama server reachable?"));
					} else {
						// memex prints its error to stderr and exits non-zero.
						reject(new MemexError(stderr.trim() || err.message));
					}
					return;
				}

				try {
					resolve(parseResults(stdout));
				} catch (e) {
					reject(new MemexError(`unexpected output from memex: ${(e as Error).message}`));
				}
			},
		);
	});
}

/** Parses `memex search --json` output. */
export function parseResults(stdout: string): MemexResult[] {
	const value: unknown = JSON.parse(stdout);
	if (!Array.isArray(value)) {
		throw new Error("expected a JSON array");
	}
	return value as MemexResult[];
}

/** A short plain-text preview of a result's content, without its heading line. */
export function snippet(r: MemexResult, maxChars = 160): string {
	const lines = r.content.split("\n");
	if (lines[0]?.startsWith("#")) {
		lines.shift();
	}
	const text = lines
		.map(plainLine)
		.filter((l) => l !== "")
		.join(" ")
		.replace(/\s+/g, " ")
		.trim();
	return text.length > maxChars ? text.slice(0, maxChars - 1).trimEnd() + "…" : text;
}

const TABLE_SEPARATOR = /^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$/;
const LIST_MARKER = /^\s*(?:[-*+]|\d+[.)])\s+/;

/** Reduces one Markdown line to readable text: table rows become "a · b", list markers go. */
function plainLine(line: string): string {
	if (TABLE_SEPARATOR.test(line)) {
		return "";
	}
	const trimmed = line.trim();
	if (trimmed.startsWith("|")) {
		return trimmed.split("|").map((c) => c.trim()).filter((c) => c !== "").join(" · ");
	}
	return line.replace(LIST_MARKER, "");
}
