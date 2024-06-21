package redirect

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) echo.HandlerFunc {
	return func(c echo.Context) error {
		const op = "handlers.url.redirect.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", c.Request().Header.Get(echo.HeaderXRequestID)),
		)

		alias := c.Param("alias")
		// if alias is empty, the request will not get into this handler because of the route definition
		//if alias == "" {
		//	log.Info("alias is empty")
		//	return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		//}
		resURL, err := urlGetter.GetURL(c.Request().Context(), alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			return echo.NewHTTPError(http.StatusNotFound, "not found")
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		log.Info("got url", slog.String("url", resURL))

		return c.Redirect(http.StatusFound, resURL)
	}
}
