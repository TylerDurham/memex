// Package chunker splits markdown content into embedding-sized chunks.
//
// Strategy: split primarily on H1-H3 headings, since that's how most
// Obsidian notes are structured. Within a heading section, if the text is
// still too large, fall back to paragraph-boundary splitting so no single
// chunk blows past the target token budget. Headless notes (journal
// entries, quick captures) are chunked by paragraph directly. A paragraph
// that is itself over budget gets force-split, so the budget is a hard cap
// rather than a hint — embedding models reject oversized inputs outright.
package chunker

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Chunk is one embeddable unit of text extracted from a note.
type Chunk struct {
	// Heading is the nearest enclosing heading text, or "" if none.
	Heading string
	// HeadingPath is the breadcrumb of enclosing headings, e.g.
	// "Q3 Planning > Runway > Assumptions", used for result display.
	HeadingPath string
	// Content is the chunk's raw text (heading line excluded).
	Content string
	// StartLine is the 1-indexed line the chunk starts on, for jump-to-line.
	StartLine int
}

var headingRe = regexp.MustCompile(`^(#{1,3})\s+(.*)$`)

// Options controls chunk sizing, in estimated tokens. See estimateTokens for
// how the estimate is derived.
type Options struct {
	// MaxTokens is the cap per chunk. 450 leaves comfortable headroom under
	// the 2048-token context of typical embedding models (nomic-embed-text,
	// mxbai-embed-large), which reject anything larger rather than clipping.
	MaxTokens int
	// MinTokens avoids emitting near-empty chunks (e.g. a heading with only
	// a one-line stub underneath); such chunks get merged into the previous
	// one, budget permitting.
	MinTokens int
}

func DefaultOptions() Options {
	return Options{MaxTokens: 450, MinTokens: 20}
}

// estimateTokens approximates the token count of s without pulling in a real
// tokenizer.
//
// Word count alone is a trap for web-clipped notes: a line like
//
//	![](assets/MINISFORUM%20N5%20AI%20NAS-r87CJOdNhK.png)
//
// is one "word" but roughly 35 tokens, because percent escapes and random IDs
// tokenize character by character. Taking the larger of the word-based and
// character-based estimates keeps prose governed by the (more accurate) word
// term while still catching URL-dense text before it overruns the model.
func estimateTokens(s string) int {
	byWords := len(strings.Fields(s)) * 13 / 10
	byChars := utf8.RuneCountInString(s) * 10 / 35
	if byChars > byWords {
		return byChars
	}
	return byWords
}

// Split breaks raw markdown content into a slice of Chunks.
func Split(content string, opts Options) []Chunk {
	lines := strings.Split(content, "\n")

	type section struct {
		headingPath []string
		body        []string
		startLine   int
	}

	var sections []section
	cur := section{startLine: 1}
	var stack []string // heading text at each level, index 0 = H1

	flush := func() {
		if len(cur.body) > 0 {
			sections = append(sections, cur)
		}
	}

	for i, line := range lines {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			flush()
			level := len(m[1])
			text := strings.TrimSpace(m[2])

			if level-1 < len(stack) {
				stack = stack[:level-1]
			}
			for len(stack) < level-1 {
				stack = append(stack, "")
			}
			stack = append(stack, text)

			cur = section{
				headingPath: append([]string(nil), stack...),
				startLine:   i + 2, // body starts after the heading line
			}
			continue
		}
		cur.body = append(cur.body, line)
	}
	flush()

	var chunks []Chunk
	for _, s := range sections {
		body := strings.TrimSpace(strings.Join(s.body, "\n"))
		if body == "" {
			continue
		}
		heading := ""
		if len(s.headingPath) > 0 {
			heading = s.headingPath[len(s.headingPath)-1]
		}
		headingPath := strings.Join(s.headingPath, " > ")

		for _, piece := range splitByParagraph(body, opts) {
			chunks = append(chunks, Chunk{
				Heading:     heading,
				HeadingPath: headingPath,
				Content:     piece,
				StartLine:   s.startLine,
			})
		}
	}

	return mergeShortChunks(chunks, opts)
}

var (
	// Markdown image embeds: ![alt](path "title"). Matched non-greedily so a
	// line holding several of them doesn't collapse into one match.
	mdImageRe = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	// Obsidian wikilink embeds: ![[Some Image.png]].
	wikiEmbedRe = regexp.MustCompile(`!\[\[[^\]]*\]\]`)
	// Three or more newlines left behind after stripping.
	blankRunRe = regexp.MustCompile(`\n{3,}`)
	// Trailing spaces on a line, which markdown uses as a hard break and
	// stripping tends to strand.
	trailingWSRe = regexp.MustCompile(`[ \t]+\n`)
)

