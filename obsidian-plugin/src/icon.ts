// Turns a logo SVG from media/squares into an Obsidian icon. Kept free of
// Obsidian imports (and of the .svg import itself) so it can be tested with
// plain Node.

/** ID the plugin registers its icon under with addIcon. */
export const ICON_ID = "memex-logo";

/** The logo's ink color, which becomes currentColor so the icon follows the theme. */
export const LOGO_INK = "#1a1d26";

/**
 * Converts a square SVG into addIcon's format: the SVG's inner content, scaled
 * to Obsidian's 0 0 100 100 viewBox, with LOGO_INK replaced by currentColor.
 */
export function toObsidianIcon(svg: string): string {
	const match = svg.match(/<svg\b[^>]*\bviewBox="0 0 ([\d.]+) ([\d.]+)"[^>]*>([\s\S]*)<\/svg>\s*$/);
	if (!match) {
		throw new Error(`logo SVG must be a single <svg> with viewBox="0 0 W H"`);
	}
	const [, width, height, inner] = match as unknown as [string, string, string, string];
	if (width !== height) {
		throw new Error(`logo SVG must be square, got ${width}x${height}`);
	}

	const scale = 100 / Number(width);
	const content = inner.replaceAll(LOGO_INK, "currentColor").trim();
	return `<g transform="scale(${scale})">${content}</g>`;
}
