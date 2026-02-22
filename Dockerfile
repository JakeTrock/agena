FROM golang:1.21-alpine

# Install bash as the code relies on it
RUN apk add --no-cache bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o bin/agena ./src

ENTRYPOINT ["bin/agena"]