// StripForEmbedding removes markup that carries no semantic signal before the
// text is handed to the embedding model.
//
// Image links are the main offender: a web-clipped article can be mostly
// `![](assets/Some%20Note-r87CJOdNhK.png)` lines, which contribute nothing to
// what the note means but consume an enormous share of the token budget (and
// dilute the resulting vector). The caller keeps the original text for display
// — only the embedding input is stripped.
func StripForEmbedding(s string) string {
	s = wikiEmbedRe.ReplaceAllString(s, "")
	s = mdImageRe.ReplaceAllString(s, "")
	s = trailingWSRe.ReplaceAllString(s, "\n")
	s = blankRunRe.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

var paraSplitRe = regexp.MustCompile(`\n\s*\n`)

// splitByParagraph breaks body text on blank-line paragraph boundaries,
// grouping consecutive paragraphs until MaxTokens is reached. Paragraphs that
// exceed the budget on their own are force-split first, so every piece
// returned fits.
func splitByParagraph(body string, opts Options) []string {
	paras := paraSplitRe.Split(body, -1)

	var pieces []string
	var buf strings.Builder
	tokens := 0

	flush := func() {
		if buf.Len() > 0 {
			pieces = append(pieces, strings.TrimSpace(buf.String()))
			buf.Reset()
			tokens = 0
		}
	}

	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		for _, part := range splitOversized(p, opts.MaxTokens) {
			pt := estimateTokens(part)

			if tokens > 0 && tokens+pt > opts.MaxTokens {
				flush()
			}
			if buf.Len() > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString(part)
			tokens += pt
		}
	}
	flush()

	if len(pieces) == 0 {
		return []string{body}
	}
	return pieces
}

// splitOversized breaks a single paragraph that is over budget into
// under-budget parts, preferring line boundaries (runs of image links and
// table rows are line-separated) and falling back to a rune-boundary cut when
// one line is still too long on its own.
func splitOversized(p string, maxTokens int) []string {
	if estimateTokens(p) <= maxTokens {
		return []string{p}
	}

	var parts []string
	var buf strings.Builder
	tokens := 0

	flush := func() {
		if buf.Len() > 0 {
			parts = append(parts, strings.TrimSpace(buf.String()))
			buf.Reset()
			tokens = 0
		}
	}

	for _, line := range strings.Split(p, "\n") {
		for _, seg := range splitLongLine(line, maxTokens) {
			st := estimateTokens(seg)
			if tokens > 0 && tokens+st > maxTokens {
				flush()
			}
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(seg)
			tokens += st
		}
	}
	flush()

	if len(parts) == 0 {
		return []string{p}
	}
	return parts
}

// splitLongLine cuts a single line on rune boundaries when even one line
// exceeds the budget (minified HTML, a base64 data URI, a giant table row).
// This is the last resort that makes the budget a genuine hard cap.
func splitLongLine(line string, maxTokens int) []string {
	if estimateTokens(line) <= maxTokens {
		return []string{line}
	}

	// Inverse of the character term in estimateTokens.
	maxRunes := maxTokens * 35 / 10
	if maxRunes < 1 {
		maxRunes = 1
	}

	var segs []string
	runes := []rune(line)
	for len(runes) > maxRunes {
		segs = append(segs, string(runes[:maxRunes]))
		runes = runes[maxRunes:]
	}
	if len(runes) > 0 {
		segs = append(segs, string(runes))
	}
	return segs
}

// mergeShortChunks folds any chunk under MinTokens into its neighbour so we
// don't waste an embedding call (and index row) on a two-line stub. A merge
// that would push the neighbour over MaxTokens is skipped — otherwise this
// step would undo the cap that splitByParagraph just enforced.
func mergeShortChunks(chunks []Chunk, opts Options) []Chunk {
	if len(chunks) <= 1 {
		return chunks
	}

	var out []Chunk
	for _, c := range chunks {
		if estimateTokens(c.Content) < opts.MinTokens && len(out) > 0 {
			prev := &out[len(out)-1]
			merged := prev.Content + "\n\n" + c.Content
			if estimateTokens(merged) <= opts.MaxTokens {
				prev.Content = merged
				continue
			}
		}
		out = append(out, c)
	}
	return out
}
