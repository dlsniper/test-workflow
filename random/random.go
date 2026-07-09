// Package random provides random number generation.
package random

import "math/rand/v2"

// Number returns a uniformly random integer in [0, 1000000].
func Number() int {
	return rand.IntN(1_000_001)
}
