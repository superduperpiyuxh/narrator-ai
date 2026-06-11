# ── Build stage ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /app/narrator-ai .

# ── Runtime stage ───────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/narrator-ai .

RUN mkdir -p /app/data/db

EXPOSE 8080

ENV PORT=8080
ENV DATABASE_PATH=/app/data/db/narratorai.db
ENV DATA_DIR=/app/data/sample_json_20260301

CMD ["./narrator-ai"]
