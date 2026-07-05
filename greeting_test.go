package greeting

import "testing"

func TestGoodbye(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "Bob", "Good bye, Bob!"},
		{"empty", "", "Good bye, World!"},
		{"whitespace only", " \t\n", "Good bye, World!"},
		{"padded name", "  Bob  ", "Good bye, Bob!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Goodbye(tt.input); got != tt.want {
				t.Errorf("Goodbye(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
