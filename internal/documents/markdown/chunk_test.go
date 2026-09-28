package markdown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Chunks(t *testing.T) {
	const input = `---
title: foo
---

Intro text.

# Top

Top body.

## Child

` + "```" + `
# not a heading
` + "```" + `

- # heading inside a list item

## Sibling
Sibling body.

# Next Top
`

	doc, err := Parse(strings.NewReader(input))
	require.NoError(t, err)

	chunks := doc.Chunks()
	require.Len(t, chunks, 5)

	assert.Equal(t, "Intro text.", chunks[0].Text)
	assert.Empty(t, chunks[0].HeadingPath)
	assert.Equal(t, 5, chunks[0].StartLine)
	assert.Equal(t, 5, chunks[0].EndLine)

	assert.Equal(t, "# Top\n\nTop body.", chunks[1].Text)
	assert.Equal(t, []string{"Top"}, chunks[1].HeadingPath)
	assert.Equal(t, 7, chunks[1].StartLine)
	assert.Equal(t, 9, chunks[1].EndLine)

	assert.Equal(t, []string{"Top", "Child"}, chunks[2].HeadingPath)
	assert.Contains(t, chunks[2].Text, "# not a heading")
	assert.Contains(t, chunks[2].Text, "heading inside a list item")
	assert.Equal(t, 11, chunks[2].StartLine)
	assert.Equal(t, 17, chunks[2].EndLine)

	assert.Equal(t, []string{"Top", "Sibling"}, chunks[3].HeadingPath)
	assert.Equal(t, []string{"Next Top"}, chunks[4].HeadingPath)
	assert.Equal(t, 22, chunks[4].StartLine)
}

func Test_ReadFrontmatter(t *testing.T) {
	fm, err := ReadFrontmatter(newTestScanner("---\ntitle: foo\ntags:\n  - a\n---\n# Body\n"))
	require.NoError(t, err)
	assert.Equal(t, "foo", fm["title"])
	assert.Equal(t, []any{"a"}, fm["tags"])

	fm, err = ReadFrontmatter(newTestScanner("# No frontmatter\n"))
	require.NoError(t, err)
	assert.Nil(t, fm)

	_, err = ReadFrontmatter(newTestScanner("---\ntitle: foo\n"))
	assert.Error(t, err)
}

func Test_ChunksMax_SplitsLongSections(t *testing.T) {
	para := func(word string) string { return strings.TrimSpace(strings.Repeat(word+" ", 10)) }
	input := "# Long\n\n" +
		para("alpha") + "\n\n" +
		para("bravo") + "\n\n" +
		"```\ncode\n\nstill code\n```\n\n" +
		para("delta") + "\n"

	doc, err := Parse(strings.NewReader(input))
	require.NoError(t, err)

	// Unlimited: one chunk for the whole section.
	require.Len(t, doc.ChunksMax(0), 1)

	chunks := doc.ChunksMax(80)
	require.Len(t, chunks, 4)
	for _, c := range chunks {
		assert.Equal(t, []string{"Long"}, c.HeadingPath)
		assert.LessOrEqual(t, len(c.Text), 80, c.Text)
	}

	// The heading packs with the first paragraph; the second doesn't fit.
	assert.Equal(t, "# Long\n\n"+para("alpha"), chunks[0].Text)
	assert.Equal(t, 1, chunks[0].StartLine)
	assert.Equal(t, 3, chunks[0].EndLine)
	assert.Equal(t, para("bravo"), chunks[1].Text)
	assert.Equal(t, 5, chunks[1].StartLine)
	// The blank line inside the fence doesn't split the code block.
	assert.Equal(t, "```\ncode\n\nstill code\n```", chunks[2].Text)
	assert.Equal(t, 7, chunks[2].StartLine)
	assert.Equal(t, 11, chunks[2].EndLine)
}

func Test_ChunksMax_PacksParagraphs(t *testing.T) {
	input := "# H\n\none\n\ntwo\n\nthree\n"
	doc, err := Parse(strings.NewReader(input))
	require.NoError(t, err)

	chunks := doc.ChunksMax(12)
	require.Len(t, chunks, 2)
	assert.Equal(t, "# H\n\none", chunks[0].Text)
	assert.Equal(t, "two\n\nthree", chunks[1].Text)
}

func Test_ChunksMax_SplitsLongParagraphByLine(t *testing.T) {
	input := "aaaa\nbbbb\ncccc\ndddd\n"
	doc, err := Parse(strings.NewReader(input))
	require.NoError(t, err)

	chunks := doc.ChunksMax(9)
	require.Len(t, chunks, 2)
	assert.Equal(t, "aaaa\nbbbb", chunks[0].Text)
	assert.Equal(t, "cccc\ndddd", chunks[1].Text)
	assert.Equal(t, 3, chunks[1].StartLine)
}
