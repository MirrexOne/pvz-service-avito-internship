package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	// Используем интерфейс MetricsCollector для слабой связности и тестируемости
	"pvz-service-avito-internship/internal/domain"
)

// PrometheusMiddleware - это Gin middleware для сбора стандартных HTTP метрик Prometheus.
func PrometheusMiddleware(collector domain.MetricsCollector) gin.HandlerFunc { // Принимаем интерфейс
	return func(c *gin.Context) {
		start := time.Now() // Засекаем время начала обработки

		// Получаем шаблон пути (например, /pvz/:pvzId) для группировки метрик
		// Если роут не найден, FullPath будет пуст, используем фактический URL.Path
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path // Fallback
		}
		method := c.Request.Method

		// Передаем управление следующему обработчику
		c.Next()

		// Код ниже выполняется после отработки всех хендлеров

		duration := time.Since(start).Seconds() // Вычисляем длительность в секундах
		statusCode := c.Writer.Status()         // Получаем HTTP статус ответа

		// Увеличиваем счетчик запросов, используя метод интерфейса
		collector.IncRequestsTotal(method, path, strconv.Itoa(statusCode))

		// Наблюдаем длительность запроса, используя метод интерфейса
		collector.ObserveRequestDuration(method, path, duration)
	}
}
