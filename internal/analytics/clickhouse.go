package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const clickHouseTimestampLayout = "2006-01-02 15:04:05.000"

var identifierRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Store persists batches of click events. The consumer only depends on this
// interface, so tests can substitute an in-memory implementation.
type Store interface {
	Insert(ctx context.Context, events []Event) error
	Ping(ctx context.Context) error
}

// ClickHouseStore writes click events through the ClickHouse HTTP interface
// using JSONEachRow, which avoids a native-protocol dependency.
type ClickHouseStore struct {
	baseURL  string
	database string
	table    string
	user     string
	password string
	client   *http.Client
}

// NewClickHouseStore returns a store for the given HTTP endpoint (for example
// http://localhost:8123). Database names are validated because they cannot be
// bound as query parameters.
func NewClickHouseStore(baseURL, database, user, password string) (*ClickHouseStore, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("clickhouse: url is required")
	}
	if database == "" {
		database = "nexuslink"
	}
	if !identifierRE.MatchString(database) {
		return nil, fmt.Errorf("clickhouse: invalid database %q", database)
	}

	return &ClickHouseStore{
		baseURL:  baseURL,
		database: database,
		table:    "clicks",
		user:     user,
		password: password,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type clickRow struct {
	ShortCode string `json:"short_code"`
	ClickedAt string `json:"clicked_at"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Referrer  string `json:"referrer"`
	RequestID string `json:"request_id"`
}

// Insert writes the events as a single JSONEachRow block.
func (s *ClickHouseStore) Insert(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	for _, event := range events {
		row := clickRow{
			ShortCode: event.ShortCode,
			ClickedAt: event.Timestamp.UTC().Format(clickHouseTimestampLayout),
			IP:        event.IP,
			UserAgent: event.UserAgent,
			Referrer:  event.Referrer,
			RequestID: event.RequestID,
		}
		if err := encoder.Encode(row); err != nil {
			return fmt.Errorf("encode click row: %w", err)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s.%s (short_code, clicked_at, ip, user_agent, referrer, request_id) FORMAT JSONEachRow",
		s.database, s.table,
	)
	endpoint := s.baseURL + "/?" + url.Values{"query": []string{query}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return fmt.Errorf("clickhouse insert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.user != "" {
		req.SetBasicAuth(s.user, s.password)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("clickhouse insert: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("clickhouse insert: status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return nil
}

// Ping checks that ClickHouse is reachable.
func (s *ClickHouseStore) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/ping", nil)
	if err != nil {
		return fmt.Errorf("clickhouse ping request: %w", err)
	}
	if s.user != "" {
		req.SetBasicAuth(s.user, s.password)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("clickhouse ping: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clickhouse ping: status %d", resp.StatusCode)
	}
	return nil
}

var _ Store = (*ClickHouseStore)(nil)
