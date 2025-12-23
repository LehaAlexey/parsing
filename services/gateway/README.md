# Gateway Service

HTTP BFF, который проксирует запросы к users и history через gRPC.

HTTP: `:8080`
Health: `GET http://localhost:8080/health`

Endpoints:
- `POST /users`
- `GET /users/{id}`
- `POST /users/{id}/urls`
- `GET /users/{id}/urls`
- `GET /history?product_id=...&from=...&to=...&limit=...` (from/to — unix seconds)

