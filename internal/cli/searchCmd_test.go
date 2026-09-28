package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/TylerDurham/memex/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testResults() []store.Result {
	return []store.Result{
		{
			Chunk: store.Chunk{
				FilePath:    "gear/dock-specs.md",
				Heading:     "Notable Features",
				HeadingPath: "Dock > Notable Features",
				Content:     "## Notable Features\n\n- Q&A: <120Gbps>",
				StartLine:   52,
			},
			Title:       "ThinkPad Dock",
			Description: "A Thunderbolt 5 dock.",
			Application: "obsidian",
			URI:         "obsidian://open?vault=v&file=gear%2Fdock-specs.md",
			Score:       0.792,
		},
		{
			// No frontmatter title, no heading (preamble chunk).
			Chunk: store.Chunk{FilePath: "notes/Go fmt.md", Content: "Intro.", StartLine: 1},
			Score: 0.5,
		},
	}
}

func Test_writeResultsText(t *testing.T) {
	var out bytes.Buffer
	writeResultsText(&out, testResults())

	assert.Equal(t, ""+
		"0.792  ThinkPad Dock\n"+
		"       gear/dock-specs.md:52 · Notable Features\n"+
		"0.500  Go fmt\n"+
		"       notes/Go fmt.md:1\n",
		out.String())
}

func Test_writeResultsJSON(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, writeResultsJSON(&out, testResults()))

	var got []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got), out.String())
	require.Len(t, got, 2)

	assert.Equal(t, map[string]any{
		"score":        0.792,
		"title":        "ThinkPad Dock",
		"description":  "A Thunderbolt 5 dock.",
		"path":         "gear/dock-specs.md",
		"line":         52.0,
		"heading":      "Notable Features",
		"heading_path": "Dock > Notable Features",
		"application":  "obsidian",
		"uri":          "obsidian://open?vault=v&file=gear%2Fdock-specs.md",
		"content":      "## Notable Features\n\n- Q&A: <120Gbps>",
	}, got[0])

	// Empty optional fields are omitted; title falls back to the file name.
	assert.Equal(t, map[string]any{
		"score":   0.5,
		"title":   "Go fmt",
		"path":    "notes/Go fmt.md",
		"line":    1.0,
		"content": "Intro.",
	}, got[1])

	assert.Contains(t, out.String(), "Q&A: <120Gbps>", "HTML characters not escaped")
}

func Test_writeResultsJSON_Empty(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, writeResultsJSON(&out, nil))
	assert.Equal(t, "[]\n", out.String())
}
