package save

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

const aliasLen = 8

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(ctx context.Context, url string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) echo.HandlerFunc {
	return func(c echo.Context) error {
		const op = "handler.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", c.Request().Header.Get(echo.HeaderXRequestID)),
		)

		req := new(Request)
		if err := json.NewDecoder(c.Request().Body).Decode(req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			return c.JSON(http.StatusBadRequest, Response{
				Error: "failed to decode request body",
			})
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			return c.JSON(http.StatusBadRequest, response.ValidationError(validateErr))
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLen)
		}

		id, err := urlSaver.SaveURL(c.Request().Context(), req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("alias already exists", slog.String("alias", req.URL))

			return c.JSON(http.StatusConflict, Response{Error: "alias already exists"})
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			return c.JSON(http.StatusInternalServerError, Response{Error: "internal error"})
		}

		log.Info("url added", slog.Int64("id", id))

		return c.JSON(http.StatusOK, Response{
			Alias: alias,
		})
	}
}
