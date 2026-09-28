import { test } from "node:test";
import assert from "node:assert/strict";
import { chmodSync, mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { MemexError, parseResults, search, searchArgs, snippet, type MemexResult, type SearchOptions } from "./memex.ts";

// fakeMemex writes an executable script that behaves like memex for tests.
function fakeMemex(body: string): string {
	const dir = mkdtempSync(join(tmpdir(), "memex-fake-"));
	const path = join(dir, "memex");
	writeFileSync(path, `#!/usr/bin/env node\n${body}\n`);
	chmodSync(path, 0o755);
	return path;
}

function opts(binary: string, extra: Partial<SearchOptions> = {}): SearchOptions {
	return { binary, repo: "kasten", query: "go printf verbs", top: 5, minScore: 0.5, ...extra };
}

const result: MemexResult = {
	score: 0.678,
	title: "Go fmt Formatting Verbs",
	path: "notes/Go fmt.md",
	line: 56,
	heading: "Structs",
	content: "## Structs\n\nUse %+v to include   field names.\n",
};

test("searchArgs puts the query after -- so leading dashes aren't flags", () => {
	assert.deepEqual(searchArgs(opts("memex", { query: "-k 99 --help" })), [
		"search", "kasten", "--json", "-k", "5", "--min-score", "0.5", "--", "-k 99 --help",
	]);
});

test("search runs memex and parses its JSON", async () => {
	// Echo the args back inside a result so the test can check them.
	const bin = fakeMemex(`
		const r = ${JSON.stringify(result)};
		r.content = JSON.stringify(process.argv.slice(2));
		console.log(JSON.stringify([r]));`);

	const got = await search(opts(bin));
	assert.equal(got.length, 1);
	assert.equal(got[0]?.title, result.title);
	assert.equal(got[0]?.line, 56);
	assert.deepEqual(JSON.parse(got[0]!.content), searchArgs(opts(bin)));
});

test("search returns [] for no matches", async () => {
	assert.deepEqual(await search(opts(fakeMemex(`console.log("[]")`))), []);
});

test("search reports memex's stderr on failure", async () => {
	const bin = fakeMemex(`console.error("repo not found: \\"kasten\\""); process.exit(1);`);
	await assert.rejects(search(opts(bin)), (e: Error) =>
		e instanceof MemexError && e.message === `repo not found: "kasten"`);
});

test("search explains a missing binary", async () => {
	await assert.rejects(search(opts("/nonexistent/memex")), (e: Error) =>
		e instanceof MemexError && e.message.includes("Set its full path"));
});

test("search rejects unexpected output", async () => {
	await assert.rejects(search(opts(fakeMemex(`console.log("not json")`))), (e: Error) =>
		e instanceof MemexError && e.message.startsWith("unexpected output from memex"));
	await assert.rejects(search(opts(fakeMemex(`console.log("{}")`))), /expected a JSON array/);
});

test("search times out", async () => {
	const bin = fakeMemex(`setTimeout(() => {}, 10_000);`);
	await assert.rejects(search(opts(bin, { timeoutMs: 200 })), (e: Error) =>
		e instanceof MemexError && e.message.includes("timed out"));
});

test("aborting kills memex and rejects with AbortError", async () => {
	const bin = fakeMemex(`setTimeout(() => console.log("[]"), 10_000);`);
	const controller = new AbortController();
	const pending = search(opts(bin, { signal: controller.signal }));
	setTimeout(() => controller.abort(), 50);
	const started = Date.now();
	await assert.rejects(pending, { name: "AbortError" });
	assert.ok(Date.now() - started < 5_000, "memex was killed, not waited for");
});

test("parseResults", () => {
	assert.deepEqual(parseResults(JSON.stringify([result])), [result]);
	assert.throws(() => parseResults("null"), /expected a JSON array/);
});

test("snippet drops the heading line, collapses whitespace and truncates", () => {
	assert.equal(snippet(result), "Use %+v to include field names.");
	assert.equal(snippet({ ...result, content: "No heading here." }), "No heading here.");
	const long = snippet({ ...result, content: "x".repeat(500) }, 20);
	assert.equal(long.length, 20);
	assert.ok(long.endsWith("…"));
});

test("snippet makes tables and lists readable", () => {
	const table = "## Tech Specs\n\n| Spec | Detail |\n|---|:---:|\n| Model | A83B5 |\n| Ports | 14 |\n";
	assert.equal(snippet({ ...result, content: table }), "Spec · Detail Model · A83B5 Ports · 14");

	const list = "## Features\n- First item\n* Second\n1. Third\n2) Fourth\n- `%v` - keeps inner dashes";
	assert.equal(snippet({ ...result, content: list }), "First item Second Third Fourth `%v` - keeps inner dashes");
});
