# syntax=docker/dockerfile:1

FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /telegram-printer-scanner

CMD ["/telegram-printer-scanner"]
