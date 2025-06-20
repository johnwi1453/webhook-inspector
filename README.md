# Webhook Inspector

Webhook Inspector is a backend-focused developer tool that allows engineers to test and debug webhook integrations by creating temporary public endpoints which log and display incoming HTTP requests. Inspired by tools like RequestBin and webhook.site, this project demonstrates backend architecture, infrastructure design, and Go proficiency.

---

## Tech Stack

### Backend
- **Go (1.24+)**
- **chi** router for HTTP endpoints
- **Redis** for request storage with TTL
- **log/slog or zap** for structured logging
- **OAuth2** authentication (via `goth`)
- **Rate limiting** using Redis token bucket
- **Docker** + **Docker Compose**
- **CI/CD** via GitHub Actions

### Frontend (Optional)
- Go templates or React (TBD)
- Tailwind CSS for styling

---

## Features

- Generate temporary webhook endpoints like `/api/hooks/:token`
- Capture and log method, headers, body, timestamp, and IP
- Store payloads in Redis with 24h TTL
- OAuth2 login with GitHub or Google
- Redis-based rate limiting (e.g. 100 requests/hour)
- Endpoint to view logs: `GET /logs`
- Admin route or CLI to clear logs
- Dockerized app with CI pipeline
- Deployable to platforms like Render/Fly.io

---

## Getting Started

### Prerequisites
- Docker Desktop
- Go 1.24+
- Redis (via Docker or system install)

---

### Local Development

```bash
# Clone and enter the project
git clone git@github.com:johnwi1453/webhook-inspector.git
cd webhook-inspector

# Build the app
go build -o server main.go

# Run the app (or use Docker)
./server
```

### Or use Docker

```bash
docker compose up --build
```

App runs at: http://localhost:8080

---

## Project Structure

```
webhook-inspector/
├── cmd/                  # main.go (moved to root temporarily)
├── internal/
│   ├── handlers/         # HTTP route logic
│   ├── redis/            # Redis client setup
│   ├── auth/             # OAuth2 login
│   └── middleware/       # Logging, rate limiting, etc
├── web/templates/        # Optional frontend templates
├── Dockerfile
├── docker-compose.yml
├── go.mod / go.sum
├── main.go
└── README.md
```

---

## Status

✅ Redis-connected Go server is up  
🛠️ Next: implement `/api/hooks/:token` for receiving requests  
💡 Frontend dashboard optional and in progress

---

## Sample Logged Webhook Format (in Redis)

```json
{
  "id": "a8e2b8d1",
  "method": "POST",
  "timestamp": "2025-06-20T03:50:00Z",
  "ip": "203.0.113.1",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{ \"event\": \"signup\", \"user_id\": 123 }",
  "token": "abc123"
}
```
