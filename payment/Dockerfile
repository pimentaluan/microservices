FROM golang:1.26 AS builder

WORKDIR /app

COPY microservices/payment ./microservices/payment
COPY microservices-proto ./microservices-proto

WORKDIR /app/microservices/payment

RUN go mod download
RUN go build -o /payment-service ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /payment-service /usr/local/bin/payment-service

EXPOSE 3001

CMD ["payment-service"]
