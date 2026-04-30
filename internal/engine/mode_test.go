package engine

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Mode
	}{
		{name: "default", raw: "", want: ModeHybrid},
		{name: "hybrid", raw: "hybrid", want: ModeHybrid},
		{name: "browser", raw: "browser", want: ModeBrowser},
		{name: "http", raw: "http", want: ModeHTTP},
		{name: "trim and case", raw: " Browser ", want: ModeBrowser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.raw)
			if err != nil {
				t.Fatalf("ParseMode(%q) returned error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("ParseMode(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseModeRejectsUnsupportedMode(t *testing.T) {
	if _, err := ParseMode("magic"); err == nil {
		t.Fatal("ParseMode(magic) returned nil error, want failure")
	}
}
