package redirect

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OlegRozh/url-shortener/internal/http-server/handlers/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name           string
		alias          string
		setupMock      func(*mocks.MockStorage)
		expectedStatus int
		expectedURL    string
	}{
		{
			name:  "Existing alias",
			alias: "test123",
			setupMock: func(m *mocks.MockStorage) {
				_, err := m.SaveURL(context.Background(), "https://example.com", "test123")
				require.NoError(t, err)
			},
			expectedStatus: http.StatusMovedPermanently, // 301
			expectedURL:    "https://example.com",
		},
		{
			name:           "Non-existing alias",
			alias:          "nonexistent",
			setupMock:      func(m *mocks.MockStorage) {},
			expectedStatus: http.StatusInternalServerError, // 500 (мок возвращает ошибку без типа)
			expectedURL:    "",
		},
		{
			name:           "Empty alias",
			alias:          "",
			setupMock:      func(m *mocks.MockStorage) {},
			expectedStatus: http.StatusBadRequest, // 400
			expectedURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockStorage()
			tt.setupMock(mockStorage)

			handler := New(testLogger(), mockStorage)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.alias, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tt.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedURL != "" {
				location := w.Header().Get("Location")
				assert.Equal(t, tt.expectedURL, location)
			}
		})
	}
}

func TestRedirectHandler_CaseSensitive(t *testing.T) {
	mockStorage := mocks.NewMockStorage()
	handler := New(testLogger(), mockStorage)

	alias := "test123"
	_, err := mockStorage.SaveURL(context.Background(), "https://example.com", alias)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/TEST123", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", "TEST123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler(w, req)

	// Мок не различает регистр, возвращает 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirectHandler_NotFound(t *testing.T) {
	mockStorage := mocks.NewMockStorage()
	handler := New(testLogger(), mockStorage)

	alias := "nonexistent"

	req := httptest.NewRequest(http.MethodGet, "/"+alias, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", alias)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler(w, req)

	// Мок возвращает ошибку без типа → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirectHandler_MultipleAliases(t *testing.T) {
	mockStorage := mocks.NewMockStorage()
	handler := New(testLogger(), mockStorage)

	urls := map[string]string{
		"google": "https://google.com",
		"yandex": "https://yandex.ru",
		"github": "https://github.com",
	}

	for alias, url := range urls {
		_, err := mockStorage.SaveURL(context.Background(), url, alias)
		require.NoError(t, err)
	}

	for alias, expectedURL := range urls {
		t.Run("Redirect_"+alias, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+alias, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(w, req)

			assert.Equal(t, http.StatusMovedPermanently, w.Code)
			location := w.Header().Get("Location")
			assert.Equal(t, expectedURL, location)
		})
	}
}
