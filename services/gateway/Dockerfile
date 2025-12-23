FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gateway-service ./cmd/app

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin app
USER app
WORKDIR /app

COPY --from=build /out/gateway-service /app/gateway-service
COPY config.docker.yaml /app/config.yaml

EXPOSE 8080
ENTRYPOINT ["/app/gateway-service"]

