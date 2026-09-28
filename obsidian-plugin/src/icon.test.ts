import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { LOGO_INK, toObsidianIcon } from "./icon.ts";

const logoPath = new URL("../../media/squares/memex-m-transparent-dark-ink.svg", import.meta.url);

test("converts the plugin's logo into an Obsidian icon", () => {
	const icon = toObsidianIcon(readFileSync(logoPath, "utf8"));

	assert.ok(icon.startsWith(`<g transform="scale(0.78125)">`), "128px viewBox scaled to 100");
	assert.ok(icon.endsWith("</g>"));
	assert.doesNotMatch(icon, /<svg|<\/svg>/, "outer <svg> removed");
	assert.ok(!icon.includes(LOGO_INK), "ink color replaced");
	assert.match(icon, /stroke="currentColor"/);
	assert.match(icon, /fill="currentColor"/);
	assert.match(icon, /#7c5cff/, "violet accent kept");
});

test("rejects SVGs it can't convert", () => {
	assert.throws(() => toObsidianIcon(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), /viewBox/);
	assert.throws(() => toObsidianIcon(`<svg viewBox="0 0 128 64"></svg>`), /square/);
});
