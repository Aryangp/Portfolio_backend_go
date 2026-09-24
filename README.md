# Portfolio REST API Backend (Go)

A clean, idiomatic, high-performance REST API backend built in Go for developer portfolios.

## 🚀 Key Features

- **Zero External Dependencies**: Built with Go's standard library `net/http` using modern HTTP pattern matching.
- **Clean Architecture**: Decoupled layers (`cmd/api`, `internal/handler`, `internal/repository`, `internal/model`, `internal/middleware`, `internal/config`).
- **Pre-Seeded In-Memory Storage**: Thread-safe repository pre-populated with sample projects; ready to swap in PostgreSQL, SQLite, or MongoDB.
- **Production Middleware**:
  - Structured HTTP Request Logger (timing, path, status, remote address)
  - Configurable CORS (for Next.js / Vite / React frontend integration)
  - Panic Recovery middleware
- **Graceful Shutdown**: Handles `SIGINT` / `SIGTERM` signals with timeout context.
- **Unit Tested**: Full test coverage for handlers and repositories.

---

## 📁 Directory Structure

```
portfolio-backend/
├── cmd/
│   └── api/
│       └── main.go               # Application entrypoint & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go             # Environment variable configuration
│   ├── handler/
│   │   ├── contact.go            # Contact form handlers
│   │   ├── health.go             # Health check handler
│   │   ├── helpers.go            # JSON parsing & response writers
│   │   ├── project.go            # Projects CRUD handlers
│   │   └── handlers_test.go      # Handler unit tests
│   ├── middleware/
│   │   └── middleware.go         # Logger, Recovery, CORS middlewares
│   ├── model/
│   │   ├── contact.go            # Contact message models & validation
│   │   ├── project.go            # Project models & DTOs
│   │   └── response.go           # Standard JSON response envelopes
│   ├── repository/
│   │   ├── contact_repo.go       # Contact messages repository
│   │   ├── project_repo.go       # Projects repository with seed data
│   │   └── project_repo_test.go  # Repository unit tests
│   └── router/
│       └── router.go             # HTTP mux & middleware wiring
├── .env.example
├── .gitignore
├── go.mod
├── Makefile
└── README.md
```

---

## ⚡ Getting Started

### Prerequisites
- Go 1.22+ (verified on Go 1.27)

### Run Server

```bash
# Using Makefile
make run

# Or directly using Go
go run ./cmd/api
```

The server will start on `http://localhost:8080`.

### Build Binary

```bash
make build
# Binary created at bin/api-server
./bin/api-server
```

### Run Tests

```bash
make test
```

---

## 📡 API Endpoints & Examples

### Base / Overview
`GET /`

```bash
curl http://localhost:8080/
```

### 1. Health Check
`GET /api/v1/health`

```bash
curl http://localhost:8080/api/v1/health
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Portfolio Backend API is running smoothly",
  "data": {
    "environment": "development",
    "status": "healthy",
    "timestamp": "2026-09-13T16:40:00Z",
    "uptime": "2m30s",
    "version": "1.0.0"
  }
}
```

---

### 2. Projects API

#### List all projects
`GET /api/v1/projects`

Query Parameters (optional):
- `category` (e.g. `?category=Fullstack`)
- `tag` (e.g. `?tag=Go`)
- `featured` (e.g. `?featured=true`)

```bash
curl http://localhost:8080/api/v1/projects
curl "http://localhost:8080/api/v1/projects?category=Systems"
curl "http://localhost:8080/api/v1/projects?featured=true"
```

#### Get project by ID
`GET /api/v1/projects/{id}`

```bash
curl http://localhost:8080/api/v1/projects/proj_1
```

#### Create new project
`POST /api/v1/projects`

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "title": "AI Code Reviewer",
    "description": "Automated pull request analysis using LLMs and AST parsing.",
    "category": "AI/ML",
    "tags": ["Go", "OpenAI", "Git"],
    "featured": true,
    "demo_url": "https://demo.example.com",
    "github_url": "https://github.com/example/ai-reviewer"
  }'
```

#### Update project
`PUT /api/v1/projects/{id}`

```bash
curl -X PUT http://localhost:8080/api/v1/projects/proj_1 \
  -H "Content-Type: application/json" \
  -d '{
    "featured": false,
    "description": "Updated project description."
  }'
```

#### Delete project
`DELETE /api/v1/projects/{id}`

```bash
curl -X DELETE http://localhost:8080/api/v1/projects/proj_1
```

---

### 3. Contact Form API

#### Submit contact message
`POST /api/v1/contact`

```bash
curl -X POST http://localhost:8080/api/v1/contact \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "email": "jane@example.com",
    "subject": "Job Opportunity",
    "message": "Hi, I checked out your portfolio and would like to chat!"
  }'
```

#### List submitted messages
`GET /api/v1/contact`

```bash
curl http://localhost:8080/api/v1/contact
```

---

## ⚙️ Environment Variables

Create a `.env` file (see `.env.example`):

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | Port for the HTTP server to listen on |
| `APP_ENV` | `development` | Application environment (`development`, `staging`, `production`) |
| `ALLOWED_ORIGINS` | `*` | Comma-separated list of allowed CORS origins or `*` |
