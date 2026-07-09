package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"named", http.MethodGet, "/hello/Pavel", http.StatusOK, "Hello, Pavel!"},
		{"trailing slash", http.MethodGet, "/hello/", http.StatusOK, "Hello, World!"},
		{"no slash", http.MethodGet, "/hello", http.StatusOK, "Hello, World!"},
		{"non-GET", http.MethodPost, "/hello/x", http.StatusMethodNotAllowed, ""},
		{"random non-GET", http.MethodPost, "/random/number", http.StatusMethodNotAllowed, ""},
		{"unknown", http.MethodGet, "/nope", http.StatusNotFound, ""},
	}

	h := Handler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" {
				body, _ := io.ReadAll(rec.Result().Body)
				if string(body) != tt.wantBody {
					t.Errorf("body = %q, want %q", body, tt.wantBody)
				}
			}
		})
	}
}

func TestRandomNumber(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/random/number", nil)
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
	}
	body, _ := io.ReadAll(rec.Result().Body)
	n, err := strconv.Atoi(string(body))
	if err != nil {
		t.Fatalf("body = %q, not an integer: %v", body, err)
	}
	if n < 0 || n > 1_000_000 {
		t.Errorf("number = %d, out of range [0, 1000000]", n)
	}
}
