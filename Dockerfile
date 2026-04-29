FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/devpulse-api ./cmd/api

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app
USER app

WORKDIR /home/app
COPY --from=builder /bin/devpulse-api /usr/local/bin/devpulse-api

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/devpulse-api"]

