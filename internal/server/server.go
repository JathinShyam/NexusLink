package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/logging"
)

type Server struct {
	log  *slog.Logger
	mux  *http.ServeMux
	http *http.Server
}

func New(log *slog.Logger, addr string) *Server {
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
