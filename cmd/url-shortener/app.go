package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
	"os"
	urlstoragegrpc "url-shortener/internal/adapter/url-storage/grpc"
	"url-shortener/internal/config"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/port/http-server/handlers/redirect"
	"url-shortener/internal/port/http-server/handlers/url/save"
	"url-shortener/internal/port/http-server/middleware/logger"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug mode")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.URLStorage.Timeout)
	defer cancel()
	urlStorage, err := urlstoragegrpc.New(
		ctx, log,
		cfg.URLStorage.Address,
		cfg.URLStorage.Timeout,
		cfg.URLStorage.Retries,
	)
	if err != nil {
		log.Error("failed to create storage", sl.Err(err))
		os.Exit(1)
	}

	e := echo.New()

	e.Use(mwLogger.New(log))
	e.Use(middleware.Recover())

	e.GET("/:alias", redirect.New(log, urlStorage))

	urlGroup := e.Group("/url")

	urlGroup.POST("", save.New(log, urlStorage))
	//urlGroup.DELETE("/:alias", del.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      e,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
