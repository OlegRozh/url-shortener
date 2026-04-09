package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		wantErr   bool
		checkSize bool
	}{
		{name: "Size 6", size: 6, wantErr: false, checkSize: true},
		{name: "Size 10", size: 10, wantErr: false, checkSize: true},
		{name: "Size 20", size: 20, wantErr: false, checkSize: true},
		{name: "Size 0", size: 0, wantErr: false, checkSize: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := NewRandomString(tt.size)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkSize {
					assert.Len(t, str, tt.size)
				}
			}
		})
	}
}

func TestNewRandomString_Uniqueness(t *testing.T) {
	generated := make(map[string]bool)
	iterations := 10000
	size := 10

	for i := 0; i < iterations; i++ {
		str, err := NewRandomString(size)
		require.NoError(t, err)
		assert.False(t, generated[str], "duplicate string generated: %s", str)
		generated[str] = true
	}

	assert.Len(t, generated, iterations, "should have %d unique strings", iterations)
}

func TestNewRandomString_CharacterSet(t *testing.T) {
	str, err := NewRandomString(100)
	require.NoError(t, err)

	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	allowedSet := make(map[byte]bool)
	for i := 0; i < len(allowedChars); i++ {
		allowedSet[allowedChars[i]] = true
	}

	for i := 0; i < len(str); i++ {
		assert.True(t, allowedSet[str[i]], "invalid character '%c' in generated string", str[i])
	}
}

func TestNewRandomString_DifferentLengths(t *testing.T) {
	lengths := []int{1, 5, 10, 15, 20, 50}
	for _, length := range lengths {
		t.Run("Length_"+string(rune(length)), func(t *testing.T) {
			str, err := NewRandomString(length)
			require.NoError(t, err)
			assert.Len(t, str, length)
		})
	}
}
