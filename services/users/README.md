# Users Service

Простой сервис хранения пользователей и их URL в PostgreSQL.
Встроен scheduler, который публикует `ParseRequested` в Kafka по интервалам URL.

## Запуск

- Docker: `docker compose up -d --build`
- Локально (Windows PowerShell): `$env:configPath = ".\\config.yaml"` и `go run .\\cmd\\app`

Health: `GET http://localhost:8071/health`
gRPC: `:50061`
