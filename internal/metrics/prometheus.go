package metrics

import (
	"net/http"
	"pvz-service-avito-internship/internal/domain"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type collector struct {
	requestsTotal          *prometheus.CounterVec   // Общее количество HTTP запросов
	requestDurationSeconds *prometheus.HistogramVec // Распределение времени ответа HTTP запросов

	pvzCreatedTotal        prometheus.Counter // Общее количество созданных ПВЗ
	receptionsCreatedTotal prometheus.Counter // Общее количество созданных Приемок
	productsAddedTotal     prometheus.Counter // Общее количество добавленных Товаров
}

func NewCollector() domain.MetricsCollector {
	c := &collector{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pvz_http_requests_total",
				Help: "Total number of processed HTTP requests.",
			},
			[]string{"method", "path", "status_code"},
		),
		requestDurationSeconds: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "pvz_http_request_duration_seconds",
				Help:    "Histogram of HTTP request durations in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		pvzCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_created_total",
				Help: "Total number of PVZs created.",
			},
		),
		receptionsCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_receptions_created_total",
				Help: "Total number of receptions created.",
			},
		),
		productsAddedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_products_added_total",
				Help: "Total number of products added.",
			},
		),
	}
	return c
}

// IncRequestsTotal увеличивает счетчик общего количества HTTP запросов.
func (c *collector) IncRequestsTotal(method, path, statusCode string) {
	c.requestsTotal.WithLabelValues(method, path, statusCode).Inc()
}

// ObserveRequestDuration записывает значение времени выполнения HTTP запроса в гистограмму.
func (c *collector) ObserveRequestDuration(method, path string, duration float64) {
	c.requestDurationSeconds.WithLabelValues(method, path).Observe(duration)
}

// IncPVZCreated увеличивает счетчик созданных ПВЗ.
func (c *collector) IncPVZCreated() {
	c.pvzCreatedTotal.Inc()
}

// IncReceptionsCreated увеличивает счетчик созданных Приемок.
func (c *collector) IncReceptionsCreated() {
	c.receptionsCreatedTotal.Inc()
}

// IncProductsAdded увеличивает счетчик добавленных Товаров.
func (c *collector) IncProductsAdded() {
	c.productsAddedTotal.Inc()
}

// RunMetricsServer создает и возвращает сконфигурированный http.Server
func RunMetricsServer(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server
}
