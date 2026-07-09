// Package httpapi serves the hello HTTP API.
package httpapi

import (
	"io"
	"net/http"
	"strconv"

	"github.com/dlsniper/test-workflow"
	"github.com/dlsniper/test-workflow/random"
)

// Handler returns the HTTP handler for the hello API.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello/{name}", hello)
	mux.HandleFunc("GET /hello/{$}", hello)
	mux.HandleFunc("GET /hello", hello)
	mux.HandleFunc("GET /random/number", randomNumber)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, greeting.Hello(r.PathValue("name")))
}

func randomNumber(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, strconv.Itoa(random.Number()))
}
