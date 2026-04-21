## URL-Shortener
#### Программа представляет собой Rest API сервис для сокращения ссылок.

**Запустить локально:**
```bash
git clone ... && cd url-shortener
docker-compose up
# API: http://localhost:8080
```

- Все запросы и ответы используют формат `JSON`.
- Устанавливайте заголовок: Content-Type: application/json

API будет доступен по адресу:  
👉 [http://localhost:8080](http://localhost:8080)

> 💡 Перед запуском создайте `.env` файл (пример ниже).
>
```POSTGRES_DB=urlshortener
//Данные подключения
POSTGRES_USER=myuser
POSTGRES_PASSWORD=url12345
//Путь для Go-приложения
DATABASE_URL=postgres://myuser:url12345@localhost:5432/urlshortener?sslmode=disable
//Для мигаций
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="user=myuser password=url12345 dbname=urlshortener host=localhost port=5432 sslmode=disable"
GOOSE_MIGRATION_DIR=./migrations
CONFIG_PATH=./config/local.yaml
//Данные авторизации
HTTP_SERVER_USER=admin
HTTP_SERVER_PASSWORD=qwerty1234
```

## 1. 🔗 Сократить URL

Создаёт короткую ссылку с автоматически сгенерированным алиасом.

### **Метод**: POST /url
### **Тело запроса (JSON)**
> json { "url": "https://example.com/very/long/path?param=value" }

> ⚠️ Поле `url` обязательно и должно быть валидным URL.
### **Успешный ответ (201 Created)**
> json { "status": "OK", "alias": "abc123defg" }

Теперь можно перейти по ссылке:  
👉 `http://localhost:8080/abc123defg`

---

## 2. 🔍 Перейти по короткой ссылке

Выполняет редирект на оригинальный URL.

### **Метод**: GET /{alias
### **Пример** GET /abc123defg
### **Ответ**
- **Статус**: `307 Temporary Redirect`
- **Заголовок**: 

Location: https://example.com/very/long/path?param=value
> 🌐 Браузер или HTTP-клиент автоматически перейдёт по оригинальной ссылке.

## 4. 🗑️ Удалить URL по алиасу

Удаляет запись из базы данных.

### **Метод** DELETE /{alias}

Пример: DELETE /abc123defg
Тело запроса не требуется.
### **Успешный ответ (200 OK)**
>json { "status": "OK" }

> ✅ Миграции применяются автоматически при старте сервера.

---

## Технологии и библиотеки

| Библиотека | Назначение |
|----------|-----------|
| `net/http` + `chi` | Роутинг и HTTP-сервер |
| `slog` + `tint` | Логирование (цветное в dev) |
| `pgx/v5` + `pgxpool` | Подключение к PostgreSQL |
| `squirrel` | Построение SQL-запросов без инъекций |
| `validator/v10` | Валидация входных данных |
| `cleanenv` | Чтение конфига из YAML и env |
| `godotenv` | Загрузка `.env` файла |
| `goose` | Миграции базы данных |
| `crypto/rand` | Генерация безопасных алиасов |

---

- **Loki** — хранилище логов
- **Promtail** — сборщик логов из Docker контейнеров
- **Grafana** — визуализация и анализ логов