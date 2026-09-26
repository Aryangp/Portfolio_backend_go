# 🚀 Portfolio Backend API & AI Chatbot (Go)

A clean, idiomatic, high-performance REST API backend built in **Go** featuring **MongoDB caching**, **Server-Sent Events (SSE) streaming**, and an **AI Chatbot powered by Google Gemini** with a **dynamic function calling / tool execution engine**.

---

## 🌟 Key Features

- **⚡ High Performance**: Built with Go's standard library `net/http` using clean architecture and modern route matching.
- **🤖 Real-Time Streaming AI Chatbot**: Server-Sent Events (SSE) streaming powered by Google Gemini (`gemini-3.5-flash-lite`, `gemini-1.5-flash`, etc.).
- **🛠️ Dynamic Function Calling Engine**:
  - Thread-safe **Tool Registry** that dynamically builds Gemini tool declarations.
  - Multi-turn execution loop handling `functionCall` and `functionResponse` with **thought signature** preservation.
  - Ready for easily registering new custom tools (LinkedIn, search, etc.).
- **🐙 Live GitHub Tool Suite**:
  - `get_github_user_profile`: Live statistics (bio, followers, public repo count, profile link).
  - `get_github_repositories`: Live public repository list with stars, forks, languages, topics, and descriptions.
  - `get_github_repo_details`: Detailed repo metadata with raw `README.md` excerpt.
- **⚡ Pikachu AI Persona**: Friendly, energetic, emoji-packed assistant that speaks in Pikachu tone, respects Aryan's portfolio guardrails, and responds in under 100 words ending with *"Pika pika! ⚡"*.
- **🗄️ Cached MongoDB Integration**: Connects to MongoDB Atlas (`portfolio_db.projects`), warms an in-memory cache at startup, and provides automatic fallback to local memory storage.
- **🛡️ Production Middleware**:
  - Streaming-safe structured HTTP logger (preserves SSE flushing).
  - Configurable CORS (for React, Vite, Next.js frontend integration).
  - Panic recovery middleware.
- **🛑 Graceful Shutdown**: Handles `SIGINT` / `SIGTERM` signals with timeout context and closes database/HTTP connections cleanly.
- **🧪 Unit Tested**: Comprehensive test coverage across handlers, tools, and repositories.

---

## 📁 Directory Structure

```
portfolio-backend/
├── cmd/
│   └── api/
│       └── main.go                 # Server entrypoint & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Environment configuration (.env loader)
│   ├── handler/
│   │   ├── chat.go                 # SSE Chatbot stream handler (GET & POST)
│   │   ├── health.go               # Health check handler
│   │   ├── helpers.go              # JSON response helpers & error formatting
│   │   └── project.go              # Project CRUD handlers
│   ├── middleware/
│   │   └── middleware.go           # Streaming logger, CORS, & recovery middlewares
│   ├── model/
│   │   ├── chat.go                 # Chat message models
│   │   ├── project.go              # Project models & DTOs
│   │   └── response.go             # Standard JSON response envelope
│   ├── prompts/
│   │   └── chat_bot_prompt.go      # Pikachu assistant system prompt & guardrails
│   ├── repository/
│   │   ├── chat_repo.go            # Gemini SSE streaming & function calling loop
│   │   ├── mongo.go                # MongoDB Atlas connection helper
│   │   └── projectRepository.go    # Cached MongoDB & in-memory project store
│   ├── router/
│   │   └── router.go               # HTTP routing & middleware wiring
│   └── tools/
│       ├── github.go               # GitHub REST API client & tool suite
│       ├── registry.go             # Dynamic thread-safe Tool Registry
│       └── types.go                # Tool interface & Gemini schema types
├── .env
├── Makefile
├── go.mod
└── README.md
```

---

## ⚡ Getting Started

### Prerequisites
- **Go 1.22+**
- (Optional) **MongoDB Atlas URI** for project storage
- (Optional) **Google Gemini API Key** for AI chatbot
- (Optional) **GitHub Token** for higher GitHub API rate limits

### 1. Configure Environment
Copy or edit `.env` in the root directory:

```env
PORT=8080
APP_ENV=development
ALLOWED_ORIGINS=*

# MongoDB (Optional)
MONGO_URI=mongodb+srv://<user>:<password>@cluster.mongodb.net/?appName=portfolio
MONGO_DB_NAME=portfolio_db
MONGO_PROJECTS_COLL=projects

# Google Gemini AI
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_MODEL=gemini-3.5-flash-lite

# GitHub Tools
GITHUB_USERNAME=Aryangp
GITHUB_TOKEN=your_github_token_here
```

### 2. Run Locally

```bash
# Using Makefile
make run

# Or directly with Go
go run ./cmd/api
```

