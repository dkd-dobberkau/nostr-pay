FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o nostr-pay ./cmd/server/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /app/nostr-pay .
RUN mkdir -p /app/data && chown -R appuser:appuser /app
USER appuser
HEALTHCHECK --interval=30s --timeout=10s CMD curl -f http://localhost:8080/api/health || exit 1
EXPOSE 8080
CMD ["./nostr-pay"]
