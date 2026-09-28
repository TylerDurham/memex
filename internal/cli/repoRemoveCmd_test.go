package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_confirm(t *testing.T) {
	tests := map[string]bool{
		"y\n":     true,
		"yes\n":   true,
		" YES \n": true,
		"Y":       true, // no trailing newline
		"\n":      false,
		"":        false, // EOF, e.g. stdin closed
		"n\n":     false,
		"no\n":    false,
		"yep\n":   false,
	}
	for input, want := range tests {
		var out bytes.Buffer
		got, err := confirm(strings.NewReader(input), &out, "Remove?")
		require.NoError(t, err, "%q", input)
		assert.Equal(t, want, got, "%q", input)
		assert.True(t, strings.HasPrefix(out.String(), "Remove? [y/N] "), "%q", input)
	}
}
