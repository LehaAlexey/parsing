# Services Workspace

This repository contains multiple services in one workspace:

- `services/parsing` - Parsing Service (Kafka consumer/producer)
- `services/users` - Users Service (PostgreSQL + scheduler + gRPC)
- `services/history` - History Service (PostgreSQL + Redis + gRPC)
- `services/gateway` - Gateway Service (HTTP BFF -> gRPC users/history)

## Quick start

`docker compose up -d --build`
