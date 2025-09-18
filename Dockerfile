FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/server ./cmd/server
COPY ./internal ./internal
COPY ./migrations ./migrations

RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/server/main.go -o cmd/docs

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -o main ./cmd/server/

FROM alpine:latest AS final

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/cmd/docs ./cmd/docs
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]