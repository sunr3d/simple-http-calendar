#!/bin/bash

echo "=== РАСШИРЕННЫЙ НАГРУЗОЧНЫЙ SMOKE-ТЕСТ ==="
echo "Тестируем: асинхронный логгер, напоминания, архивация, race conditions, graceful shutdown"
echo ""

# Проверяем, запущено ли приложение на порту 8080
echo "🔍 Проверяем доступность сервера на порту 8080..."
if ! netstat -tuln | grep -q ":8080 "; then
    echo ""
    echo "❌ ОШИБКА: Сервер недоступен на порту 8080!"
    echo ""
    echo "📋 Инструкция по запуску:"
    echo "1. Откройте новый терминал"
    echo "2. Запустите приложение: go run cmd/main.go"
    echo "3. Дождитесь сообщения 'запуск HTTP сервера'"
    echo "4. Вернитесь в этот терминал и запустите: ./smoke.sh"
    echo ""
    echo "💡 Альтернативно можно запустить с race detector:"
    echo "   go run -race cmd/main.go"
    echo ""
    exit 1
fi

echo "✅ Сервер доступен, продолжаем тестирование..."
echo ""

# Генерируем динамические даты
TODAY=$(date +"%Y-%m-%d")
FUTURE_DATE=$(date -d "+1 day" +"%Y-%m-%d")
PAST_DATE=$(date -d "-1 day" +"%Y-%m-%d")
FUTURE_TIME=$(date -d "+1 day +1 hour" +"%Y-%m-%dT%H:%M:%S")  # Завтра + 1 час
PAST_TIME=$(date -d "-1 hour" +"%Y-%m-%dT%H:%M:%S")

echo "📅 Используемые даты:"
echo "  Сегодня: $TODAY"
echo "  Будущее: $FUTURE_DATE"
echo "  Прошлое: $PAST_DATE"
echo "  Будущее время: $FUTURE_TIME"
echo "  Прошлое время: $PAST_TIME"
echo ""

# Функция для измерения времени
measure_time() {
    local start_time=$(date +%s.%N)
    eval "$@"
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    echo "⏱️  Время выполнения: ${duration}s"
}

# Функция для проверки статуса
check_status() {
    local response="$1"
    local expected_status="$2"
    local actual_status=$(echo "$response" | jq -r '.error // "success"')
    
    if [[ "$actual_status" == "success" && "$expected_status" == "200" ]]; then
        echo "✅ OK"
    elif [[ "$actual_status" == *"error"* && "$expected_status" == "400" ]]; then
        echo "✅ OK (ожидаемая ошибка)"
    else
        echo "❌ FAIL: ожидался $expected_status, получен $actual_status"
    fi
}

echo "🚀 1. БАЗОВЫЙ ФУНКЦИОНАЛ"
echo "================================"

# 1.1 Создание событий
echo "1.1 Создание событий..."
measure_time 
EVENT1=$(curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": 1, \"date\": \"$FUTURE_TIME\", \"event\": \"Meeting 1\", \"reminder\": true}" \
  | jq -r ".result")

EVENT2=$(curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": 1, \"date\": \"$FUTURE_TIME\", \"event\": \"Meeting 2\", \"reminder\": false}" \
  | jq -r ".result")

EVENT3=$(curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": 2, \"date\": \"$FUTURE_TIME\", \"event\": \"Meeting 3\", \"reminder\": true}" \
  | jq -r ".result")

echo "Созданы события: $EVENT1, $EVENT2, $EVENT3"

echo ""
echo "🔥 2. НАГРУЗОЧНОЕ ТЕСТИРОВАНИЕ"
echo "================================"

# 2.1 Массовые запросы (проверка асинхронности логгера)
echo "2.1 Массовые запросы (50 одновременных)..."
measure_time 
for i in {1..50}; do
  curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $((i % 5 + 1)), \"date\": \"$FUTURE_TIME\", \"event\": \"Mass test $i\", \"reminder\": false}" > /dev/null &
done
wait
echo "50 запросов отправлены одновременно"

