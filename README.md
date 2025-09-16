## simple-http-calendar

Небольшой HTTP‑сервис для работы с календарём событий. Поддерживает CRUD и выборку за день/неделю/месяц. Хранение — in‑memory.

## Быстрый старт
```bash
go mod tidy
go run ./cmd
# или через docker
make up && make logs
```

## Конфигурация (ENV)
- `HTTP_PORT` — порт HTTP сервера, по умолчанию `8080`
- `HTTP_TIMEOUT` — таймауты HTTP сервера, по умолчанию `20s`
- `LOG_LEVEL` — уровень логгирования (`debug|info|warn|error`), по умолчанию `info`

## API
- POST `/create_event` — создать событие
  - Тело: JSON или `application/x-www-form-urlencoded`
  - Поля: `user_id` (int), `date` (YYYY-MM-DD), `event` (string)
  - Успех: `{"result":"<event_id>"}`

- POST `/update_event` — обновить событие
  - Тело: JSON или `application/x-www-form-urlencoded`
  - Поля: `event_id` (string), `user_id` (int), `date` (YYYY-MM-DD), `event` (string)
  - Успех: `{"result":"ok"}`

- POST `/delete_event` — удалить событие
  - Тело: JSON или `application/x-www-form-urlencoded`
  - Поля: `event_id` (string)
  - Успех: `{"result":"ok"}`

- GET `/events_for_day?user_id=1&date=2025-01-02`
  - Успех: `{"result":[ ...events ]}`

- GET `/events_for_week?user_id=1&date=2025-01-02`
  - Успех: `{"result":[ ...events ]}`

- GET `/events_for_month?user_id=1&date=2025-01-02`
  - Успех: `{"result":[ ...events ]}`

Коды ответов:
- 200 — успех
- 400 — ошибки ввода (валидация)
- 503 — ошибки бизнес-логики (например, удаление несуществующего)
- 500 — прочие (recovery)

Примеры curl:
```bash
curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"date":"2025-01-02","event":"meeting"}'

curl -s -X POST http://localhost:8080/update_event \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "event_id=<ID>&user_id=1&date=2025-01-03&event=rescheduled"

curl -s -X POST http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{"event_id":"<ID>"}'

curl -s "http://localhost:8080/events_for_week?user_id=1&date=2025-01-02"
```

## Структура проекта (основные директории)
- `cmd/` — вход (`main.go`)
- `internal/config/` — конфигурация (`Config`, чтение ENV или дефолт значения)
- `internal/logger/` — инициализация логгера
- `internal/server/` — HTTP сервер и graceful shutdown
- `internal/middleware/` — логирование, JSON/Form валидатор, recovery
- `internal/handlers/http/` — HTTP‑хендлеры, модели запросов/ответов, хелперы
- `internal/handlers/validators/` — валидация payload и фильтров
- `internal/interfaces/` — контракты слоёв (`services`, `infra`)
- `internal/services/calendarsvc/` — бизнес-логика и тесты
- `internal/infra/inmem/` — in‑memory реализация инфры
- `internal/entrypoint/` — сборка зависимостей и запуск сервера
- `models/` — доменные модели (`Event`, `EventsByDay`)
- `internal/httpx/` — утилиты HTTP (ответы/ошибки)
- `Dockerfile`, `docker-compose.yml`, `Makefile` — контейнеризация и утилиты

## Тестирование
```bash
go test ./...
```
Покрыты базовые сценарии сервиса: создание/выборка за день, обновление, удаление, ошибки валидации.

## Логи
- Middleware пишет: метод, путь, `duration_ms` в stdout.
- Уровень логов — `LOG_LEVEL` (`info` по умолчанию).