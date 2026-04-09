package delete

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

func TestDeleteHandler(t *testing.T) {
	tests := []struct {
		name           string
		alias          string
		setupMock      func(*mocks.MockStorage)
		expectedStatus int
	}{
		{
			name:  "Existing alias",
			alias: "test123",
			setupMock: func(m *mocks.MockStorage) {
				_, err := m.SaveURL(context.Background(), "https://example.com", "test123")
				require.NoError(t, err)
			},
			expectedStatus: http.StatusOK, // 200
		},
		{
			name:           "Non-existing alias",
			alias:          "nonexistent",
			setupMock:      func(m *mocks.MockStorage) {},
			expectedStatus: http.StatusInternalServerError, // 500 (мок возвращает ошибку без типа)
		},
		{
			name:           "Empty alias",
			alias:          "",
			setupMock:      func(m *mocks.MockStorage) {},
			expectedStatus: http.StatusBadRequest, // 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockStorage()
			tt.setupMock(mockStorage)

			handler := New(testLogger(), mockStorage)

			req := httptest.NewRequest(http.MethodDelete, "/urls/"+tt.alias, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tt.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDeleteHandler_Idempotency(t *testing.T) {
	mockStorage := mocks.NewMockStorage()
	handler := New(testLogger(), mockStorage)

	alias := "test123"
	_, err := mockStorage.SaveURL(context.Background(), "https://example.com", alias)
	require.NoError(t, err)

	addAliasToRequest := func(req *http.Request, alias string) *http.Request {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("alias", alias)
		return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	// Первое удаление - успешно
	req1 := addAliasToRequest(httptest.NewRequest(http.MethodDelete, "/urls/"+alias, nil), alias)
	w1 := httptest.NewRecorder()
	handler(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Второе удаление - мок возвращает ошибку → 500
	req2 := addAliasToRequest(httptest.NewRequest(http.MethodDelete, "/urls/"+alias, nil), alias)
	w2 := httptest.NewRecorder()
	handler(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}
