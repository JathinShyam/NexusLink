package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	httpRequests      *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
	cacheHits         prometheus.Counter
	cacheMisses       prometheus.Counter
	rateLimitExceeded *prometheus.CounterVec
	clickEventsOK     prometheus.Counter
	clickEventsDrop   prometheus.Counter
	clickEventsErr    prometheus.Counter
}

func New() *Metrics {
	return &Metrics{
		httpRequests: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		}, []string{"method", "path", "status"}),
		httpDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
		cacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Redirect cache hits",
		}),
		cacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Redirect cache misses",
		}),
		rateLimitExceeded: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Rate limit rejections",
		}, []string{"endpoint"}),
		clickEventsOK: promauto.NewCounter(prometheus.CounterOpts{
			Name: "click_events_published_total",
			Help: "Click events delivered to the analytics sink",
		}),
		clickEventsDrop: promauto.NewCounter(prometheus.CounterOpts{
			Name: "click_events_dropped_total",
			Help: "Click events dropped because the buffer was full",
		}),
		clickEventsErr: promauto.NewCounter(prometheus.CounterOpts{
			Name: "click_events_errors_total",
			Help: "Click events that failed to reach the analytics sink",
		}),
	}
}

func (m *Metrics) CacheHit()  { m.cacheHits.Inc() }
func (m *Metrics) CacheMiss() { m.cacheMisses.Inc() }

func (m *Metrics) ClickEventPublished() { m.clickEventsOK.Inc() }
func (m *Metrics) ClickEventDropped()   { m.clickEventsDrop.Inc() }
func (m *Metrics) ClickEventError()     { m.clickEventsErr.Inc() }

func (m *Metrics) RateLimitExceeded(endpoint string) {
	m.rateLimitExceeded.WithLabelValues(endpoint).Inc()
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		path := r.URL.Path
		m.httpRequests.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		m.httpDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
