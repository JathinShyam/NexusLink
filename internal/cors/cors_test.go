package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	handler := Middleware([]string{"http://localhost:5173"})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Allow-Origin = %q", got)
	}
}

func TestMiddlewarePreflight(t *testing.T) {
	handler := Middleware([]string{"http://localhost:5173"})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("next should not run on OPTIONS")
		}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/shorten", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestParseOrigins(t *testing.T) {
	got := ParseOrigins(" http://localhost:5173 , http://127.0.0.1:5173 ")
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
}
