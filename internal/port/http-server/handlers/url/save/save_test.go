package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/port/http-server/handlers/url/save"
	"url-shortener/internal/port/http-server/handlers/url/save/mocks"

	"github.com/stretchr/testify/require"

	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		status    int
		mockError error
	}{
		{
			name:   "Success",
			alias:  "test_alias",
			url:    "https://google.com",
			status: http.StatusOK,
		},
		{
			name:   "Empty alias",
			alias:  "",
			url:    "https://google.com",
			status: http.StatusOK,
		},
		{
			name:   "Empty URL",
			url:    "",
			alias:  "some_alias",
			status: http.StatusBadRequest,
		},
		{
			name:   "Invalid URL",
			url:    "some invalid URL",
			alias:  "some_alias",
			status: http.StatusBadRequest,
		},
		{
			name:      "SaveURL Error",
			alias:     "test_alias",
			url:       "https://google.com",
			status:    http.StatusInternalServerError,
			mockError: errors.New("unexpected error"),
		},
	}

	e := echo.New()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.status == http.StatusOK || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			resp := new(save.Response)
			if assert.NoError(t, save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)(c)) {
				assert.Equal(t, tc.status, rec.Code)

				if tc.status == http.StatusOK {
					assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), resp))
					assert.Empty(t, resp.Error)
					assert.NotEmpty(t, resp.Alias)
				}
			}
		})
	}
}
