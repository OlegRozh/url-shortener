package mocks

import (
	"context"
	"errors"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
)

type MockStorage struct {
	urls map[string]string // alias -> url
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		urls: make(map[string]string),
	}
}

func (m *MockStorage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	if _, exists := m.urls[alias]; exists {
		return 0, ErrURLExists
	}
	m.urls[alias] = urlToSave
	return int64(len(m.urls)), nil
}

func (m *MockStorage) GetURL(ctx context.Context, alias string) (string, error) {
	url, exists := m.urls[alias]
	if !exists {
		return "", ErrURLNotFound
	}
	return url, nil
}

func (m *MockStorage) DeleteURL(ctx context.Context, alias string) error {
	if _, exists := m.urls[alias]; !exists {
		return ErrURLNotFound
	}
	delete(m.urls, alias)
	return nil
}

func (m *MockStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStorage) Close() {}
