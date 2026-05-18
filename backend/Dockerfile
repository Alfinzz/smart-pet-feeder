# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.21

WORKDIR /app
RUN apk add --no-cache ca-certificates wget \
    && mkdir -p /app/uploads/pets

COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY migrations /app/migrations
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENV APP_PORT=8080
ENV GIN_MODE=release
ENV MIGRATIONS_PATH=/app/migrations
ENV UPLOAD_DIR=/app/uploads

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/api"]
