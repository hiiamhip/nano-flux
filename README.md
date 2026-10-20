# Nanoflux

A lightweight URL shortener and asynchronous click tracking pipeline built with Go, PostgreSQL, and Clean Architecture principles.

> **Note**: This is a personal hands-on project created to explore idiomatic Go patterns, concurrency pipelines, and Clean Architecture while transitioning from traditional MVC frameworks (such as Laravel) to Go backend development.

---

## Overview

Nanoflux provides core URL shortening capabilities paired with a high-throughput, decoupled click tracking pipeline:

1. **URL Shortening**: Generates deterministic/compact Base62 identifiers for long URLs, with conflict handling and validation.
2. **Read/Write Decoupled Tracking**: Instead of writing a database record on every redirect (which can degrade database performance under read-heavy traffic), redirect requests push click events into an in-memory buffered channel.
3. **Batch Event Ingestion**: A dedicated background worker gathers click events and flushes them in batches (either when reaching batch size or on a time ticker interval) into PostgreSQL using `pgx.CopyFrom` (PostgreSQL COPY protocol).
4. **Resilient Lifecycle**: Implements graceful shutdown to ensure in-flight HTTP requests finish and buffered in-memory events are completely drained and persisted before closing connections.

---

## Architecture & Layout

The project follows Clean Architecture (Hexagonal) conventions, keeping domain rules independent of external frameworks or databases:

```
nano-flux/
├── cmd/
│   └── api/
│       └── main.go          # Application entrypoint & explicit dependency injection
├── internal/
│   ├── domain/              # Core entities (URL, Click) & domain errors
│   ├── usecase/             # Application business rules & consumer interfaces (ports)
│   ├── adapter/             # Infrastructure & driver adapters
│   │   ├── handler/http/    # Chi HTTP router, handlers, middlewares & Swagger UI
│   │   ├── repository/      # PostgreSQL repositories (pgxpool & CopyFrom batching)
│   │   ├── shortener/       # Base62 hash-based code generation
│   │   └── tracker/         # Asynchronous in-memory click worker & batcher
│   └── config/              # Environment-based configuration
├── migrations/              # PostgreSQL schema migrations
├── docker-compose.yml       # Local PostgreSQL setup
├── Dockerfile               # Multi-stage container build
└── Makefile                 # Convenient build and test commands
```

### Key Design Principles Applied:
- **Consumer-driven Interfaces**: Interfaces (`URLRepository`, `Shortener`, `ClickTracker`) are declared in `internal/usecase` where they are consumed, rather than in adapter packages.
- **Explicit Dependency Injection**: No global singletons or service containers; all dependencies are passed explicitly via constructors (`New...`).
- **Controlled Goroutine Lifecycles**: The tracker background worker runs in a managed goroutine coordinated with `sync.WaitGroup` and flushes remaining records upon channel closure.
- **Non-blocking Request Handling**: URL resolution does not block on database writes for analytics.

---

## Getting Started

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- `make` (optional, for running Makefile targets)

### 1. Clone & Setup Environment
```bash
git clone https://github.com/hiiamhip/nano-flux.git
cd nano-flux
cp .env.example .env # if needed, ensure DB credentials match docker-compose.yml
```

### 2. Start PostgreSQL
```bash
docker compose up -d postgres
```
Tables will automatically be created via the migration scripts mounted in `/docker-entrypoint-initdb.d`.

### 3. Run the Application
```bash
make run
# or directly:
go run ./cmd/api
```
The server will start listening on `:8080` (or the configured `PORT`).

---

## API Endpoints

| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/api/urls` | Create a short URL from an original URL |
| `GET` | `/{code}` | Redirect to the original URL (tracks click asynchronously) |
| `PUT` | `/api/urls/{id}` | Update the target of an existing short URL |
| `DELETE` | `/api/urls/{id}` | Delete a short URL by ID |
| `GET` | `/swagger` | Interactive Swagger API documentation |

### Example Request

**Create a Short URL:**
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://go.dev/doc/effective_go"}'
```

**Response:**
```json
{
  "short_code": "a1B2c3",
  "short_url": "http://localhost:8080/a1B2c3"
}
```

---

## Testing

Run unit and race detection tests:
```bash
make test
```

Generate coverage report:
```bash
make test-coverage
```
