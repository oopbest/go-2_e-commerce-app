# 🛍️ Go E-Commerce REST API Backend (Production-Ready)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io)
[![Docker](https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=flat&logo=docker)](https://www.docker.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, enterprise-grade E-Commerce REST API backend built in **Go (Golang)** following **Clean Architecture**, **SOLID Principles**, and **12-Factor App methodology**.

---

## 🌟 Key Features & Engineering Highlights

* 🏗️ **Clean Architecture & Domain-Driven Design**: Strict separation of concerns across Domain, Service, Repository, and HTTP Presentation layers with Interface decoupling.
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
[ Client (Web / Mobile / Postman) ]
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
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Core Business Logic (Service Layer)                      │
│    - Business Validations, Price Calculations, Auth Logic   │
└──────────────────────────────┬──────────────────────────────┘
                               │ (Go Interfaces)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Data Access Layer (Repositories)                         │
│    ├── Cached Repository (Decorator: Redis 7 In-Memory)     │
│    └── PostgreSQL Repository (Database Transactions & Locks)│
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go                         # Server Composition Root & Graceful Shutdown
├── internal/
│   ├── config/                             # 12-Factor Environment Configuration Loader
│   │   └── config.go
│   ├── database/                           # PostgreSQL & Redis Connection Pools
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── domain/                             # Core Entities, DTOs, Interfaces & Errors
│   │   ├── order.go
│   │   ├── product.go
│   │   └── user.go
│   ├── middleware/                         # HTTP Middlewares (Auth, RBAC, Logger, Recovery)
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── recovery.go
│   ├── order/                              # Order & Checkout Module (Transactions + Locking)
│   │   ├── handler.go
│   │   ├── repository.go
│   │   └── service.go
│   ├── product/                            # Product Module (Postgres + Redis Cache Decorator)
│   │   ├── handler.go
│   │   ├── handler_test.go                 # HTTP Handler Unit Tests (httptest)
│   │   ├── repository_cached.go
│   │   ├── repository_postgres.go
│   │   ├── service.go
│   │   └── service_test.go                 # Service Unit Tests (testify/mock)
│   └── user/                               # User & Authentication Module
│       ├── handler.go
│       ├── repository.go
│       └── service.go
├── pkg/
│   └── security/                           # Reusable Security Packages (JWT, Bcrypt)
│       ├── jwt.go
│       ├── jwt_test.go
│       ├── password.go
│       └── password_test.go
├── migrations/
│   └── init.sql                            # PostgreSQL Schema & Seed Data
├── walkthroughs/                           # Complete Step-by-Step Learning Walkthroughs (Phases 1-9)
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-9/
├── .dockerignore
├── .env.example
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

Start all services (API, PostgreSQL 16, Redis 7) with a single command:

```bash
docker compose up --build -d
```

Check container status and health:
```bash
docker compose ps
```

View real-time structured logs:
```bash
docker compose logs -f api
```

The API is now running at `http://localhost:8080`!

---

### Option 2: Run Natively on Host Machine

1. Start database and cache dependencies:
   ```bash
   docker compose up -d postgres redis
   ```
2. Copy environment variables:
   ```bash
   cp .env.example .env
   ```
3. Run the Go server:
   ```bash
   go run ./cmd/api
   ```

---

## 📡 API Endpoints Reference

### 🔐 Authentication & Users
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `POST` | `/api/auth/register` | 🌐 Public | Register new user (Default Role: `customer` or `admin`) |
| `POST` | `/api/auth/login` | 🌐 Public | Authenticate user and receive JWT Bearer token |

### 📦 Products Catalog (Cached in Redis)
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/products` | 🌐 Public | List all products (Cache-Aside, 5m TTL) |
| `GET` | `/api/products/{id}` | 🌐 Public | Get product details by ID |
| `POST` | `/api/products` | 🔒 Admin Only | Create new product (Invalidates Redis Cache) |
| `PUT` | `/api/products/{id}` | 🔒 Admin Only | Update product by ID (Invalidates Redis Cache) |
| `DELETE` | `/api/products/{id}` | 🔒 Admin Only | Delete product by ID (Invalidates Redis Cache) |

### 🛒 Orders & Checkout (Database Transactions & Locking)
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `POST` | `/api/orders/checkout` | 🔒 User | Checkout items with atomic stock deduction |
| `GET` | `/api/orders` | 🔒 User | List order history (Customer: own / Admin: all) |
| `GET` | `/api/orders/{id}` | 🔒 User | Get order details with itemized breakdown |

### 🩺 System & Health
| Method | Endpoint | Access Level | Description |
| :--- | :--- | :---: | :--- |
| `GET` | `/health` | 🌐 Public | Health check report (DB, Redis, Environment) |

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

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
