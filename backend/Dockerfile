FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY backend/ .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY backend/migrations /app/migrations

USER app

EXPOSE 8080

ENTRYPOINT ["/app/server"]