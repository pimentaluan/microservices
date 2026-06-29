# Microsservicos com Go, gRPC e Arquitetura Hexagonal

Projeto da disciplina de Programacao Distribuida com tres microsservicos:

- `order`: recebe pedidos de compra
- `payment`: registra a cobranca do pedido
- `shipping`: calcula e registra o prazo de entrega

O projeto usa:

- Go
- gRPC
- Protocol Buffers
- MySQL
- GORM
- Arquitetura Hexagonal

## Estrutura

```txt
atividade2/
├── microservices/
│   ├── init.sql
│   ├── docker-compose.yml
│   ├── order/
│   ├── payment/
│   └── shipping/
└── microservices-proto/
    ├── order/
    ├── payment/
    ├── shipping/
    └── run.sh
```

## O que foi implementado

### Parte 1 - Microsservico Order com gRPC

- definicao do `order.proto`
- geracao dos stubs Go
- implementacao do microsservico `order`
- persistencia dos pedidos no MySQL
- exposicao do servico `order.Order/Create`

### Parte 2 - Comunicacao entre servicos com gRPC

- definicao do `payment.proto`
- geracao dos stubs Go de `payment`
- criacao do microsservico `payment`
- integracao do `order` com o `payment`
- calculo do valor total do pedido antes da cobranca

### Parte 3 - Tratamento de erros

- `payment` retorna `InvalidArgument` quando o valor total passa de `1000`
- `order` retorna `InvalidArgument` quando o pedido passa de `50` itens no total
- erros internos sao convertidos para `Internal`
- status do pedido no banco e atualizado para:
  - `Canceled` quando ocorre erro
  - `Paid` quando a cobranca e concluida

### Parte 4 - Comunicacao resiliente

- timeout de `2s` por chamada do `order` para o `payment`
- log especifico quando ocorre `DeadlineExceeded`
- retry automatico no cliente gRPC do `payment`
- retry configurado para `Unavailable` e `ResourceExhausted`
- maximo de `5` tentativas com `BackoffLinear(time.Second)`

### Parte 5 - Shipping, estoque e deploy com Docker

- criacao do microsservico `shipping` com protobuf e arquitetura hexagonal
- integracao do `order` com `shipping` apenas apos pagamento bem-sucedido
- calculo do prazo de entrega com base na quantidade total de itens
- validacao de estoque no `order` antes de salvar o pedido
- retorno de erro `NotFound` para produtos inexistentes
- Dockerfiles para `order`, `payment` e `shipping`
- `docker-compose.yml` para subir todo o ambiente

## Como gerar os stubs protobuf

Dentro de `microservices-proto`:

```bash
cd microservices-proto
sh run.sh
```

## Execucao

### 1. MySQL

```bash
cd microservices

docker run --name microsservices-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=minhasenha \
  -v "$(pwd)/init.sql:/docker-entrypoint-initdb.d/init.sql" \
  -d mysql
```

Se o container ja existir:

```bash
docker start microsservices-mysql
```

### 2. Payment

```bash
cd microservices/payment

DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/payment?parseTime=true" \
APPLICATION_PORT=3001 \
ENV=development \
go run cmd/main.go
```

### 3. Shipping

```bash
cd microservices/shipping

DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/shipping?parseTime=true" \
APPLICATION_PORT=3002 \
ENV=development \
go run cmd/main.go
```

### 4. Order

```bash
cd microservices/order

DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order?parseTime=true" \
PAYMENT_SERVICE_URL="127.0.0.1:3001" \
SHIPPING_SERVICE_URL="127.0.0.1:3002" \
APPLICATION_PORT=3000 \
ENV=development \
go run cmd/main.go
```

### 5. Verificacao do servico

```bash
grpcurl -plaintext 127.0.0.1:3000 list
```

Saida esperada: `order.Order`.

## Testes

### Parte 1 - Teste basico do Order

Comando:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Retorno esperado:

```json
{
  "orderId": 1
}
```

### Parte 2 - Teste da comunicacao Order -> Payment

Com os tres servicos em execucao, rode a mesma requisicao da Parte 1:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Se houver retorno de `orderId`, a chamada ao `payment` foi concluida sem erro.

### Parte 3 - Testes de tratamento de erros

#### 1. Erro por mais de 50 itens

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":51,"unit_price":1}]}' \
  127.0.0.1:3000 order.Order/Create
```

Retorno esperado:

```txt
ERROR:
  Code: InvalidArgument
  Message: orders over 50 items are not allowed
```

#### 2. Erro por pagamento acima de 1000

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":10,"unit_price":150}]}' \
  127.0.0.1:3000 order.Order/Create
```

Retorno esperado:

```txt
ERROR:
  Code: InvalidArgument
  Message: payment over 1000 is not allowed
```

#### 3. Caso de sucesso

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Retorno esperado:

```json
{
  "orderId": 3
}
```

#### 4. Verificar o status salvo no banco

```bash
docker exec -it microsservices-mysql mysql -uroot -pminhasenha
```

Consulta:

```sql
USE `order`;
SELECT id, customer_id, status FROM orders;
```

Verificacao:

- pedidos com erro ficam como `Canceled`
- pedidos com sucesso ficam como `Paid`

### Parte 4 - Testes de timeout e retry

#### 1. Retry

Com o `order` rodando, pare o `payment` com `Ctrl+C` e execute:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Verificacao:

- a chamada falha no final
- antes de falhar, o cliente do `order` tenta novamente automaticamente

#### 2. Timeout

Para simular lentidao no `payment`, adicione temporariamente este trecho no inicio da funcao `Create` de [payment/internal/adapters/grpc/server.go](/Users/luanpimenta/Documents/projetos/ifpb/pd/atividade2/microservices/payment/internal/adapters/grpc/server.go:29):

```go
time.Sleep(3 * time.Second)
```

Importe `time`, reinicie o `payment` e execute:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Verificacao:

- a chamada falha por `DeadlineExceeded`
- o `order` registra no log que houve timeout ao chamar o `payment`

Remova o `time.Sleep` apos o teste.

### Parte 5 - Testes de shipping, estoque e Docker

#### 1. Produto inexistente

Os produtos `a`, `b` e `c` sao cadastrados automaticamente no estoque.

Teste com um produto fora dessa lista:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"nao-existe","quantity":1,"unit_price":10}]}' \
  127.0.0.1:3000 order.Order/Create
```

Retorno esperado:

```txt
ERROR:
  Code: NotFound
  Message: product nao-existe was not found in inventory
```

#### 2. Fluxo com shipping

Com os tres servicos em execucao:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":6,"unit_price":10}]}' \
  127.0.0.1:3000 order.Order/Create
```

Verificacao:

- o pedido e salvo
- o pagamento e criado
- o `shipping` e chamado
- o prazo de entrega calculado para `6` unidades e `2` dias

Regra do prazo:

- de `1` a `5` unidades -> `1` dia
- de `6` a `10` unidades -> `2` dias
- de `11` a `15` unidades -> `3` dias

Consulta no banco:

```bash
docker exec -it microsservices-mysql mysql -uroot -pminhasenha
```

```sql
USE `shipping`;
SELECT id, order_id, delivery_forecast_days FROM shippings;
```

#### 3. Subir tudo com Docker Compose

```bash
cd microservices
docker compose up --build
```

Depois, execute os testes do `order` pela porta `3000`.

Para encerrar:

```bash
docker compose down
```

## Comandos uteis

Instalar o `grpcurl` no macOS:

```bash
brew install grpcurl
```

Compilar os servicos:

```bash
cd microservices/order
go build ./...

cd ../payment
go build ./...

cd ../shipping
go build ./...
```
