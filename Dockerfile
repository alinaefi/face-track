FROM golang:1.21.8-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy && go build -o face-track main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/face-track .
COPY --from=build /app/internal/database/migrations /app/internal/database/migrations

RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.1/migrate.linux-amd64.tar.gz | tar xz -C /usr/local/bin

COPY entrypoint.sh /app/
RUN chmod +x /app/entrypoint.sh

EXPOSE 4221

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./face-track"]


