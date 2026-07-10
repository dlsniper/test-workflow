// Command hello serves the hello HTTP API.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"

	"github.com/dlsniper/test-workflow/httpapi"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	log.Printf("Hello, World!")
	printRandomLetters(log.Writer())
	fmt.Printf("random number: %d\n", rand.IntN(10))

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, httpapi.Handler()); err != nil {
		log.Fatal(err)
	}
}

func printRandomLetters(w io.Writer) {
	const letters = "abcdefghijklmnopqrstuvwxyz"

	value := make([]byte, 5)
	for i := range value {
		value[i] = letters[rand.IntN(len(letters))]
	}

	fmt.Fprintf(w, "random letters: %s\n", value)
}
