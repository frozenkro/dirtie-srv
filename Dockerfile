#syntax=docker/dockerfile:1

FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./src/* ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /dirtie-srv

EXPOSE 8000

CMD ["/dirtie-srv"]
