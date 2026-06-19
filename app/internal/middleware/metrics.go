package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/PKSonlem/testtask-secunda-api/pkg/metrics"
)

func Metrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			metrics.RequestsTotal.
				WithLabelValues(c.Request().Method, c.Path(), strconv.Itoa(c.Response().Status)).
				Inc()
			metrics.RequestDuration.
				WithLabelValues(c.Request().Method, c.Path()).
				Observe(time.Since(start).Seconds())
			return err
		}
	}
}
