package greeting

import (
	"fmt"
	"strings"
)

// Hello returns a greeting.
func Hello(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// Goodbye returns a farewell.
func Goodbye(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Good bye, %s!", name)
}
