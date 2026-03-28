package output_test

import (
	"bytes"
	"testing"

	"github.com/joelazar/kagi/internal/output"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    output.Format
		wantErr bool
	}{
		{"json", output.FormatJSON, false},
		{"compact", output.FormatCompact, false},
		{"pretty", output.FormatPretty, false},
		{"markdown", output.FormatMarkdown, false},
		{"csv", output.FormatCSV, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := output.ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidFormats(t *testing.T) {
	formats := output.ValidFormats()
	if len(formats) != 5 {
		t.Errorf("expected 5 formats, got %d", len(formats))
	}
}

func TestWriteJSONTo(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}

	if err := output.WriteJSONTo(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "{\n  \"key\": \"value\"\n}\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestWriteCompactTo(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}

	if err := output.WriteCompactTo(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "{\"key\":\"value\"}\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}
