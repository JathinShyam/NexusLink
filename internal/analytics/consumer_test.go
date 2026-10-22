package analytics

import (
	"testing"

	"github.com/JathinShyam/NexusLink/internal/logging"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestDecodeClickRecord(t *testing.T) {
	record := &kgo.Record{
		Topic: "url_clicks",
		Value: []byte(`{"short_code":"abc","timestamp":"2026-10-21T12:00:00Z","request_id":"r1"}`),
	}

	event, ok := decodeClickRecord(record, logging.New("error", "test"))
	if !ok {
		t.Fatal("decodeClickRecord() = false, want true")
	}
	if event.ShortCode != "abc" || event.RequestID != "r1" {
		t.Errorf("event = %+v", event)
	}
}

func TestDecodeClickRecordMalformed(t *testing.T) {
	record := &kgo.Record{Topic: "url_clicks", Value: []byte(`{"short_code":`)}
	if _, ok := decodeClickRecord(record, logging.New("error", "test")); ok {
		t.Fatal("decodeClickRecord() = true for malformed JSON, want false")
	}
}
