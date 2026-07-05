package greeting

import (
	"fmt"
	"strings"
)

// Goodbye returns a farewell.
func Goodbye(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Good bye, %s!", name)
}
