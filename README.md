# Product Catalog Service

A backend service for managing a product catalog with activation, discounts, and an outbox pattern.  
Built using **Clean Architecture / DDD / CQRS** and **Google Spanner Emulator**.

---

## Architecture Overview

- **domain/** — pure business logic (aggregates, value objects, invariants, domain events)
- **usecases/** — command side (write), orchestration of business flows
- **queries/** — read side, returns DTOs without domain hydration
- **contracts/** — interfaces (ports) between application and infrastructure
- **repo/** — Spanner-based implementations of contracts (mutations, read model)
- **outbox** — domain events are persisted in the same transaction as business data
- **transport/grpc/** — gRPC API (thin transport layer)
- **pkg/clock/** — time abstraction for deterministic tests

---

## Requirements

- Go 1.21+
- Docker and docker-compose
- gcloud CLI

---

## Running Instructions

```bash
# Start Spanner emulator
docker-compose up -d

# Run migrations (create instance, database, schema)
make migrate

# Run tests (unit + e2e)
make test

# Start server
make run
