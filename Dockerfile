# yanxo Telegram bot — production image for Render (and similar)
# Set env in Render: BOT_TOKEN, CHANNEL_ID, TURSO_DATABASE_URL, TURSO_AUTH_TOKEN
# Optional: CHANNEL_URL, CHANNEL_USERNAME

FROM golang:1.24-bookworm AS builder

WORKDIR /src

# Honor go.mod toolchain (e.g. 1.25) when building
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary: no CGO needed for this bot (Turso/libSQL over HTTPS)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/yanxo-bot \
    ./cmd/bot

# --- runtime ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 -S app \
    && adduser -u 10001 -S -G app -h /app app

WORKDIR /app

# HTTP health: / , /health , /healthz → {"status":"ok"} (override with PORT / HEALTH_ADDR)
EXPOSE 8080
ENV PORT=8080

COPY --from=builder /out/yanxo-bot /app/yanxo-bot
COPY migrations /app/migrations

USER app:app

# Migrations path is ./migrations relative to CWD (see internal/bot/app.go)
ENV TZ=Asia/Tashkent

ENTRYPOINT ["/app/yanxo-bot"]
