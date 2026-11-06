package main

import (
	"testing"
)

func TestParseProductVersion(t *testing.T) {
	if got := parseProductVersion(wailsConfigJSON); got == "" || got == "unknown" {
		t.Fatalf("embedded product version = %q", got)
	}
	if got := parseProductVersion([]byte(`{"info":{"productVersion":" 9.8.7.6 "}}`)); got != "9.8.7.6" {
		t.Fatalf("trimmed version = %q", got)
	}
	if got := parseProductVersion([]byte(`{"info":{}}`)); got != "unknown" {
		t.Fatalf("missing version = %q", got)
	}
}
