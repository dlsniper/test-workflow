package greeting

import "testing"

func TestHello(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "Bob", "Hello, Bob!"},
		{"empty", "", "Hello, World!"},
		{"whitespace only", " \t\n", "Hello, World!"},
		{"padded name", "  Bob  ", "Hello, Bob!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hello(tt.input); got != tt.want {
				t.Errorf("Hello(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
