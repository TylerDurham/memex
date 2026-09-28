package obsidian

import (
	"testing"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FormatObsidianURL(t *testing.T) {
	tests := map[string]struct{ vault, path, want string }{
		"doc example": {"Tech-Kasten", "development/go/Go `fmt` Formatting Verbs",
			"obsidian://open?vault=Tech-Kasten&file=development%2Fgo%2FGo%20%60fmt%60%20Formatting%20Verbs"},
		"query metacharacters": {"My Vault", "Q&A = 1+1?.md",
			"obsidian://open?vault=My%20Vault&file=Q%26A%20%3D%201%2B1%3F.md"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, FormatObsidianURL(tc.vault, tc.path))
		})
	}
}

func Test_LoadFileMetadata_URI(t *testing.T) {
	doc := document.IndexDocument{
		RepoPath: "/home/me/vaults/Tech-Kasten",
		DocPath:  "/home/me/vaults/Tech-Kasten/notes/Go fmt.md",
		RelPath:  "notes/Go fmt.md",
	}
	p := NewObsidianIndexDocumentProvider()
	require.NoError(t, p.LoadFileMetadata(&doc, nil))

	want := "obsidian://open?vault=Tech-Kasten&file=notes%2FGo%20fmt.md"
	assert.Equal(t, want, doc.URI)
	assert.Equal(t, "obsidian", doc.Application)

	launch, err := p.LaunchURL(&doc)
	require.NoError(t, err)
	assert.Equal(t, want, launch)
}
