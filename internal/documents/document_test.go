package documents

import (
	"encoding/json"
	"testing"
)

func TestToJSONString(t *testing.T) {
	doc := Document{
		Abs:       "/vault/note.md",
		Chunks:     []Chunk{{Text: "hello", HeadingPath: []string{"Intro"}, StartLine: 1, EndLine: 3}},
		Properties: Properties{"tags": []string{"go"}},
	}

	s, err := doc.ToJSONString()
	if err != nil {
		t.Fatalf("ToJSONString: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(s), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, s)
	}
	if got["path"] != "/vault/note.md" {
		t.Errorf("path = %v, want /vault/note.md", got["path"])
	}

	chunks, ok := got["chunks"].([]any)
	if !ok || len(chunks) != 1 {
		t.Fatalf("chunks = %v, want one chunk", got["chunks"])
	}
	chunk := chunks[0].(map[string]any)
	for _, key := range []string{"text", "headingPath", "startLine", "endLine"} {
		if _, ok := chunk[key]; !ok {
			t.Errorf("chunk is missing key %q: %v", key, chunk)
		}
	}
}

func TestToJSONStringUnsupportedProperty(t *testing.T) {
	// JSON can't encode a channel, so marshaling must fail with an error.
	doc := Document{Properties: Properties{"bad": make(chan int)}}
	if _, err := doc.ToJSONString(); err == nil {
		t.Error("ToJSONString succeeded, want error")
	}
}
