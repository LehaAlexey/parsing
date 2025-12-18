# Parsing Service

Сервис читает `ParseRequested` из Kafka, скачивает страницу товара, извлекает цену и публикует `PriceMeasured` в Kafka.

## Kafka события (JSON)

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

Health: `GET http://localhost:8070/health`

