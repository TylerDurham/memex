package documents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Properties_String(t *testing.T) {
	p := Properties{
		"title":    "  Go fmt Verbs ",
		"version":  2,
		"ratio":    1.5,
		"draft":    true,
		"date":     time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC),
		"datetime": time.Date(2026, 8, 20, 18, 20, 16, 0, time.UTC),
		"tags":     []any{"go"},
		"nested":   map[string]any{"a": 1},
		"empty":    nil,
	}

	tests := map[string]string{
		"title":    "Go fmt Verbs",
		"version":  "2",
		"ratio":    "1.5",
		"draft":    "true",
		"date":     "2026-08-20",
		"datetime": "2026-08-20T18:20:16Z",
		"tags":     "",
		"nested":   "",
		"empty":    "",
		"missing":  "",
	}
	for key, want := range tests {
		assert.Equal(t, want, p.String(key), key)
	}

	var none Properties
	assert.Equal(t, "", none.String("title"), "nil Properties")
}
