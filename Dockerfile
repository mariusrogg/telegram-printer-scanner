# syntax=docker/dockerfile:1

FROM golang:1.24
ARG USERNAME=nonroot
ARG USER_UID=1000
ARG USER_GID=$USER_UID

WORKDIR /app

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && chown -R $USER_UID:$USER_GID /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /telegram-printer-scanner

CMD ["/telegram-printer-scanner"]

RUN chown -R $USER_UID:$USER_GID /telegram-printer-scanner
RUN mkdir /nonexistent
RUN chown -R $USER_UID:$USER_GID /nonexistent

USER $USERNAME