package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/OlegRozh/url-shortener/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

type Storage struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
)

func NewPostgresStorage(ctx context.Context, cfg *config.Config) (*Storage, error) {
	connURL := cfg.DatabaseURL

	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	if err := storage.Migrate(ctx, connURL); err != nil {
		pool.Close()
		fmt.Printf("Migration failed: %v\n", err)
		return nil, err
	}
	return storage, nil
}

func (s *Storage) Migrate(ctx context.Context, connURL string) error {
	db, err := sql.Open("pgx", connURL)
	if err != nil {
		return fmt.Errorf("failed to open db for migrations: %w", err)
	}
	defer db.Close()
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}
	return nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// SaveURL сохраняет URL с указанным алиасом.
// Если алиас уже существует, возвращает ошибку ErrURLExists.

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	query, args, err := s.builder.
		Insert("url").
		Columns("alias", "url").
		Values(alias, urlToSave).
		Suffix("ON CONFLICT (alias) DO NOTHING RETURNING id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var id int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrURLExists
		}
		return 0, fmt.Errorf("failed to save url: %w", err)
	}
	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	query, args, err := s.builder.
		Select("url").
		From("url").
		Where(squirrel.Eq{"alias": alias}).
		Limit(1).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}

	var url string
	err = s.pool.QueryRow(ctx, query, args...).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}
		return "", fmt.Errorf("failed to get url: %w", err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	query, args, err := s.builder.
		Delete("url").
		Where(squirrel.Eq{"alias": alias}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	cmdTag, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete url: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrURLNotFound
	}
	return nil
}
