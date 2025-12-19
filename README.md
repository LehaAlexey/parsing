# Parsing Service

Сервис читает события `ParseRequested` из Kafka, скачивает страницу товара, извлекает цену и публикует `PriceMeasured` обратно в Kafka.

## Формат событий (JSON)

ParseRequested:
```json
{"event_id":"...","occurred_at":"2025-01-01T12:00:00Z","correlation_id":"...","product_id":"1","url":"https://...","scheduled_at":"2025-01-01T12:00:00Z","priority":0}
```

PriceMeasured:
```json
{"event_id":"...","occurred_at":"2025-01-01T12:00:00Z","correlation_id":"...","product_id":"1","price":990,"currency":"RUB","parsed_at":"2025-01-01T12:00:00Z","source_url":"https://...","meta_hash":"..."}
```

## Запуск

- Docker: `docker compose up -d --build`
- Локально (Windows PowerShell): `$env:configPath = ".\\config.yaml"` и `go run .\\cmd\\app`

Health check: `GET http://localhost:8070/health`

## Ограничения

Парсер не идеален: страницы с авторизацией, капчей, нестандартным HTML или JS-рендером могут требовать доп. заголовки, куки или отдельные правила. Проверяйте результат и при необходимости дополняйте обработчик.***
