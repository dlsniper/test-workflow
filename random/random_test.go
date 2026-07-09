package random

import "testing"

func TestNumber(t *testing.T) {
	for range 2500 {
		n := Number()
		if n < 0 || n > 1_000_000 {
			t.Fatalf("Number() = %d, out of range [0, 1000000]", n)
		}
	}
}
