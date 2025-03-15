package http

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func metricsMiddleware(histogram *prometheus.HistogramVec, counter *prometheus.CounterVec) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			start := time.Now()
			err := next(context)
			if err != nil {
				// populate the response
				context.Error(err)
			}
			duration := time.Since(start)
			method := context.Request().Method
			path := context.Path()
			response := context.Response()
			if response != nil {
				status := fmt.Sprintf("%d", context.Response().Status)
				if status == "404" {
					path = "?"
				}
				histogram.With(prometheus.Labels{"method": method, "path": path}).Observe(duration.Seconds())
				counter.With(prometheus.Labels{"method": method, "status": status, "path": path}).Inc()
			} else {
				slog.Error(fmt.Sprintf("Response in metrics middleware is nil for %s %s", method, path))
			}
			// return nil car we already called the error handler middleware here
			return nil
		}
	}
}
