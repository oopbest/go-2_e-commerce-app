# 🛍️ Go E-Commerce REST API Backend (Production-Ready)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io)
[![Migrations](https://img.shields.io/badge/golang--migrate-embed.FS-blue?style=flat&logo=postgresql)](https://github.com/golang-migrate/migrate)
[![Asynq Workers](https://img.shields.io/badge/Asynq-Task%20Queue-FF6B6B?style=flat&logo=redis)](https://github.com/hibiken/asynq)
[![Swagger](https://img.shields.io/badge/Swagger-OpenAPI%202.0-85EA2D?style=flat&logo=swagger)](http://localhost:8080/swagger/index.html)
[![Docker](https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=flat&logo=docker)](https://www.docker.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, enterprise-grade E-Commerce REST API and Event-Driven Background Worker system built in **Go (Golang)** following **Clean Architecture**, **SOLID Principles**, and **12-Factor App methodology**.

---

## 🌟 Key Features & Engineering Highlights

* 🏗️ **Clean Architecture & Domain-Driven Design**: Strict separation of concerns across Domain, Service, Repository, and HTTP Presentation layers with Interface decoupling.
* 🗄️ **Enterprise Database Versioning & Auto-Migrations**:
  * Programmatic migrations with **`golang-migrate/migrate/v4`**.
  * Embedded SQL files inside static Go binary using **`embed.FS`** (zero external dependencies).
  * Safe **UP (upgrade)** and **DOWN (rollback)** schema changelogs (`schema_migrations` tracking).
  * Zero-downtime schema evolution (e.g. `000002_add_categories_table`).
* ⚡ **Event-Driven & Asynchronous Background Workers**:
  * Redis-backed Distributed Task Queue powered by **`hibiken/asynq`**.
  * **Immediate Async Jobs**: Simulated email notifications and receipts dispatched in background threads.
  * **Delayed Scheduled Tasks**: Automated order cancellation and **stock restoration** for unpaid orders after timeout.
  * Fault-tolerant task retries with exponential backoff and dedicated worker pools.
* 📖 **Interactive Swagger & OpenAPI Documentation**:
  * Auto-generated OpenAPI 2.0 specifications via **`swaggo/swag`**.
  * Interactive **Swagger UI** (`/swagger/index.html`) supporting live API execution and **BearerAuth** token testing.
* 🔒 **Security & Authentication**:
  * Passwords hashed using **`bcrypt`** (salted & timing-attack resistant).
  * Stateless **JWT Authentication (HS256)** with expiration and role payload.
  * Role-Based Access Control (**RBAC**) middleware (`admin` vs `customer`).
  * Sensitive data sanitization using struct tag `json:"-"`.
* 🛒 **Shopping Cart & Checkout System**:
  * Atomic multi-table **Database Transactions (`db.BeginTx`)** ensuring all-or-nothing order creation.
  * 🛡️ **Pessimistic Row Locking (`SELECT ... FOR UPDATE`)** to prevent race conditions and overselling during high-concurrency checkout.
  * Historical price snapshotting (`price_at_purchase`).
* ⚡ **High Performance Caching**:
  * **Redis 7 In-Memory Caching** with **Decorator Pattern** wrapping repository layers.
  * **Cache-Aside Strategy (Lazy Loading)** with automated TTL eviction.
  * **Automated Cache Invalidation** on product mutations (`Create`, `Update`, `Delete`).
  * **Sub-millisecond (0ms)** response latency for cached endpoints.
* 🛡️ **Production Reliability & Resilience**:
  * **Structured Logging** using Go 1.21+ built-in **`log/slog`** (Text format for Dev, JSON format for Cloud Production).
  * **Custom ResponseWriter** middleware capturing HTTP status codes and exact request latency.
  * **Panic Recovery Middleware** with stack trace logging.
  * **Graceful Shutdown**: Traps OS signals (`SIGINT`, `SIGTERM`), finishes in-flight requests within a 10s timeout, and closes database/redis pools cleanly.
* 🧪 **Automated Testing Suite**:
  * Idiomatic **Table-Driven Tests**.
  * Mocking data access layers using **`testify/mock`**.
  * HTTP handler testing using Go's **`net/http/httptest`** without binding TCP ports.
  * Interactive HTML Code Coverage reporting.
* 🐳 **Cloud & Container Ready**:
  * **Multi-Stage Dockerfile** producing an ultra-lean **~8 MB** static binary container (`CGO_ENABLED=0`, `-ldflags="-w -s"`).
  * Runs with an unprivileged non-root user (`appuser`).
  * Full-stack orchestration via **Docker Compose** with health checks (`pg_isready`, `redis-cli ping`).

---

## 🏛️ Architecture Overview

```text
[ Client (Web / Mobile / Swagger UI / Postman) ]
                        │
                        ▼ (HTTP Requests)
┌─────────────────────────────────────────────────────────────┐
│ 1. Global Middlewares (Recovery -> Logger -> Auth / RBAC)   │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. HTTP Presentation Layer (Handler / Controllers)          │
│    - product.Handler, user.Handler, order.Handler           │
│    - Swagger UI (/swagger/index.html)                       │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Core Business Logic (Service Layer)                      │
│    - Business Validations, Price Calculations, Auth Logic   │
│    - Enqueues background tasks via worker.TaskDistributor   │
└──────────────┬───────────────────────────────┬──────────────┘
               │ (Go Interfaces)               │ (Task Queue)
               ▼                               ▼
┌──────────────────────────────┐ ┌─────────────────────────────┐
│ 4. Data Access Layer         │ │ 5. Redis Asynq Task Queue   │
│    - Redis Cache (Decorator) │ └─────────────┬───────────────┘
│    - PostgreSQL Repositories │               │ (Workers Pull)
│    - Auto-Migrations (embed) │               ▼
└──────────────────────────────┘ ┌─────────────────────────────┐
                                 │ 6. Background Worker Server │
                                 │    - Instant Email Worker   │
                                 │    - 1-Min Auto Cancel &    │
                                 │      Stock Restoral Worker  │
                                 └─────────────────────────────┘
```

---

## 📁 Project Structure

```text
.
├── cmd/
│   ├── api/
│   │   └── main.go                         # REST API Entry Point & Auto-Migration Trigger
│   └── worker/
│       └── main.go                         # Asynq Background Worker Server Entry Point
├── docs/                                   # Auto-Generated Swagger / OpenAPI Documentation
├── internal/
│   ├── config/                             # 12-Factor Environment Configuration Loader
│   ├── database/                           # PostgreSQL, Redis & Auto-Migration Runner
│   │   ├── migration.go                    # golang-migrate runner with embed.FS
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── domain/                             # Core Entities, DTOs, Interfaces & Errors
│   ├── middleware/                         # HTTP Middlewares (Auth, RBAC, Logger, Recovery)
│   ├── order/                              # Order & Checkout Module (Transactions + Locking)
│   ├── product/                            # Product Module (Postgres + Redis Cache Decorator)
│   ├── user/                               # User & Authentication Module
│   └── worker/                             # Event-Driven Background Worker Module
│       ├── distributor.go                  # Task Producer (Interface + Asynq Client)
│       ├── processor.go                    # Task Consumer (Email & Auto-Cancel Handlers)
│       └── task.go                         # Task Definitions & Payloads
├── pkg/
│   └── security/                           # Reusable Security Packages (JWT, Bcrypt)
├── migrations/                             # Versioned Database Migrations (Embedded via embed.FS)
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_add_categories_table.up.sql
│   ├── 000002_add_categories_table.down.sql
│   └── migrations.go                       # //go:embed *.sql
├── walkthroughs/                           # Complete Step-by-Step Learning Walkthroughs (Phases 1-12)
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-12/
├── .dockerignore
├── .env.example
├── .gitignore
├── Dockerfile                              # Multi-Stage Lean Container Build
├── docker-compose.yml                      # Full Stack Compose Definition
├── go.mod
└── go.sum
```

---

## 🚀 Quick Start & Running the Stack

### Prerequisites
* [Docker Desktop](https://www.docker.com/products/docker-desktop) installed and running.
* [Go 1.24+](https://golang.org/dl/) (optional, if running natively).

### Option 1: Run Full Stack with Docker Compose (Recommended)

Start all services (API, Background Worker, PostgreSQL 16, Redis 7) with a single command:

```bash
docker compose up --build -d
```

Check container status and health:
```bash
docker compose ps
```

View real-time structured logs:
```bash
docker compose logs -f api worker
```

* 🌐 **API Base URL**: `http://localhost:8080`
* 📖 **Swagger UI Documentation**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
* 🗄️ **PostgreSQL Port (Host)**: `localhost:15432`

---

## 📚 Complete Learning Walkthroughs

This project was built progressively through a hands-on, zero-to-hero curriculum. You can inspect detailed technical walkthroughs and verification screenshots in the [`walkthroughs/`](walkthroughs/) directory:

1. [Walkthrough 1: In-Memory REST API & Goroutine Concurrency](walkthroughs/walkthrough-1/walkthrough-1.md)
2. [Walkthrough 2: Clean Architecture & Standard Go Layout](walkthroughs/walkthrough-2/walkthrough-2.md)
3. [Walkthrough 3: Real Database with PostgreSQL & Docker Compose](walkthroughs/walkthrough-3/walkthrough-3.md)
4. [Walkthrough 4: User Authentication, bcrypt, JWT Security & RBAC](walkthroughs/walkthrough-4/walkthrough-4.md)
5. [Walkthrough 5: Cart Checkout, DB Transactions & Pessimistic Locking](walkthroughs/walkthrough-5/walkthrough-5.md)
6. [Walkthrough 6: Production Readiness, Structured Logging (`slog`) & Graceful Shutdown](walkthroughs/walkthrough-6/walkthrough-6.md)
7. [Walkthrough 7: High Performance with Redis Caching & Decorator Pattern](walkthroughs/walkthrough-7/walkthrough-7.md)
8. [Walkthrough 8: Automated Testing, Mocking (`testify`) & httptest](walkthroughs/walkthrough-8/walkthrough-8.md)
9. [Walkthrough 9: Multi-Stage Containerization & Full Stack Docker Compose](walkthroughs/walkthrough-9/walkthrough-9.md)
10. [Walkthrough 10: Interactive API Documentation with Swagger & OpenAPI](walkthroughs/walkthrough-10/walkthrough-10.md)
11. [Walkthrough 11: Event-Driven & Asynchronous Background Workers](walkthroughs/walkthrough-11/walkthrough-11.md)
12. [Walkthrough 12: Enterprise Database Migrations & Versioning](walkthroughs/walkthrough-12/walkthrough-12.md)

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
