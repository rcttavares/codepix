# CodePix

A Go microservice that acts as a central intermediary for Pix transactions between banks, using **gRPC** for key registration/lookup and **Apache Kafka** for asynchronous transaction processing.

## Architecture

```text
Banks ──► Kafka (transactions topic) ──► CodePix ──► Kafka (bank-specific topic)
                                            │
                                            └──► gRPC (PixKey registration & lookup)
```

The service exposes two integration points:

- **gRPC server** (port `50051`) — registers Pix keys and looks them up by kind/value
- **Kafka consumer** — processes incoming transactions and confirmations, routing them to the destination bank's topic

## Domain

| Entity | Description |
| ------ | ----------- |
| `Bank` | Financial institution |
| `Account` | Bank account linked to a bank |
| `PixKey` | Key (`email` or `cpf`) tied to an account |
| `Transaction` | Transfer between two accounts via a Pix key |

## Project Structure

```text
.
├── cmd/                   # CLI commands (cobra)
│   ├── grpc.go            # starts gRPC server only
│   ├── kafka.go           # starts Kafka consumer only
│   └── all.go             # starts both concurrently
├── domain/model/          # core domain entities and business rules
├── application/
│   ├── grpc/              # gRPC server implementation + protobuf definitions
│   ├── kafka/             # Kafka producer and consumer/processor
│   ├── usecase/           # application use cases (pix, transaction)
│   ├── factory/           # use case factories
│   └── model/             # application-level transaction model (JSON marshalling)
└── infrastructure/
    ├── db/                # database connection (PostgreSQL / SQLite)
    └── repository/        # GORM repository implementations
```

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

## Getting Started

**1. Start all services** (app, PostgreSQL, Kafka, Zookeeper, Control Center, pgAdmin):

```bash
docker compose up
```

**2. Enter the app container:**

```bash
docker compose exec app bash
```

**3. Seed the database with fixture data:**

```bash
go run main.go fixtures
```

**4. Start gRPC server + Kafka consumer:**

```bash
go run main.go all
```

## CLI Commands

| Command | Description |
| ------- | ----------- |
| `go run main.go grpc` | Start the gRPC server only (default port `50051`) |
| `go run main.go kafka` | Start the Kafka consumer only |
| `go run main.go all` | Start both gRPC and Kafka concurrently |
| `go run main.go fixtures` | Seed the database with sample banks and accounts |

### Flags

```text
grpc --port, -p       gRPC server port (default: 50051)
all  --grpc-port, -p  gRPC port when running both services (default: 500051)
```

## gRPC API

Defined in [application/grpc/protofiles/pixkey.proto](application/grpc/protofiles/pixkey.proto).

```protobuf
service PixService {
  rpc RegisterPixKey (PixKeyRegistration) returns (PixKeyCreatedResult);
  rpc Find           (PixKey)             returns (PixKeyInfo);
}
```

You can interact with the server using [Evans](https://github.com/ktr0731/evans) (included in the Docker image):

```bash
evans --host localhost --port 50051 --reflection repl
```

## Kafka Topics

| Topic | Direction | Description |
| ----- | --------- | ----------- |
| `transactions` | inbound | New transaction requests from banks |
| `transaction_confirmation` | inbound | Confirmation/completion from destination bank |
| `bank<CODE>` | outbound | Routed transaction forwarded to the destination bank |

Configuration is provided via environment variables:

| Variable | Description |
| -------- | ----------- |
| `kafkaBootstrapServers` | Kafka broker address |
| `kafkaConsumerGroupId` | Consumer group ID |
| `kafkaTransactionTopic` | Inbound transaction topic name |
| `kafkaTransactionConfirmationTopic` | Inbound confirmation topic name |

## Infrastructure Services

| Service | URL / Port | Credentials |
| ------- | ---------- | ----------- |
| PostgreSQL | `localhost:5432` | `POSTGRES_PASSWORD=root`, DB: `codepix` |
| pgAdmin | [localhost:9000](http://localhost:9000) | `admin@user.com` / `123456` |
| Kafka | `localhost:9094` (external) | — |
| Confluent Control Center | [localhost:9021](http://localhost:9021) | — |

## Tech Stack

- **Go 1.24**
- **gRPC / protobuf** — `google.golang.org/grpc`
- **Apache Kafka** — `confluentinc/confluent-kafka-go`
- **PostgreSQL** — `lib/pq` + GORM
- **SQLite** — for testing via GORM
- **Cobra + Viper** — CLI and configuration
- **govalidator** — domain model validation

## License

[MIT](LICENSE)
