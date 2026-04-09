package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
	}{
		{
			name:           "OK response",
			status:         http.StatusOK,
			data:           Response{Status: StatusOK},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Created response",
			status:         http.StatusCreated,
			data:           Response{Status: StatusOK},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Error response",
			status:         http.StatusBadRequest,
			data:           Response{Status: StatusError, Error: "something went wrong"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			JSON(w, r, tt.status, tt.data)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			var resp Response
			err := json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
		})
	}
}

func TestOK(t *testing.T) {
	resp := OK()
	assert.Equal(t, StatusOK, resp.Status)
	assert.Empty(t, resp.Error)
}

func TestError(t *testing.T) {
	msg := "test error message"
	resp := Error(msg)
	assert.Equal(t, StatusError, resp.Status)
	assert.Equal(t, msg, resp.Error)
}

func TestValidationError(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name          string
		value         interface{}
		tag           string
		expectedError string
	}{
		{
			name:          "Required field",
			value:         "",
			tag:           "required",
			expectedError: "field  is a required field",
		},
		{
			name:          "Invalid URL",
			value:         "not-a-url",
			tag:           "url",
			expectedError: "field  is not a valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Var(tt.value, tt.tag)
			require.Error(t, err)

			var errs validator.ValidationErrors
			errors.As(err, &errs)
			resp := ValidationError(errs)

			assert.Equal(t, StatusError, resp.Status)
			assert.Contains(t, resp.Error, tt.expectedError)
		})
	}
}

func TestValidationError_MultipleErrors(t *testing.T) {
	type TestStruct struct {
		URL   string `validate:"required,url"`
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	validate := validator.New()
	testStruct := TestStruct{
		URL:   "invalid",
		Name:  "",
		Email: "not-an-email",
	}

	err := validate.Struct(testStruct)
	require.Error(t, err)

	var errs validator.ValidationErrors
	errors.As(err, &errs)
	resp := ValidationError(errs)

	assert.Equal(t, StatusError, resp.Status)
	assert.NotEmpty(t, resp.Error)

	// Проверяем, что все три ошибки присутствуют
	expectedErrors := []string{"URL", "Name", "Email"}
	for _, expected := range expectedErrors {
		assert.Contains(t, resp.Error, expected)
	}
}
