# 🛍️ Go E-Commerce REST API Backend (Production-Ready)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io)
[![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-E6522C?style=flat&logo=prometheus)](http://localhost:9090)
[![Grafana](https://img.shields.io/badge/Grafana-Dashboard-F46800?style=flat&logo=grafana)](http://localhost:3000)
[![Migrations](https://img.shields.io/badge/golang--migrate-embed.FS-blue?style=flat&logo=postgresql)](https://github.com/golang-migrate/migrate)
[![Asynq Workers](https://img.shields.io/badge/Asynq-Task%20Queue-FF6B6B?style=flat&logo=redis)](https://github.com/hibiken/asynq)
[![Swagger](https://img.shields.io/badge/Swagger-OpenAPI%202.0-85EA2D?style=flat&logo=swagger)](http://localhost:8080/swagger/index.html)
[![Docker](https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=flat&logo=docker)](https://www.docker.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, enterprise-grade E-Commerce REST API, Event-Driven Background Worker, and Cloud Observability system built in **Go (Golang)** following **Clean Architecture**, **SOLID Principles**, and **12-Factor App methodology**.

---

## 🌟 Key Features & Engineering Highlights

* 🏗️ **Clean Architecture & Domain-Driven Design**: Strict separation of concerns across Domain, Service, Repository, and HTTP Presentation layers with Interface decoupling.
* 🔒 **Security & Authentication**:
  * 👑 **Secure Admin Creation CLI (`cmd/create-admin`)**: Out-of-band admin management tool with terminal password masking (`golang.org/x/term`), RFC 5322 email validation, 12+ character enforcement, and **Privilege Escalation prevention** (public registration restricted to `customer`).
  * Passwords hashed using **`bcrypt`** (salted & timing-attack resistant).
  * Stateless **JWT Authentication (HS256)** with expiration and role payload.
  * Role-Based Access Control (**RBAC**) middleware (`admin` vs `customer`) with centralized domain constants.
  * Sensitive data sanitization using struct tag `json:"-"`.
* 📊 **Cloud Observability & Real-Time Metrics (Prometheus + Grafana)**:
  * Prometheus metrics instrumentation using **`prometheus/client_golang`**.
  * Custom **HTTP Metrics Middleware**: RPS counter, in-flight requests gauge, and **P95/P99 latency histogram**.
  * **Cardinality Protection**: Automated URL path normalization (e.g. `/api/products/{id}`).
  * **Business Analytics Metrics**: Real-time order placement counters and total sales revenue in THB.
  * Live **Grafana Dashboards** visualizing application health and Goroutine allocation.
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
  * 6-Service orchestration via **Docker Compose** with health checks (`pg_isready`, `redis-cli ping`).

---

## 🏛️ Architecture Overview

```text
[ Client (Web / Mobile / Swagger UI / Postman) ]
                        │
                        ▼ (HTTP Requests)
┌─────────────────────────────────────────────────────────────┐
│ 1. Global Middlewares                                       │
│    - RecoveryMiddleware -> MetricsMiddleware -> Logger      │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. HTTP Presentation Layer (Handler / Controllers)          │
│    - product.Handler, user.Handler, order.Handler           │
│    - Swagger UI (/swagger/index.html)                       │
│    - Prometheus Metrics Endpoint (/metrics)                 │
└──────────────┬───────────────────────────────┬──────────────┘
               │                               │
               ▼                               ▼ (Scrapes every 5s)
┌──────────────────────────────┐ ┌─────────────────────────────┐
│ 3. Core Business Logic       │ │ 4. Prometheus Server (9090) │
│    - Order, Product, User    │ └─────────────┬───────────────┘
│    - Auto-Migrations (embed) │               │ (PromQL)
└──────────────┬───────────────┘               ▼
               │ (Tasks)         ┌─────────────────────────────┐
               ▼                 │ 5. Grafana Dashboard (3000) │
┌──────────────────────────────┐ └─────────────────────────────┘
│ 6. Redis Task Queue (asynq)  │
└──────────────┬───────────────┘
               │ (Workers Pull)
               ▼
┌──────────────────────────────┐
│ 7. Background Worker Server  │
│    - Instant Email Worker    │
│    - 1-Min Auto-Cancel &     │
│      Stock Restoral Worker   │
└──────────────────────────────┘
```

---

## 📁 Project Structure

```text
.
├── cmd/
│   ├── api/
│   │   └── main.go                         # REST API Entry Point, Metrics & Migrations
│   ├── create-admin/                       # Secure Admin Creation CLI Tool
│   │   ├── main.go
│   │   └── main_test.go
│   └── worker/
│       └── main.go                         # Asynq Background Worker Server Entry Point
├── deploy/
│   └── prometheus/
│       └── prometheus.yml                  # Prometheus Scrape Target Configuration
├── docs/                                   # Auto-Generated Swagger / OpenAPI Documentation
├── internal/
│   ├── config/                             # 12-Factor Environment Configuration Loader
│   ├── database/                           # PostgreSQL, Redis & Auto-Migration Runner
│   │   ├── migration.go                    # golang-migrate runner with embed.FS
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── domain/                             # Core Entities, DTOs, Role Constants & Interfaces
│   ├── metrics/                            # Prometheus Metric Definitions (RPS, Latency, Business)
│   ├── middleware/                         # HTTP Middlewares (Auth, RBAC, Logger, Recovery, Metrics)
│   ├── order/                              # Order & Checkout Module (Transactions + Locking)
│   ├── product/                            # Product Module (Postgres + Redis Cache Decorator)
│   ├── user/                               # User & Authentication Module
│   └── worker/                             # Event-Driven Background Worker Module
├── pkg/
│   └── security/                           # Reusable Security Packages (JWT, Bcrypt)
├── migrations/                             # Versioned Database Migrations (Embedded via embed.FS)
├── walkthroughs/                           # Complete Step-by-Step Learning Walkthroughs (Phases 1-13)
├── .dockerignore
├── .env.example
├── .gitignore
├── Dockerfile                              # Multi-Stage Lean Container Build
├── docker-compose.yml                      # Full Stack (API, Worker, Postgres, Redis, Prometheus, Grafana)
├── go.mod
└── go.sum
```

---

## 🚀 Quick Start & Running the Stack

### Prerequisites
* [Docker Desktop](https://www.docker.com/products/docker-desktop) installed and running.
* [Go 1.24+](https://golang.org/dl/) (optional, if running natively).

### Option 1: Run Full 6-Container Stack with Docker Compose (Recommended)

Start all services with a single command:

```bash
docker compose up --build -d
```

Check container status and health:
```bash
docker compose ps
```

* 🌐 **API Base URL**: `http://localhost:8080`
* 📖 **Swagger UI Documentation**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
* 📊 **Prometheus Server**: [http://localhost:9090](http://localhost:9090)
* 📈 **Grafana Live Dashboard**: [http://localhost:3000](http://localhost:3000) *(User: `admin` / Password: `admin`)*
* 🗄️ **PostgreSQL Port (Host)**: `localhost:15432`

---

### 👑 Creating an Administrator Account (CLI)

Public registration (`POST /api/auth/register`) strictly registers users with the `customer` role to prevent privilege escalation. To create an `admin` user with elevated access (for product management), run the interactive CLI tool with hidden password input:

```bash
go run ./cmd/create-admin/main.go --email admin@example.com
```

*(You will be prompted to enter and confirm a password with a minimum of 12 characters)*.

---

## 📡 API Endpoints Reference

### 🔐 Authentication & Users
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `POST` | `/api/auth/register` | 🌐 Public | Register new customer account (Role: `customer`) |
| `POST` | `/api/auth/login` | 🌐 Public | Authenticate user and receive JWT Bearer token |

### 📦 Products Catalog (Cached in Redis)
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/products` | 🌐 Public | List all products (Cache-Aside, 5m TTL) |
| `GET` | `/api/products/{id}` | 🌐 Public | Get product details by ID |
| `POST` | `/api/products` | 🔒 Admin Only | Create new product (Invalidates Redis Cache) |
| `PUT` | `/api/products/{id}` | 🔒 Admin Only | Update product by ID (Invalidates Redis Cache) |
| `DELETE` | `/api/products/{id}` | 🔒 Admin Only | Delete product by ID (Invalidates Redis Cache) |

### 🛒 Orders & Checkout (Database Transactions & Asynchronous Workers)
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `POST` | `/api/orders/checkout` | 🔒 User | Checkout items with atomic stock deduction & async tasks |
| `GET` | `/api/orders` | 🔒 User | List order history (Customer: own / Admin: all) |
| `GET` | `/api/orders/{id}` | 🔒 User | Get order details with itemized breakdown |

### 🩺 System, Health & Observability
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `GET` | `/health` | 🌐 Public | Health check report (DB, Redis, Environment) |
| `GET` | `/metrics` | 🌐 Public | Prometheus raw metrics scraping endpoint |
| `GET` | `/swagger/index.html` | 🌐 Public | Interactive Swagger UI API Documentation |

---

## 🧪 Testing & Code Coverage

Run the full automated test suite with verbose output:
```bash
go test -v ./...
```

Generate and view interactive HTML code coverage in your browser:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

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
13. [Walkthrough 13: Cloud Observability & Real-Time Metrics](walkthroughs/walkthrough-13/walkthrough-13.md)

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
