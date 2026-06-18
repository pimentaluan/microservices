# Microsserviço Order

Projeto desenvolvido para a disciplina de Programação Distribuída utilizando:

* Go
* gRPC
* Protocol Buffers
* MySQL
* GORM
* Arquitetura Hexagonal

O microsserviço é responsável pelo cadastro de pedidos de compra.

---

# Estrutura do Projeto

```txt
cmd/
config/
internal/
```

A aplicação segue uma arquitetura hexagonal, separando:

* domínio
* regras de negócio
* portas
* adaptadores

---

# Executando o Banco de Dados

É necessário possuir Docker instalado.

Executar:

```bash
docker run --name order-mysql -p 3306:3306 \
-e MYSQL_ROOT_PASSWORD=minhasenha \
-e MYSQL_DATABASE=order \
-d mysql
```

---

# Executando o Microsserviço

Dentro da pasta `order`:

```bash
DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order?parseTime=true" \
APPLICATION_PORT=3000 \
ENV=development \
go run cmd/main.go
```

---

# Testando com grpcurl

Instalar:

```bash
brew install grpcurl
```

Listar serviços:

```bash
grpcurl -plaintext localhost:3000 list
```

Executar requisição:

```bash
grpcurl \
-d '{"costumer_id":123,"order_items":[{"product_code":"prod","quantity":4,"unit_price":12}]}' \
-plaintext localhost:3000 \
order.Order/Create
```

---

# Repositório protobuf

Os arquivos protobuf e stubs gerados estão disponíveis no repositório:

`microservices-proto`