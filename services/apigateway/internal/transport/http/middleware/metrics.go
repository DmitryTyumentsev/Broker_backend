package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type PrometheusMetrics struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewPrometheusMetrics(serviceName string) *PrometheusMetrics {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		serviceName = "apigateway"
	}

	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "broker",
			Subsystem: serviceName,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "broker",
			Subsystem: serviceName,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		requestsTotal,
		requestDuration,
	)

	return &PrometheusMetrics{
		registry:        registry,
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
	}
}

func (m *PrometheusMetrics) Middleware() fiber.Handler {
	if m == nil {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()

		statusCode := c.Response().StatusCode()
		if err != nil && statusCode < fiber.StatusBadRequest {
			statusCode = fiber.StatusInternalServerError
		}

		status := strconv.Itoa(statusCode)
		route := routePattern(c)

		m.requestsTotal.WithLabelValues(c.Method(), route, status).Inc()
		m.requestDuration.WithLabelValues(c.Method(), route, status).Observe(time.Since(startedAt).Seconds())

		return err
	}
}

func (m *PrometheusMetrics) Handler() fiber.Handler {
	handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
	fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(handler)

	return func(c *fiber.Ctx) error {
		fasthttpHandler(c.Context())
		return nil
	}
}

func routePattern(c *fiber.Ctx) string {
	route := c.Route()
	if route != nil && route.Path != "" {
		return route.Path
	}

	return c.Path()
}
