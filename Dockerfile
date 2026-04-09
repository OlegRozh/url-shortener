# Этап 1: Сборка
FROM golang:1.25.1-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o url-shortener ./cmd/main.go

# Этап 2: Финальный образ
FROM alpine:latest

# Устанавливаем CA-сертификаты и timezone
RUN apk --no-cache add ca-certificates tzdata

# Создаём рабочую директорию
WORKDIR /app

# Копируем бинарник из этапа сборки
COPY --from=builder /app/url-shortener .

# Копируем конфиги и миграции
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Открываем порт
EXPOSE 8082

# Запускаем приложение
CMD ["./url-shortener"]