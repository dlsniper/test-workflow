package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestPrintRandomLetters(t *testing.T) {
	var output bytes.Buffer

	printRandomLetters(&output)

	if matched := regexp.MustCompile(`^random letters: [a-z]{5}\n$`).MatchString(output.String()); !matched {
		t.Fatalf("printRandomLetters() output = %q, want one random-letters line containing five lowercase ASCII letters", output.String())
	}
}
