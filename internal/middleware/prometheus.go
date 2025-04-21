package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"pvz-service-avito-internship/internal/domain"
)

func PrometheusMiddleware(collector domain.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()

		collector.IncRequestsTotal(method, path, strconv.Itoa(statusCode))

		collector.ObserveRequestDuration(method, path, duration)
	}
}