The server will start on `http://localhost:8080`.

### 3. Build & Test

```bash
# Run all unit tests
make test

# Build production binary
make build
./bin/api-server
```

---

## 📡 API Endpoints

### 1. AI Chatbot (SSE Real-Time Stream)

#### `GET /api/v1/chat/stream`
Stream a single prompt in real-time via Server-Sent Events (SSE).

```bash
curl -N "http://localhost:8080/api/v1/chat/stream?message=What+repositories+does+Aryan+have+on+GitHub%3F"
```

#### `POST /api/v1/chat/stream`
Stream with conversation history and prompt via JSON payload.

```bash
curl -N -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "How many repositories does Aryan have?",
    "history": [
      {"role": "user", "content": "Hi there!"},
      {"role": "model", "content": "Pika pika! ⚡ Hello! How can I help you explore Aryan'\''s projects?"}
    ]
  }'
```

**Stream Output Format:**
```
data: {"chunk":"⚡ Hey"}
data: {"chunk":" there! Aryan has 12"}
data: {"chunk":" public repositories on GitHub! 🚀"}
data: {"chunk":"\n\nPika pika! ⚡💛"}
data: {"done":true}
data: [DONE]
```

---

### 2. Projects API

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/v1/projects` | List all projects (supports `?category=`, `?tag=`, `?featured=true`) |
| `GET` | `/api/v1/projects/{id}` | Get project by ID |
| `POST` | `/api/v1/projects` | Create a new project |
| `PUT` | `/api/v1/projects/{id}` | Update an existing project |
| `DELETE` | `/api/v1/projects/{id}` | Delete a project |

#### Example: List Projects
```bash
curl http://localhost:8080/api/v1/projects
curl "http://localhost:8080/api/v1/projects?featured=true"
curl "http://localhost:8080/api/v1/projects?category=AI/ML"
```

---

### 3. Health Check

#### `GET /api/v1/health`
```bash
curl http://localhost:8080/api/v1/health
```

**Response:**
```json
{
  "success": true,
  "message": "Portfolio Backend API is running smoothly",
  "data": {
    "environment": "development",
    "status": "healthy",
    "timestamp": "2026-09-26T21:30:00Z",
    "uptime": "5m12s",
    "version": "1.0.0"
  }
}
```

---

## 🛠️ Adding New AI Tools

The Tool System is designed to be **completely dynamic**. To add a new tool (e.g., LinkedIn Profile, Weather, Resume Fetcher):

1. **Implement the `Tool` interface** ([`internal/tools/types.go`](internal/tools/types.go)):
```go
type MyCustomTool struct {}

func (t *MyCustomTool) Name() string { return "my_custom_tool" }
func (t *MyCustomTool) Description() string { return "Describes what the tool does for Gemini" }
func (t *MyCustomTool) Declaration() tools.FunctionDeclaration {
    return tools.FunctionDeclaration{
        Name:        t.Name(),
        Description: t.Description(),
        Parameters: &tools.Schema{
            Type: tools.TypeObject,
            Properties: map[string]*tools.Schema{
                "query": {Type: tools.TypeString, Description: "Search query"},
            },
            Required: []string{"query"},
        },
    }
}
func (t *MyCustomTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // Perform tool logic here
    return map[string]string{"result": "data"}, nil
}
```

2. **Register it in [`cmd/api/main.go`](cmd/api/main.go)**:
```go
toolRegistry.Register(&MyCustomTool{})
```

The tool will automatically be declared to Gemini and routed during chat interactions without modifying any other part of the system!

---

## ⚙️ Environment Variables Reference

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | Port for the HTTP server to listen on |
| `APP_ENV` | `development` | Application environment (`development`, `staging`, `production`) |
| `ALLOWED_ORIGINS` | `*` | Comma-separated CORS origins (e.g. `http://localhost:3000,https://myportfolio.com`) |
| `MONGO_URI` | `""` | MongoDB connection URI (fallback to in-memory store if empty) |
| `MONGO_DB_NAME` | `portfolio_db` | MongoDB Database name |
| `MONGO_PROJECTS_COLL` | `projects` | MongoDB collection for projects |
| `GEMINI_API_KEY` | `""` | Google Gemini API key (chatbot runs in offline fallback if empty) |
| `GEMINI_MODEL` | `gemini-1.5-flash` | Gemini model name (e.g. `gemini-3.5-flash-lite`, `gemini-1.5-flash`) |
| `GITHUB_USERNAME` | `Aryangp` | Default GitHub username for GitHub tools |
| `GITHUB_TOKEN` | `""` | Optional GitHub Personal Access Token for higher rate limits |

---

## 📜 License

MIT License. Built for Aryan Gupta's Interactive Developer Portfolio.
