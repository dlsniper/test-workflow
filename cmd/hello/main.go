// Command hello serves the hello HTTP API.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/dlsniper/test-workflow/httpapi"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, httpapi.Handler()); err != nil {
		log.Fatal(err)
	}
}
