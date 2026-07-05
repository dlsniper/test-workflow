package greeting

import (
	"fmt"
	"strings"
)

// Hello returns a greeting for name. The name is trimmed of surrounding
// whitespace; an empty or whitespace-only name greets "World".
func Hello(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello, %s!", name)
}