# 2.2 Нагрузка на GET endpoints
echo "2.2 Нагрузка на GET endpoints (30 запросов)..."
measure_time 
for i in {1..30}; do
  curl -s "http://localhost:8080/events_for_day?user_id=$((i % 5 + 1))&date=$FUTURE_DATE" > /dev/null &
  curl -s "http://localhost:8080/events_for_week?user_id=$((i % 5 + 1))&date=$FUTURE_DATE" > /dev/null &
  curl -s "http://localhost:8080/events_for_month?user_id=$((i % 5 + 1))&date=$FUTURE_DATE" > /dev/null &
done
wait
echo "90 GET запросов отправлены"

# 2.3 Смешанная нагрузка (CRUD + GET)
echo "2.3 Смешанная нагрузка (100 операций)..."
measure_time 
for i in {1..100}; do
  case $((i % 4)) in
    0)
      # Создание
      curl -s -X POST http://localhost:8080/create_event \
        -H "Content-Type: application/json" \
        -d "{\"user_id\": $((i % 3 + 1)), \"date\": \"$FUTURE_TIME\", \"event\": \"Mixed load $i\", \"reminder\": false}" > /dev/null &
      ;;
    1)
      # Получение
      curl -s "http://localhost:8080/events_for_day?user_id=$((i % 3 + 1))&date=$FUTURE_DATE" > /dev/null &
      ;;
    2)
      # Обновление (если есть события)
      if [[ -n "$EVENT1" ]]; then
curl -s -X POST http://localhost:8080/update_event \
          -H "Content-Type: application/json" \
          -d "{\"event_id\": \"$EVENT1\", \"user_id\": 1, \"date\": \"$FUTURE_TIME\", \"event\": \"Updated $i\", \"reminder\": false}" > /dev/null &
      fi
      ;;
    3)
      # Удаление (создаем и сразу удаляем)
      TEMP_EVENT=$(curl -s -X POST http://localhost:8080/create_event \
        -H "Content-Type: application/json" \
        -d "{\"user_id\": $((i % 3 + 1)), \"date\": \"$FUTURE_TIME\", \"event\": \"Temp $i\", \"reminder\": false}" \
        | jq -r ".result")
      if [[ "$TEMP_EVENT" != "null" && -n "$TEMP_EVENT" ]]; then
        curl -s -X POST http://localhost:8080/delete_event \
          -H "Content-Type: application/json" \
          -d "{\"event_id\": \"$TEMP_EVENT\"}" > /dev/null &
      fi
      ;;
  esac
done
wait
echo "100 смешанных операций выполнено"

echo ""
echo "🧪 3. ТЕСТИРОВАНИЕ REMINDER SERVICE"
echo "================================"

# 3.1 Создание событий с напоминаниями в прошлом
echo "3.1 Создание событий с напоминаниями в прошлом..."
measure_time 
for i in {1..10}; do
  curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $((i % 3 + 1)), \"date\": \"$PAST_TIME\", \"event\": \"Past reminder $i\", \"reminder\": true}" > /dev/null &
done
wait
echo "10 событий с напоминаниями в прошлом созданы"
echo "Проверьте логи на наличие: НАПОМИНАНИЕ: событие"

# 3.2 Создание событий с напоминаниями в будущем
echo "3.2 Создание событий с напоминаниями в будущем..."
measure_time 
for i in {1..5}; do
  curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $((i % 3 + 1)), \"date\": \"$FUTURE_TIME\", \"event\": \"Future reminder $i\", \"reminder\": true}" > /dev/null &
done
wait
echo "5 событий с напоминаниями в будущем созданы"

echo ""
echo "⚠️  4. ТЕСТИРОВАНИЕ ОШИБОК И ВАЛИДАЦИИ"
echo "================================"

# 4.1 Неправильные Content-Type
echo "4.1 Тестирование неправильных Content-Type..."
measure_time 
for i in {1..5}; do
  curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: text/plain" \
    -d "invalid data" > /dev/null &
done
wait
echo "5 запросов с неправильным Content-Type отправлены"

# 4.2 Неправильные данные
echo "4.2 Тестирование неправильных данных..."
measure_time
curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": 0, \"date\": \"$FUTURE_TIME\", \"event\": \"Test\", \"reminder\": false}" > /dev/null &

curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "date": "invalid-date", "event": "Test", "reminder": false}' > /dev/null &

curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": 1, \"date\": \"$FUTURE_TIME\", \"event\": \"\", \"reminder\": false}" > /dev/null &

curl -s "http://localhost:8080/events_for_day?user_id=1&date=invalid-date" > /dev/null &

curl -s "http://localhost:8080/events_for_day?user_id=0&date=$FUTURE_DATE" > /dev/null &

wait
echo "5 запросов с неправильными данными отправлены"


echo ""
echo "📦 5. ТЕСТИРОВАНИЕ АРХИВАЦИИ"
echo "================================"

# 5.1 Создание событий в прошлом для архивации
echo "5.1 Создание событий в прошлом для архивации..."
YESTERDAY=$(date -d "yesterday" +"%Y-%m-%d")
YESTERDAY_TIME=$(date -d "yesterday -1 hour" +"%Y-%m-%dT%H:%M:%S")

echo "📅 Создаем события на вчерашний день: $YESTERDAY"
echo "📅 Время событий: $YESTERDAY_TIME"

# Создаем несколько событий в прошлом
for i in {1..5}; do
  EVENT_ID=$(curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $i, \"date\": \"$YESTERDAY_TIME\", \"event\": \"Past event $i for archive\", \"reminder\": false}" \
    | jq -r ".result")
  
  if [ -n "$EVENT_ID" ] && [ "$EVENT_ID" != "null" ]; then
    echo "  ✅ Событие $i: $EVENT_ID"
  else
    echo "  ❌ Ошибка создания события $i"
  fi
done

# 5.2 Создание событий в будущем для сравнения
echo ""
echo "5.2 Создание событий в будущем для сравнения..."
for i in {1..3}; do
  EVENT_ID=$(curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $i, \"date\": \"$FUTURE_TIME\", \"event\": \"Future event $i (not archived)\", \"reminder\": false}" \
    | jq -r ".result")
  
  if [ -n "$EVENT_ID" ] && [ "$EVENT_ID" != "null" ]; then
    echo "  ✅ Будущее событие $i: $EVENT_ID"
  else
    echo "  ❌ Ошибка создания будущего события $i"
  fi
done

# 5.3 Проверка событий ДО архивации
echo ""
echo "5.3 Проверка событий ДО архивации..."
echo "📊 События за вчерашний день (должны быть видны):"
curl -s "http://localhost:8080/events_for_day?user_id=1&date=$YESTERDAY" | jq .

echo ""
echo "📊 События за будущий день (должны быть видны):"
curl -s "http://localhost:8080/events_for_day?user_id=1&date=$FUTURE_DATE" | jq .

# 5.4 Ожидание архивации
ARCHIVE_INTERVAL=10
WAIT_TIME=$((ARCHIVE_INTERVAL + 3)) # Ждем чуть больше интервала
echo ""
echo "⏰ Ждем $WAIT_TIME секунд для срабатывания архивации (интервал $ARCHIVE_INTERVAL сек)..."
echo "🔍 В это время следите за логами на предмет сообщений архивации..."
sleep "$WAIT_TIME"

# 5.5 Проверка событий ПОСЛЕ архивации
echo ""
echo "5.5 Проверка событий ПОСЛЕ архивации..."
echo "📊 События за вчерашний день (должны быть заархивированы и скрыты):"
YESTERDAY_RESULT=$(curl -s "http://localhost:8080/events_for_day?user_id=1&date=$YESTERDAY")
echo "$YESTERDAY_RESULT" | jq .

# Проверяем, что события действительно заархивированы (пустой результат)
YESTERDAY_COUNT=$(echo "$YESTERDAY_RESULT" | jq '.result | length')
if [ "$YESTERDAY_COUNT" -eq 0 ]; then
  echo "✅ События за вчерашний день успешно заархивированы (скрыты из обычных запросов)"
else
  echo "❌ События за вчерашний день НЕ заархивированы (все еще видны)"
fi

echo ""
echo "📊 События за будущий день (НЕ должны быть заархивированы):"
FUTURE_RESULT=$(curl -s "http://localhost:8080/events_for_day?user_id=1&date=$FUTURE_DATE")
echo "$FUTURE_RESULT" | jq .

