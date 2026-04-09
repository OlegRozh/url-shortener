package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/OlegRozh/url-shortener/internal/lib/api/response"
	"github.com/OlegRozh/url-shortener/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var ErrURLNotFound = errors.New("url not found")

type URLDeleter interface {
	DeleteURL(ctx context.Context, alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем alias из URL
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty")
			resp.JSON(w, r, http.StatusBadRequest, resp.Error("invalid request: alias is required"))
			return
		}

		log.Info("deleting url", slog.String("alias", alias))

		// Удаляем URL из БД по alias
		err := urlDeleter.DeleteURL(r.Context(), alias)
		if err != nil {
			if errors.Is(err, ErrURLNotFound) {
				log.Info("url not found", slog.String("alias", alias))
				resp.JSON(w, r, http.StatusNotFound, resp.Error("url not found"))
				return
			}
			log.Error("failed to delete url", sl.Err(err))
			resp.JSON(w, r, http.StatusInternalServerError, resp.Error("internal server error"))
			return
		}

		// Успешный ответ
		log.Info("url deleted successfully", slog.String("alias", alias))
		resp.JSON(w, r, http.StatusOK, resp.OK())
	}
}
