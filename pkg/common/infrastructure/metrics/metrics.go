package metrics

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewPrometheusMetricsHandler(appID string) (PrometheusMetricsHandler, error) {
	labelNames := []string{"endpoint", "method"}

	latencyHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_request_latency_seconds",
		Help:    "Application Request Latency.",
		Buckets: prometheus.DefBuckets,
	}, labelNames)
	err := prometheus.Register(latencyHistogram)
	if err != nil {
		return nil, err
	}

	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_request_count",
		Help: "Application Request Count.",
	}, labelNames)
	err = prometheus.Register(requestCounter)
	if err != nil {
		return nil, err
	}

	serverErrorCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_server_error_count",
		Help: "Application Server Error (5xx) Count.",
	}, labelNames)
	err = prometheus.Register(serverErrorCounter)
	if err != nil {
		return nil, err
	}

	return &prometheusMetricsHandler{
		latencyHistogram:   latencyHistogram,
		requestCounter:     requestCounter,
		serverErrorCounter: serverErrorCounter,
		appID:              appID,
	}, nil
}

type PrometheusMetricsHandler interface {
	AddMetricsMiddleware(router *mux.Router)
}

type prometheusMetricsHandler struct {
	latencyHistogram   *prometheus.HistogramVec
	requestCounter     *prometheus.CounterVec
	serverErrorCounter *prometheus.CounterVec
	appID              string
}

func (p *prometheusMetricsHandler) AddMetricsMiddleware(router *mux.Router) {
	router.Handle("/metrics", promhttp.Handler())

	router.Use(p.prometheusMiddleware())
}

func (p *prometheusMetricsHandler) prometheusMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			code := http.StatusOK

			defer func() {
				httpDuration := time.Since(start)
				var labels []string
				re := regexp.MustCompile("(^*)/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
				url := re.ReplaceAllString(req.URL.Path, `$1`)
				labels = append(labels, "/"+p.appID+url)
				labels = append(labels, req.Method)
				p.latencyHistogram.WithLabelValues(labels...).Observe(httpDuration.Seconds())
				p.requestCounter.WithLabelValues(labels...).Inc()
				if code >= 500 {
					p.serverErrorCounter.WithLabelValues(labels...).Inc()
				}
			}()

			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, req)
			code = rw.statusCode
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
