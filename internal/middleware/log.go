package middleware

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Logger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogMethod:   true,
		HandleError: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {

			fields := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("URI", v.URI),
				slog.Int("status", v.Status),
				slog.String("IP", c.RealIP()),
				slog.String("request", v.RequestID),
				slog.String("latency", v.Latency.String()),
			}

			logger := c.Logger()

			if v.Error != nil && v.Status >= 500 {
				logger.LogAttrs(context.Background(), slog.LevelError, v.Error.Error(), fields...)
			} else if v.Latency.Milliseconds() > 500 {
				logger.LogAttrs(context.Background(), slog.LevelWarn, "SLOW_REQUEST", fields...)
			} else if v.Status != 404 {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST", fields...)
			}

			return nil
		},
	})
}
