FROM golang:1.25-alpine AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/api /app/api
RUN mkdir -p /app/snapshots && chown -R app:app /app/snapshots

USER app
EXPOSE 8080

ENTRYPOINT ["/app/api"]
