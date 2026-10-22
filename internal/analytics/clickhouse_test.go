package analytics

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClickHouseStoreInsert(t *testing.T) {
	var gotQuery string
	var gotRows []map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("query")

		scanner := bufio.NewScanner(r.Body)
		for scanner.Scan() {
			var row map[string]string
			if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
				t.Errorf("row is not JSON: %v", err)
				continue
			}
			gotRows = append(gotRows, row)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store, err := NewClickHouseStore(srv.URL+"/", "nexuslink", "", "")
	if err != nil {
		t.Fatalf("NewClickHouseStore() error = %v", err)
	}

	clicked := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	err = store.Insert(context.Background(), []Event{
		{ShortCode: "abc", Timestamp: clicked, IP: "203.0.113.7", UserAgent: "ua", Referrer: "ref", RequestID: "r1"},
		{ShortCode: "def", Timestamp: clicked.Add(time.Second), RequestID: "r2"},
	})
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	if !strings.Contains(gotQuery, "INSERT INTO nexuslink.clicks") {
		t.Errorf("query = %q, want insert into nexuslink.clicks", gotQuery)
	}
	if !strings.Contains(gotQuery, "FORMAT JSONEachRow") {
		t.Errorf("query = %q, want JSONEachRow", gotQuery)
	}
	if len(gotRows) != 2 {
		t.Fatalf("inserted %d rows, want 2", len(gotRows))
	}
	if gotRows[0]["short_code"] != "abc" || gotRows[0]["request_id"] != "r1" {
		t.Errorf("row = %+v", gotRows[0])
	}
	if gotRows[0]["clicked_at"] != "2026-10-21 12:00:00.000" {
		t.Errorf("clicked_at = %q, want ClickHouse DateTime64 layout", gotRows[0]["clicked_at"])
	}
}

func TestClickHouseStoreInsertError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Table nexuslink.clicks does not exist", http.StatusInternalServerError)
	}))
	defer srv.Close()

	store, err := NewClickHouseStore(srv.URL, "nexuslink", "", "")
	if err != nil {
		t.Fatalf("NewClickHouseStore() error = %v", err)
	}

	err = store.Insert(context.Background(), []Event{{ShortCode: "abc", RequestID: "r1"}})
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v, want status 500", err)
	}
}

func TestClickHouseStorePing(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Ok.\n"))
	}))
	defer srv.Close()

	store, err := NewClickHouseStore(srv.URL, "nexuslink", "", "")
	if err != nil {
		t.Fatalf("NewClickHouseStore() error = %v", err)
	}

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if path != "/ping" {
		t.Errorf("ping path = %q, want /ping", path)
	}
}

func TestNewClickHouseStoreRejectsInvalidDatabase(t *testing.T) {
	if _, err := NewClickHouseStore("http://localhost:8123", "bad-name; DROP", "", ""); err == nil {
		t.Fatal("expected error for invalid database identifier")
	}
}
