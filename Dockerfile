FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ai-japanese-learning ./cmd/server

FROM alpine:3.21

WORKDIR /app
COPY --from=build /out/ai-japanese-learning /app/ai-japanese-learning
COPY internal/web/assets /app/internal/web/assets

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/health >/dev/null || exit 1

CMD ["/app/ai-japanese-learning"]
