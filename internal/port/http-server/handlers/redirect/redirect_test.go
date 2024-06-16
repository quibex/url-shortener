package redirect_test

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/port/http-server/handlers/redirect"
	"url-shortener/internal/port/http-server/handlers/redirect/mocks"
	"url-shortener/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		alias     string
		status    int
		mockError error
	}{
		{
			name:   "Success",
			url:    "https://www.google.com/",
			alias:  "test_alias",
			status: http.StatusFound,
		},
		{
			name:      "URL not found",
			alias:     "test_alias",
			status:    http.StatusNotFound,
			mockError: storage.ErrURLNotFound,
		},
		{
			name:      "Internal server error",
			alias:     "test_alias",
			status:    http.StatusInternalServerError,
			mockError: errors.New("internal error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocks.NewURLGetter(t)
			urlGetterMock.On("GetURL", tc.alias).Return(tc.url, tc.mockError).Once()

			req, err := http.NewRequest(http.MethodGet, "/"+tc.alias, nil)
			assert.NoError(t, err)

			rec := httptest.NewRecorder()

			e := echo.New()
			e.GET("/:alias", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			if tc.status == http.StatusFound {
				location := rec.Header().Get("Location")
				assert.Equal(t, tc.url, location)
			}
		})
	}
}
