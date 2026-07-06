package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/logging"
	"github.com/JathinShyam/NexusLink/internal/metrics"
)

type ReadyChecker func(ctx context.Context) error

type Options struct {
	Shorten  http.Handler
	Redirect http.Handler
	Delete   http.Handler
	Metrics  *metrics.Metrics
	Ready    ReadyChecker
}

type Server struct {
	log  *slog.Logger
	mux  *http.ServeMux
	http *http.Server
}

func New(log *slog.Logger, addr string, opts Options) *Server {
	mux := http.NewServeMux()
	s := &Server{
		log: log,
		mux: mux,
		http: &http.Server{
			Addr:              addr,
			Handler:           buildHandler(log, opts.Metrics, mux),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	mux.HandleFunc("GET /health", s.handleHealth)
	if opts.Ready != nil {
		mux.HandleFunc("GET /ready", s.handleReady(opts.Ready))
	}
	if opts.Metrics != nil {
		mux.Handle("GET /metrics", opts.Metrics.Handler())
	}
	if opts.Shorten != nil {
		mux.Handle("POST /api/v1/shorten", opts.Shorten)
	}
	if opts.Delete != nil {
		mux.Handle("DELETE /api/v1/links/{short_code}", opts.Delete)
	}
	if opts.Redirect != nil {
		mux.Handle("GET /{short_code}", opts.Redirect)
	}

	return s
}

func buildHandler(log *slog.Logger, m *metrics.Metrics, mux *http.ServeMux) http.Handler {
	var handler http.Handler = mux
	handler = requestLogger(log, handler)
	if m != nil {
		handler = m.Middleware(handler)
	}
	return handler
}

func (s *Server) Run() error {
	s.log.Info("server listening", slog.String("addr", s.http.Addr))
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("server shutting down")
	return s.http.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReady(check ReadyChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := check(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		logging.LogHTTPRequest(log, r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
