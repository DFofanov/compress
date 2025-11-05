# Dockerfile для Compress
# Multi-stage сборка для оптимизации размера образа

# Стадия 1: Сборка приложения
FROM golang:1.24-alpine AS builder

# Метаданные
LABEL maintainer="Compress Team"
LABEL description="Compress - автоматическое сжатие PDF и изображений"
LABEL version="1.0.0"

# Установка необходимых пакетов для сборки
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev

# Создание пользователя для сборки
RUN adduser -D -s /bin/sh -u 1001 builder

# Установка рабочей директории
WORKDIR /app

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей (кэшируемый слой)
RUN go mod download && go mod verify

# Копирование исходного кода
COPY . .

# Изменение владельца файлов
RUN chown -R builder:builder /app
USER builder

# Сборка приложения с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o compress cmd/main.go

# Стадия 2: Минимальный runtime образ
FROM alpine:3.19

# Установка runtime зависимостей
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Создание пользователя для runtime
RUN addgroup -g 1001 -S pdfuser && \
    adduser -u 1001 -S pdfuser -G pdfuser

# Создание рабочих директорий
RUN mkdir -p /app/input /app/output /app/config /app/logs && \
    chown -R pdfuser:pdfuser /app

# Копирование скомпилированного приложения
COPY --from=builder /app/compress /usr/local/bin/compress

# Копирование конфигурационного файла по умолчанию
COPY config.yaml /app/config/config.yaml

# Установка переменных окружения
ENV APP_CONFIG_PATH="/app/config/config.yaml"
ENV APP_LOG_LEVEL="info"
ENV APP_INPUT_DIR="/app/input"
ENV APP_OUTPUT_DIR="/app/output"
ENV APP_LOGS_DIR="/app/logs"

# Переключение на непривилегированного пользователя
USER pdfuser

# Установка рабочей директории
WORKDIR /app

# Открытие портов (если потребуется web интерфейс в будущем)
# EXPOSE 8080

# Volumes для данных
VOLUME ["/app/input", "/app/output", "/app/config", "/app/logs"]

# Healthcheck для мониторинга
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD compress --version || exit 1

# Команда запуска
ENTRYPOINT ["compress"]
CMD ["--config", "/app/config/config.yaml"]

# Метаданные образа
LABEL org.opencontainers.image.title="Compress"
LABEL org.opencontainers.image.description="Автоматическое сжатие PDF файлов с TUI интерфейсом"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.created="2024"
LABEL org.opencontainers.image.vendor="Compress Team"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.documentation="https://github.com/your-username/compress"
LABEL org.opencontainers.image.source="https://github.com/your-username/compress"
