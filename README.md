# Webhook Inspector

Webhook Inspector is a backend-focused developer tool that allows engineers to test and debug webhook integrations by creating temporary public endpoints which log and display incoming HTTP requests. Inspired by tools like RequestBin and webhook.site, this project demonstrates backend architecture, infrastructure design, secure session management, and Go proficiency.

---

## Tech Stack

### Backend

* **Go (1.24+)**
* **chi** router for HTTP endpoints
* **Redis** for request storage and rate limiting
* **OAuth2** login with GitHub (via `golang.org/x/oauth2`)
* **Cookie-based session management**
* **Per-token rate limiting** (anonymous: 5 req/day, GitHub: 500 req/day)
* **Docker** + **Docker Compose**

### Frontend (Planned)

* **React** + Tailwind CSS
* Token and log dashboard for authenticated users

---

## Features

* Generate temporary webhook endpoints like `/api/hooks/:token`
* Store and retrieve webhook payloads in Redis with 24h TTL
* Inspect headers, method, body, and timestamp
* Anonymous session support with unique token generation via `/create`
* GitHub login support with persistent tokens and elevated rate limits
* View webhook logs via `/logs` or `/logs/:token`
* Get token info via `/token` and login state via `/me`
* `/status` and `/reset` endpoints for managing usage and cleaning up
* Full API testing support with curl, Postman, or browser

---

## Getting Started

### Prerequisites

* Docker Desktop
* GitHub OAuth App (Client ID + Secret)

### .env Example

```
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
REDIS_ADDR=redis:6379
```

---

### Local Development

```bash
# Clone and enter the project
git clone git@github.com:johnwi1453/webhook-inspector.git
cd webhook-inspector

# Run with Docker
docker compose up --build
```

App runs at: [http://localhost:8080](http://localhost:8080)

---

## API Overview

| Endpoint                    | Description                                    |
| --------------------------- | ---------------------------------------------- |
| `GET /create`               | Assigns a new anonymous token in cookie        |
| `POST /api/hooks`           | Submit a webhook (uses cookie token)           |
| `POST /api/hooks/:token`    | Submit a webhook to a specific token           |
| `GET /logs`                 | View recent webhooks (via cookie token)        |
| `GET /logs/:token`          | View webhooks for a specific token             |
| `GET /auth/github`          | Initiate GitHub OAuth2 login                   |
| `GET /auth/github/callback` | OAuth2 redirect URL                            |
| `GET /me`                   | Show GitHub login session status               |
| `GET /token`                | Get your assigned webhook token (if logged in) |
| `GET /status`               | Show rate limit + TTL for current token        |
| `POST /reset`               | Clear all logs + usage for current token       |

---

## Project Structure

```
webhook-inspector/
├── internal/
│   ├── handlers/         # HTTP route logic
│   ├── redis/            # Redis client setup
│   ├── auth/             # OAuth2 login config
├── Dockerfile
├── docker-compose.yml
├── .env
├── main.go
└── README.md
```

---

## Example Webhook Payload (stored in Redis)

```json
{
  "id": "f6f8b2a3",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{ \"event\": \"signup\", \"user_id\": 123 }",
  "timestamp": "2025-06-22T17:43:00Z"
}
```

---

## Status

✅ Backend logic complete (anonymous + GitHub support)
✅ Redis TTL + rate limiting + storage
✅ Full test coverage via browser + Postman
🛠️ Next: build frontend with React for log visibility and token UX
