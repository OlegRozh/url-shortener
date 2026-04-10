package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OlegRozh/url-shortener/internal/config"
	"github.com/OlegRozh/url-shortener/internal/http-server/handlers/redirect"
	"github.com/OlegRozh/url-shortener/internal/http-server/handlers/url/delete"
	"github.com/OlegRozh/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/OlegRozh/url-shortener/internal/http-server/middleware/logger"
	"github.com/OlegRozh/url-shortener/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
	cfg        *config.Config
	storage    Storage
}

// Storage интерфейс для хранилища
type Storage interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) error
	Ping(ctx context.Context) error // 👈 добавить
	Close()
}

// New создаёт новый сервер (конструктор)
func New(log *slog.Logger, cfg *config.Config, storage Storage) *Server {
	return &Server{
		log:     log,
		cfg:     cfg,
		storage: storage,
	}
}

// setupRouter настраивает маршруты
func (s *Server) setupRouter() *chi.Mux {
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(s.log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Timeout(60 * time.Second))

	// Healthcheck
	//Healthcheck
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{"status": "ok"}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			// Если ошибка, логируем и отправляем 500
			s.log.Error("failed to encode health response", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Protected routes
	router.Route("/urls", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			s.cfg.User: s.cfg.Password,
		}))
		r.Post("/url", save.New(s.log, s.storage))
		r.Delete("/{alias}", delete.New(s.log, s.storage))
	})

	// Public routes
	router.Get("/{alias}", redirect.New(s.log, s.storage))

	return router
}

func (s *Server) Start() error {
	router := s.setupRouter()

	s.httpServer = &http.Server{
		Addr:         s.cfg.Address,
		Handler:      router,
		ReadTimeout:  s.cfg.Timeout,
		WriteTimeout: s.cfg.Timeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	go func() {
		s.log.Info("starting server", slog.String("address", s.cfg.Address))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("failed to start server", sl.Err(err))
		}
	}()

	// Ждём сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.log.Info("shutting down server gracefully...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("server forced to shutdown", sl.Err(err))
		return err
	}

	s.log.Info("server stopped")
	return nil
}
