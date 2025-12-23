# History Service

Сервис читает `PriceMeasured` из Kafka, пишет в History DB и обновляет Redis (последние измерения).

gRPC: `:50062`
Health: `http://localhost:8072/health`

