package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/OlegRozh/url-shortener/internal/lib/api/response"
	"github.com/OlegRozh/url-shortener/internal/lib/logger/sl"
	"github.com/OlegRozh/url-shortener/internal/lib/random"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

var ErrURLExists = errors.New("url already exists")

type Request struct {
	URL string `json:"url" validate:"required,url"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}

const (
	aliasLength = 10
	maxRetries  = 3
)

type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		// 1. Декодируем запрос
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			resp.JSON(w, r, http.StatusBadRequest, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		// 2. Валидируем URL
		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			resp.JSON(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))
			return
		}

		var (
			id    int64
			alias string
		)

		// 3. Пытаемся сохранить с генерацией алиаса (retry при конфликте)
		for i := 0; i < maxRetries; i++ {
			// Генерируем случайный алиас
			alias, err = random.NewRandomString(aliasLength)
			if err != nil {
				log.Error("failed to generate random alias", sl.Err(err))
				resp.JSON(w, r, http.StatusInternalServerError, resp.Error("internal server error"))
				return
			}

			// Пытаемся сохранить
			id, err = urlSaver.SaveURL(r.Context(), req.URL, alias)
			if err == nil {
				// Успех - выходим из цикла
				break
			}

			// Если ошибка не из-за конфликта алиаса - выходим с ошибкой
			if !errors.Is(err, ErrURLExists) {
				log.Error("failed to save url", sl.Err(err))
				resp.JSON(w, r, http.StatusInternalServerError, resp.Error("failed to save url"))
				return
			}

			// Конфликт алиаса - пробуем снова с новым алиасом
			log.Info("alias conflict, retrying", slog.String("alias", alias), slog.Int("attempt", i+1))
		}

		// Если после всех retry всё равно ошибка
		if err != nil {
			log.Error("failed to save url after retries", sl.Err(err))
			resp.JSON(w, r, http.StatusInternalServerError, resp.Error("failed to save url"))
			return
		}

		// 4. Успешный ответ
		log.Info("url saved successfully", slog.Int64("id", id), slog.String("alias", alias))
		resp.JSON(w, r, http.StatusCreated, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
