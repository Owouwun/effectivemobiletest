FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g ./cmd/server/main.go -o cmd/docs

RUN go build -o main ./cmd/server/

EXPOSE 8080

CMD ["./main"]