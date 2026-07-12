package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatContracts(t *testing.T) {
	for _, format := range []string{FormatTable, FormatText, FormatJSON, FormatCSV} {
		if err := ValidateSnapshotFormat(format); err != nil {
			t.Fatalf("snapshot %q: %v", format, err)
		}
	}
	for _, format := range []string{FormatText, FormatJSONL, FormatCSV} {
		if err := ValidateStreamFormat(format); err != nil {
			t.Fatalf("stream %q: %v", format, err)
		}
	}
	if err := ValidateSnapshotFormat(FormatJSONL); err == nil {
		t.Fatal("jsonl accepted for snapshot")
	}
	if err := ValidateStreamFormat(FormatJSON); err == nil {
		t.Fatal("json accepted for stream")
	}
}

func TestCSVStreamWritesHeaderOnce(t *testing.T) {
	var out bytes.Buffer
	stream, err := NewStream(&out, FormatCSV, []string{"name", "value"})
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Write([]string{"a", "1"}, nil); err != nil {
		t.Fatal(err)
	}
	if err := stream.Write([]string{"b", "2"}, nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(out.String(), "name,value"); got != 1 {
		t.Fatalf("header count = %d; output=%q", got, out.String())
	}
}

func TestJSONLStreamIsLineDelimited(t *testing.T) {
	var out bytes.Buffer
	stream, err := NewStream(&out, FormatJSONL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Write(nil, map[string]int{"value": 1}); err != nil {
		t.Fatal(err)
	}
	if err := stream.Write(nil, map[string]int{"value": 2}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(strings.TrimSpace(out.String()), "\n"); got != 1 {
		t.Fatalf("line count separator = %d; output=%q", got, out.String())
	}
}