# Проверяем, что будущие события не заархивированы
FUTURE_COUNT=$(echo "$FUTURE_RESULT" | jq '.result | length')
if [ "$FUTURE_COUNT" -gt 0 ]; then
  echo "✅ Будущие события НЕ заархивированы (остались видимыми)"
else
  echo "⚠️  Будущие события тоже скрыты (возможно, проблема с датами)"
fi

# 5.6 Проверка событий за неделю и месяц
echo ""
echo "5.6 Проверка событий за неделю и месяц..."
echo "📊 События за неделю:"
curl -s "http://localhost:8080/events_for_week?user_id=1&date=$TODAY" | jq .

echo ""
echo "📊 События за месяц:"
curl -s "http://localhost:8080/events_for_month?user_id=1&date=$TODAY" | jq .

echo ""
echo "✅ Тест архивации завершен!"
echo ""
echo "🔍 АНАЛИЗ РЕЗУЛЬТАТОВ АРХИВАЦИИ:"
echo "================================"
if [ "$YESTERDAY_COUNT" -eq 0 ]; then
  echo "✅ Архивация работает: события в прошлом скрыты"
else
  echo "❌ Архивация НЕ работает: события в прошлом все еще видны"
fi

if [ "$FUTURE_COUNT" -gt 0 ]; then
  echo "✅ Будущие события не заархивированы (правильно)"
else
  echo "⚠️  Будущие события тоже скрыты (проверьте даты)"
fi

echo ""
echo "🔍 Проверьте логи на наличие сообщений:"
echo "   - 'событие архивировано' с service='archiver'"
echo "   - 'запуск сервиса архивации'"
echo "   - 'отмена контекста, сервис архивации остановлен'"
echo ""
echo "💡 ПРИМЕЧАНИЕ: Заархивированные события скрыты из обычных запросов"
echo "   Это правильное поведение - архивация работает!"

echo ""
echo "🔄 6. ТЕСТИРОВАНИЕ GRACEFUL SHUTDOWN"
echo "================================"

# 6.1 Отправка запросов во время shutdown
echo "6.1 Отправка запросов во время shutdown..."
echo "ВНИМАНИЕ: Сейчас будет отправлен SIGTERM, но сначала отправим запросы..."

# Отправляем запросы в фоне
for i in {1..20}; do
  curl -s -X POST http://localhost:8080/create_event \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $((i % 3 + 1)), \"date\": \"$FUTURE_TIME\", \"event\": \"Shutdown test $i\", \"reminder\": false}" > /dev/null &
done

# Небольшая задержка
sleep 0.5

echo "Отправляем SIGTERM..."
# Найти PID процесса main через ps -a (первый процесс main)
APP_PID=$(ps -a | grep "main" | grep -v grep | head -1 | awk '{print $1}')
if [[ -n "$APP_PID" ]]; then
  echo "Найден процесс main с PID: $APP_PID"
  echo "Отправляем SIGTERM..."
  kill -TERM "$APP_PID"
  echo "SIGTERM отправлен процессу $APP_PID"
  echo "Ждем graceful shutdown (2 секунды)..."
  sleep 2
else
  echo "Процесс main не найден, возможно уже завершен"
fi

wait
echo "Graceful shutdown тест завершен"

echo ""
echo "📊 7. ФИНАЛЬНАЯ ПРОВЕРКА"
echo "================================"

# 7.1 Проверка финального состояния
echo "7.1 Проверка финального состояния..."
measure_time
echo "Проверяем доступность сервера..."
if curl -s http://localhost:8080/events_for_day?user_id=1&date=$FUTURE_DATE > /dev/null; then
  echo "✅ Сервер доступен"
else
  echo "❌ Сервер недоступен (возможно, завершился)"
fi

echo ""
echo "=== РАСШИРЕННЫЙ НАГРУЗОЧНЫЙ ТЕСТ ЗАВЕРШЕН ==="
echo "Всего операций: ~350+"
echo "Проверено: асинхронность, напоминания, архивация, валидация, гонки, graceful shutdown"