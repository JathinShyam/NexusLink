package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/logging"
	"github.com/JathinShyam/NexusLink/internal/redirect"
	"github.com/JathinShyam/NexusLink/internal/shorten"
)

type Options struct {
	Shorten  *shorten.Handler
	Redirect *redirect.Handler
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
			Handler:           requestLogger(log, mux),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	mux.HandleFunc("GET /health", s.handleHealth)

	if opts.Shorten != nil {
		mux.HandleFunc("POST /api/v1/shorten", opts.Shorten.Create)
	}
	if opts.Redirect != nil {
		mux.HandleFunc("GET /{short_code}", func(w http.ResponseWriter, r *http.Request) {
			opts.Redirect.Serve(w, r)
		})
	}

	return s
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
