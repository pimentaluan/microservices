# Microsservicos com Go, gRPC e Arquitetura Hexagonal

Projeto da disciplina de Programacao Distribuida com dois microsservicos:

- `order`: recebe pedidos de compra
- `payment`: registra a cobranca do pedido

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
│   ├── order/
│   └── payment/
└── microservices-proto/
    ├── order/
    ├── payment/
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

## Como gerar os stubs protobuf

Os arquivos gerados ja estao no repositorio, mas o script tambem foi mantido.

Dentro de `microservices-proto`:

```bash
cd microservices-proto
sh run.sh
```

## Como subir o ambiente

### 1. Subir o MySQL

Abra o Docker Desktop e depois execute:

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

### 2. Subir o microsservico Payment

Em outro terminal:

```bash
cd microservices/payment

DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/payment?parseTime=true" \
APPLICATION_PORT=3001 \
ENV=development \
go run cmd/main.go
```

### 3. Subir o microsservico Order

Em outro terminal:

```bash
cd microservices/order

DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order?parseTime=true" \
PAYMENT_SERVICE_URL="127.0.0.1:3001" \
APPLICATION_PORT=3000 \
ENV=development \
go run cmd/main.go
```

### 4. Confirmar que o Order esta exposto

```bash
grpcurl -plaintext 127.0.0.1:3000 list
```

O esperado e aparecer `order.Order`.

## Como testar cada parte

### Parte 1 - Teste basico do Order

Cria um pedido valido no microsservico `order`:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Resultado esperado:

```json
{
  "orderId": 1
}
```

### Parte 2 - Teste da comunicacao Order -> Payment

Com os dois servicos rodando, execute a mesma requisicao valida:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Se o `orderId` for retornado, o `order` conseguiu salvar o pedido e chamar o `payment` com sucesso.

### Parte 3 - Testes de tratamento de erros

#### 1. Erro por mais de 50 itens

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":51,"unit_price":1}]}' \
  127.0.0.1:3000 order.Order/Create
```

Resultado esperado:

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

Resultado esperado:

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

Resultado esperado:

```json
{
  "orderId": 3
}
```

#### 4. Verificar o status salvo no banco

```bash
docker exec -it microsservices-mysql mysql -uroot -pminhasenha
```

Dentro do MySQL:

```sql
USE `order`;
SELECT id, customer_id, status FROM orders;
```

Esperado:

- pedidos com erro ficam como `Canceled`
- pedidos com sucesso ficam como `Paid`

### Parte 4 - Testes de timeout e retry

#### 1. Teste de retry

Com o `order` rodando, pare o `payment` com `Ctrl+C` e execute:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Comportamento esperado:

- a chamada falha no final
- antes de falhar, o cliente do `order` tenta novamente automaticamente

#### 2. Teste de timeout

Para simular lentidao no `payment`, adicione temporariamente este trecho no inicio da funcao `Create` de [payment/internal/adapters/grpc/server.go](/Users/luanpimenta/Documents/projetos/ifpb/pd/atividade2/microservices/payment/internal/adapters/grpc/server.go:29):

```go
time.Sleep(3 * time.Second)
```

Tambem importe `time`, reinicie o `payment` e execute:

```bash
grpcurl -plaintext \
  -d '{"costumer_id":1,"order_items":[{"product_code":"a","quantity":2,"unit_price":100}]}' \
  127.0.0.1:3000 order.Order/Create
```

Comportamento esperado:

- a chamada falha por `DeadlineExceeded`
- o `order` registra no log que houve timeout ao chamar o `payment`

Depois do teste, remova o `time.Sleep`.

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
```
