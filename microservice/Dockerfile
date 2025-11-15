# 1 этап: Сборка микросервиса
FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o review-manager ./cmd

# 2 этап: запуск проекта на alpine
FROM alpine:3.20
WORKDIR /app

COPY --from=build /app/review-manager /review-manager

EXPOSE 8080

ENTRYPOINT ["/review-manager"]