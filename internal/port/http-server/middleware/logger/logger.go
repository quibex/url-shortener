package mwLogger

import (
	"github.com/labstack/echo/v4"
	"log/slog"
	"time"
)

func New(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log := log.With(
				slog.String("component", "middleware/logger"),
			)

			//log.Info("logger middleware enabled")

			req := c.Request()
			res := c.Response()

			entry := log.With(
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("remote_addr", req.RemoteAddr),
				slog.String("user_agent", req.UserAgent()),
				slog.String("request_id", req.Header.Get(echo.HeaderXRequestID)),
			)

			t1 := time.Now()
			err := next(c)
			dur := time.Since(t1)
			if err != nil {
				c.Error(err)
			}
			entry.Info("request completed",
				slog.Int("status", res.Status),
				slog.Int64("bytes", res.Size),
				slog.String("duration", dur.String()),
			)
			return nil
		}
	}
}
