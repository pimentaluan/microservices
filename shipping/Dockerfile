FROM golang:1.26 AS builder

WORKDIR /app

COPY microservices/shipping ./microservices/shipping
COPY microservices-proto ./microservices-proto

WORKDIR /app/microservices/shipping

RUN go mod download
RUN go build -o /shipping-service ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /shipping-service /usr/local/bin/shipping-service

EXPOSE 3002

CMD ["shipping-service"]
