package markdown

import (
	"strings"

	"github.com/TylerDurham/memex/internal/documents"
	"github.com/yuin/goldmark/v2/ast"
)

// section marks where a top-level heading starts in the body.
type section struct {
	line  int // 0-based line index into the body
	level int
	title string
}

// DefaultMaxChunkChars caps chunk size for Chunks. At roughly 4 characters
// per token it keeps chunks near 500 tokens, which fits common embedding
// models.
const DefaultMaxChunkChars = 2000

// Chunks splits the body into chunks of at most DefaultMaxChunkChars. See
// ChunksMax.
func (d *Document) Chunks() []documents.Chunk {
	return d.ChunksMax(DefaultMaxChunkChars)
}

// ChunksMax splits the body into one chunk per heading section, for semantic
// indexing. Any text before the first heading becomes its own chunk with an
// empty HeadingPath. Only top-level headings split sections; headings nested
// in lists or blockquotes stay inside their parent chunk. Line numbers are
// 1-based and refer to the original file (frontmatter included).
//
// Sections longer than maxChars are split at paragraph boundaries (blank
// lines outside fenced code), packing as many paragraphs per chunk as fit. A
// single paragraph longer than maxChars is split between lines; lines are
// never split. maxChars <= 0 disables splitting.
func (d *Document) ChunksMax(maxChars int) []documents.Chunk {
	lines := strings.Split(strings.TrimSuffix(string(d.Source), "\n"), "\n")
	lineStarts := make([]int, len(lines))
	for i, off := 1, 0; i < len(lines); i++ {
		off += len(lines[i-1]) + 1
		lineStarts[i] = off
	}

	sections := []section{{line: 0}} // preamble
	for n := d.Body.FirstChild(); n != nil; n = n.NextSibling() {
		h, ok := n.(*ast.Heading)
		if !ok {
			continue
		}
		src := h.Source()
		if len(src) == 0 {
			continue // empty heading such as a bare "#"; nothing to anchor on
		}
		var title strings.Builder
		for _, seg := range src {
			title.Write(seg.Bytes(d.Source))
		}
		sections = append(sections, section{
			line:  lineOf(lineStarts, src[0].Start),
			level: h.Level,
			title: strings.TrimSpace(title.String()),
		})
	}

	var chunks []documents.Chunk
	var path []section
	for i, s := range sections {
		end := len(lines)
		if i+1 < len(sections) {
			end = sections[i+1].line
		}

		if i > 0 {
			for len(path) > 0 && path[len(path)-1].level >= s.level {
				path = path[:len(path)-1]
			}
			path = append(path, s)
		}

		// Trim blank lines at both edges so StartLine/EndLine are exact.
		start := s.line
		for start < end && strings.TrimSpace(lines[start]) == "" {
			start++
		}
		for end > start && strings.TrimSpace(lines[end-1]) == "" {
			end--
		}
		if start == end {
			continue
		}

		headings := make([]string, len(path))
		for j, p := range path {
			headings[j] = p.title
		}

		for _, r := range splitSection(lines, start, end, maxChars) {
			chunks = append(chunks, documents.Chunk{
				Text:        strings.Join(lines[r.start:r.end], "\n"),
				HeadingPath: headings,
				StartLine:   d.BodyLine + r.start,
				EndLine:     d.BodyLine + r.end - 1,
			})
		}
	}

	return chunks
}

// lineRange is a half-open range of 0-based line indexes.
type lineRange struct{ start, end int }

// textLen is the length of lines[r.start:r.end] joined with newlines.
func textLen(lines []string, r lineRange) int {
	n := r.end - r.start - 1
	for _, l := range lines[r.start:r.end] {
		n += len(l)
	}
	return n
}

// splitSection splits the non-blank-edged range [start, end) into ranges of at
// most maxChars, preferring paragraph boundaries.
func splitSection(lines []string, start, end, maxChars int) []lineRange {
	whole := lineRange{start, end}
	if maxChars <= 0 || textLen(lines, whole) <= maxChars {
		return []lineRange{whole}
	}

	// Break oversized paragraphs into line groups so every piece fits, except
	// a single line that is itself too long.
	var pieces []lineRange
	for _, p := range paragraphs(lines, start, end) {
		if textLen(lines, p) <= maxChars {
			pieces = append(pieces, p)
			continue
		}
		cur := lineRange{p.start, p.start + 1}
		for i := p.start + 1; i < p.end; i++ {
			if textLen(lines, lineRange{cur.start, i + 1}) > maxChars {
				pieces = append(pieces, cur)
				cur = lineRange{i, i + 1}
			} else {
				cur.end = i + 1
			}
		}
		pieces = append(pieces, cur)
	}

	// Greedily pack consecutive pieces. A packed range spans the blank lines
	// between its pieces, so measure the whole span.
	var out []lineRange
	cur := pieces[0]
	for _, p := range pieces[1:] {
		if textLen(lines, lineRange{cur.start, p.end}) > maxChars {
			out = append(out, cur)
			cur = p
		} else {
			cur.end = p.end
		}
	}
	return append(out, cur)
}

// paragraphs returns the blank-line-separated blocks in [start, end). Blank
// lines inside fenced code blocks don't separate paragraphs.
func paragraphs(lines []string, start, end int) []lineRange {
	var out []lineRange
	var fence string
	pStart := -1
	for i := start; i < end; i++ {
		trimmed := strings.TrimSpace(lines[i])

		if fence == "" {
			if marker := fenceMarker(trimmed); marker != "" {
				fence = marker
			}
		} else if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
			fence = ""
		}

		if trimmed == "" && fence == "" {
			if pStart >= 0 {
				out = append(out, lineRange{pStart, i})
				pStart = -1
			}
		} else if pStart < 0 {
			pStart = i
		}
	}
	if pStart >= 0 {
		out = append(out, lineRange{pStart, end})
	}
	return out
}

// fenceMarker returns the opening code fence (e.g. "```" or "~~~~") that line
// starts with, or "" if it doesn't open a fence.
func fenceMarker(line string) string {
	if line == "" || (line[0] != '`' && line[0] != '~') {
		return ""
	}
	n := len(line) - len(strings.TrimLeft(line, line[:1]))
	if n < 3 {
		return ""
	}
	return line[:n]
}

// lineOf returns the 0-based index of the line containing byte offset off.
func lineOf(lineStarts []int, off int) int {
	lo, hi := 0, len(lineStarts)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if lineStarts[mid] <= off {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}
