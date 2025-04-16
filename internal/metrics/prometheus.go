package metrics

import (
	"net/http"
	"pvz-service-avito-internship/internal/domain"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto" // Используем promauto для автоматической регистрации
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// Импорт домена не требуется для этого файла, так как интерфейс MetricsCollector определен в domain/interfaces.go
	// "pvz-service-avito-internship/internal/domain"
)

// collector - это конкретная реализация интерфейса domain.MetricsCollector.
// Она содержит счетчики и гистограммы Prometheus.
// Эта структура не экспортируется, взаимодействие происходит через интерфейс.
type collector struct {
	// Технические метрики HTTP
	requestsTotal          *prometheus.CounterVec   // Общее количество HTTP запросов
	requestDurationSeconds *prometheus.HistogramVec // Распределение времени ответа HTTP запросов

	// Бизнесовые метрики
	pvzCreatedTotal        prometheus.Counter // Общее количество созданных ПВЗ
	receptionsCreatedTotal prometheus.Counter // Общее количество созданных Приемок
	productsAddedTotal     prometheus.Counter // Общее количество добавленных Товаров
}

// NewCollector создает новый экземпляр коллектора метрик.
// Он инициализирует все метрики с помощью promauto, что автоматически
// регистрирует их в стандартном регистре Prometheus (prometheus.DefaultRegisterer).
// Возвращает интерфейс domain.MetricsCollector для использования в других частях приложения.
func NewCollector() domain.MetricsCollector { // Возвращаем интерфейс из domain
	c := &collector{
		// --- Инициализация Технических Метрик ---
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pvz_http_requests_total",                  // Имя метрики (префикс pvz_)
				Help: "Total number of processed HTTP requests.", // Описание метрики
			},
			[]string{"method", "path", "status_code"}, // Лейблы для разделения метрик
		),
		requestDurationSeconds: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "pvz_http_request_duration_seconds",               // Имя метрики
				Help: "Histogram of HTTP request durations in seconds.", // Описание
				// Buckets определяют корзины для гистограммы времени ответа.
				// prometheus.DefBuckets - стандартный набор бакетов (от 0.005 до 10 секунд).
				// Можно задать свои бакеты: Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"}, // Лейблы (статус код здесь не нужен, т.к. это распределение времени)
		),

		// --- Инициализация Бизнесовых Метрик ---
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
	// Возвращаем созданный коллектор как интерфейс domain.MetricsCollector.
	// Это позволяет легко мокировать метрики в юнит-тестах сервисов.
	return c
}

// --- Реализация методов интерфейса domain.MetricsCollector ---

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

// RunMetricsServer создает и возвращает сконфигурированный http.Server,
// который будет отдавать собранные метрики Prometheus по адресу addr/metrics.
// Этот сервер должен быть запущен в отдельной горутине.
func RunMetricsServer(addr string) *http.Server {
	// Создаем новый ServeMux, чтобы не использовать DefaultServeMux (более безопасно).
	mux := http.NewServeMux()

	// Регистрируем стандартный обработчик Prometheus (/metrics), который использует
	// prometheus.DefaultRegisterer, где были зарегистрированы наши метрики через promauto.
	mux.Handle("/metrics", promhttp.Handler())

	// Создаем HTTP сервер с указанным адресом и нашим mux'ом.
	server := &http.Server{
		Addr:    addr, // Адрес вида ":9000"
		Handler: mux,
	}

	// Возвращаем сервер, готовый к запуску через ListenAndServe().
	return server
}
