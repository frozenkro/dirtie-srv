#syntax=docker/dockerfile:1

FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/dirtie-srv/main.go

EXPOSE 8000

CMD ["/app/main"]
