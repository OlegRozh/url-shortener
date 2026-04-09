package redirect

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

type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// 1. Получаем alias из URL (например, /goog → alias = "goog")
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			resp.JSON(w, r, http.StatusBadRequest, resp.Error("invalid request"))
			return
		}

		log.Info("redirecting", slog.String("alias", alias))

		// 2. Ищем оригинальный URL в БД
		originalURL, err := urlGetter.GetURL(r.Context(), alias)
		if err != nil {
			if errors.Is(err, ErrURLNotFound) {
				log.Info("alias not found", slog.String("alias", alias))
				resp.JSON(w, r, http.StatusNotFound, resp.Error("url not found"))
				return
			}
			log.Error("failed to get url", sl.Err(err))
			resp.JSON(w, r, http.StatusInternalServerError, resp.Error("internal error"))
			return
		}
		// 3. Редирект
		log.Info("redirecting", slog.String("URL", originalURL))
		http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
	}
}
