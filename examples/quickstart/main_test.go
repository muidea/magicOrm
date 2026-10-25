package main

import (
	"os"
	"strings"
	"testing"
)

// The published complete snippet is the same compilation unit built by CI.
func TestPublishedQuickstartMatchesCompilableSource(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	begin := "<!-- quickstart-source -->\n```go\n"
	_, rest, ok := strings.Cut(string(readme), begin)
	if !ok {
		t.Fatal("missing quickstart marker")
	}
	snippet, _, ok := strings.Cut(rest, "```\n<!-- /quickstart-source -->")
	if !ok || snippet != string(source) {
		t.Fatal("README quickstart differs from compilable source")
	}
}
