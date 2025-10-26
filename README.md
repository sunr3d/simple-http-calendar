# simple-http-calendar

HTTP-сервис для управления календарём событий с поддержкой напоминаний и автоматической архивации.

## Архитектура

- **Clean Architecture** с разделением на слои (domain, services, infrastructure, handlers)
- **In-Memory хранилище** для быстрого доступа к данным
- **Асинхронный логгер** с неблокирующей записью
- **Background сервисы** для напоминаний и архивации
- **Graceful shutdown** с корректным завершением всех горутин

## Функциональность

- ✅ **CRUD операции** для событий календаря
- ✅ **Выборка событий** за день/неделю/месяц
- ✅ **ReminderService** - автоматические напоминания о событиях
- ✅ **ArchiveService** - автоматическая архивация старых событий
- ✅ **AsyncLogger** - асинхронное логирование для высокой производительности
- ✅ **Graceful shutdown** - корректное завершение всех сервисов
- ✅ **Race-free** - проверено race detector'ом
- ✅ **No goroutine leaks** - проверено goleak

## Установка и запуск

### 1. Быстрый старт

```bash
# Установка зависимостей
go mod tidy

# Запуск приложения
go run cmd/main.go

# Или через Docker
make up && make logs
```

### 2. Запуск с race detector

```bash
# Проверка race conditions
go run -race cmd/main.go
make test-smoke
```

## Конфигурация

Настройки через переменные окружения или `.env` файл:

```bash
# HTTP сервер
HTTP_PORT=8080
HTTP_TIMEOUT=20s

# Логгер
LOG_LEVEL=info
LOG_CHAN_SIZE=100

# ReminderService
REMINDER_CHAN_SIZE=100
REMINDER_INTERVAL=2s

# ArchiveService
ARCHIVE_INTERVAL=10s
```

## API Endpoints

### Создание события

```bash
POST /create_event
Content-Type: application/json

{
  "user_id": 1,
  "date": "2025-10-27T14:30:00",
  "event": "Созвон с командой",
  "reminder": true
}

# Ответ: {"result": "event-uuid"}
```

### Обновление события

```bash
POST /update_event
Content-Type: application/json

{
  "event_id": "event-uuid",
  "user_id": 1,
  "date": "2025-10-27T15:00:00",
  "event": "Созвон с командой (перенос)",
  "reminder": false
}

# Ответ: {"result": "ok"}
```

### Удаление события

```bash
POST /delete_event
Content-Type: application/json

{
  "event_id": "event-uuid"
}

# Ответ: {"result": "ok"}
```

### Получение событий

```bash
# За день
GET /events_for_day?user_id=1&date=2025-10-27

# За неделю
GET /events_for_week?user_id=1&date=2025-10-27

# За месяц
GET /events_for_month?user_id=1&date=2025-10-27

# Ответ: {"result": [...events]}
```

### HTTP коды ответов

- `200` — успех
- `400` — ошибка валидации
- `503` — ошибка бизнес-логики
- `500` — внутренняя ошибка сервера

## Примеры использования

### Базовые команды

```bash
# Создание события
curl -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "date": "2025-10-27T14:30:00", "event": "Событие", "reminder": true}'

# Обновление события
curl -X POST http://localhost:8080/update_event \
  -H "Content-Type: application/json" \
  -d '{"event_id": "uuid", "user_id": 1, "date": "2025-10-27T15:00:00", "event": "Обновленное событие", "reminder": false}'

# Удаление события
curl -X POST http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{"event_id": "uuid"}'

# Получение событий за день
curl "http://localhost:8080/events_for_day?user_id=1&date=2025-10-27"

# Получение событий за неделю
curl "http://localhost:8080/events_for_week?user_id=1&date=2025-10-27"

# Получение событий за месяц
curl "http://localhost:8080/events_for_month?user_id=1&date=2025-10-27"
```

## Структура проекта

```
├── cmd/
│   └── main.go              # Точка входа приложения
├── internal/
│   ├── config/              # Конфигурация
│   ├── logger/              # Асинхронный логгер
│   ├── server/              # HTTP сервер
│   ├── middleware/          # HTTP middleware
│   ├── handlers/http/       # HTTP обработчики
│   ├── handlers/validators/ # Валидация запросов
│   ├── services/            # Бизнес-логика
│   │   ├── calendarsvc/     # Сервис календаря
│   │   ├── remindersvc/     # Сервис напоминаний
│   │   └── archiversvc/     # Сервис архивации
│   ├── infra/               # Инфраструктура
│   │   ├── inmemdb/         # In-memory БД
│   │   └── inmembroker/     # In-memory брокер
│   ├── interfaces/          # Интерфейсы слоев
│   ├── httpx/               # HTTP утилиты
│   └── entrypoint/          # Сборка зависимостей
├── models/                  # Доменные модели
├── smoke.sh                 # Smoke тесты
├── Dockerfile               # Docker образ
├── docker-compose.yml       # Docker Compose
└── Makefile                 # Команды сборки
```

### Команды разработки

```bash
# Запуск тестов
make test

# Smoke тесты
make test-smoke

# Линтер
make lint

# Форматирование кода
make fmt

# Сборка бинарника
make build

# Docker
make up      # Запуск в Docker
make down    # Остановка
make logs    # Логи
```

### Тестирование

```bash
# Все тесты
go test -v ./...

# Тесты с race detector
go test -race -v ./...

# Тесты с проверкой утечек горутин
GOROUTINE_LEAK_DETECTION=1 go test -v ./...

# Smoke тесты (требуется запущенный сервер)
./smoke.sh
```

### Мониторинг

```bash
# Статус контейнеров
docker compose ps

# Логи приложения
docker compose logs -f app

# Перезапуск
docker compose restart
```

## Производительность

- **Асинхронный логгер**: Неблокирующая запись логов с fallback механизмом
- **Background сервисы**: Параллельная обработка напоминаний и архивации
- **In-Memory хранилище**: Быстрый доступ к данным без задержек БД
- **Graceful shutdown**: Корректное завершение всех горутин без потери данных
- **Race-free**: Проверено race detector'ом, никаких data races
- **No leaks**: Проверено goleak, никаких утечек горутин
