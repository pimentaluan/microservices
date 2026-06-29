FROM golang:1.26 AS builder

WORKDIR /app

COPY microservices/order ./microservices/order
COPY microservices-proto ./microservices-proto

WORKDIR /app/microservices/order

RUN go mod download
RUN go build -o /order-service ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /order-service /usr/local/bin/order-service

EXPOSE 3000

CMD ["order-service"]
