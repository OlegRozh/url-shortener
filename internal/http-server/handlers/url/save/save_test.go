package save

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OlegRozh/url-shortener/internal/http-server/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func TestSaveHandler(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedAlias  bool
	}{
		{
			name:           "Valid URL",
			url:            "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedAlias:  true,
		},
		{
			name:           "Invalid URL",
			url:            "not-a-url",
			expectedStatus: http.StatusBadRequest,
			expectedAlias:  false,
		},
		{
			name:           "Empty URL",
			url:            "",
			expectedStatus: http.StatusBadRequest,
			expectedAlias:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockStorage()
			handler := New(testLogger(), mockStorage)

			body := Request{URL: tt.url}
			jsonBody, err := json.Marshal(body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/urls/url", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedAlias {
				var resp Response
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.Alias, "alias should not be empty")
			}
		})
	}
}

func TestSaveHandlerWithExistingAlias(t *testing.T) {
	mockStorage := mocks.NewMockStorage()
	handler := New(testLogger(), mockStorage)

	// Сохраняем первый URL
	body1 := Request{URL: "https://example1.com"}
	jsonBody1, err := json.Marshal(body1)
	require.NoError(t, err)

	req1 := httptest.NewRequest(http.MethodPost, "/urls/url", bytes.NewReader(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	var resp1 Response
	err = json.NewDecoder(w1.Body).Decode(&resp1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Сохраняем второй URL
	body2 := Request{URL: "https://example2.com"}
	jsonBody2, err := json.Marshal(body2)
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodPost, "/urls/url", bytes.NewReader(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler(w2, req2)

	var resp2 Response
	err = json.NewDecoder(w2.Body).Decode(&resp2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// Алиасы должны быть разными
	assert.NotEqual(t, resp1.Alias, resp2.Alias, "aliases should be different for different URLs")
}
