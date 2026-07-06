package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
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
