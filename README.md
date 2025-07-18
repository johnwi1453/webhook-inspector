# Webhook Inspector

Webhook Inspector is a backend-focused developer tool that allows engineers to test and debug webhook integrations by creating temporary public endpoints which log and display incoming HTTP requests. Inspired by tools like RequestBin and webhook.site.

---

## Tech Stack

### Backend

* **Go (1.24+)**
* **chi** router for HTTP endpoints
* **Redis** for request storage and rate limiting
* **OAuth2** login with GitHub (via `golang.org/x/oauth2`)
* **Cookie-based session management**
* **Per-token rate limiting** (anonymous: 50 req/day, GitHub: 500 req/day)
* **Docker** + **Docker Compose**

### Frontend

* **React** + **Vite** for fast development
* **Tailwind CSS** for styling
* **React Router** for client-side routing
* Interactive dashboard for webhook inspection
* Real-time log viewing and management

---

## Features

* Generate temporary webhook endpoints like `/hooks/:token`
* Store and retrieve webhook payloads in Redis with 24h TTL
* Inspect headers, method, body, and timestamp
* Anonymous session support with unique token generation via `/create`
* GitHub login support with persistent tokens and elevated rate limits
* View webhook logs via `/logs`
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

## How to Use

### Quick Start Guide

#### 1. **Get Your Webhook URL**

**Option A: Anonymous User (50 requests/day)**
```bash
# Get a temporary webhook token (sets cookie)
curl -c cookies.txt http://localhost:8080/create

# Response: "Assigned new anonymous token: abc123..."
# Your webhook URL: http://localhost:8080/hooks
```

**Option B: GitHub User (500 requests/day)**
```bash
# Login with GitHub for higher limits
open http://localhost:8080/auth/github

# After login, your webhook URL: http://localhost:8080/hooks
```

#### 2. **Send Test Webhooks**

```bash
# Send a test webhook (uses cookie authentication)
curl -b cookies.txt -X POST http://localhost:8080/hooks \
  -H "Content-Type: application/json" \
  -d '{"event": "user.signup", "user_id": 12345, "email": "test@example.com"}'

# Send with custom headers
curl -b cookies.txt -X POST http://localhost:8080/hooks \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Source: stripe" \
  -H "X-Signature: sha256=abc123" \
  -d '{"type": "payment.succeeded", "amount": 2000}'

# Alternative: Send to specific token (no cookie needed)
curl -X POST http://localhost:8080/hooks/your-token \
  -H "Content-Type: application/json" \
  -d '{"event": "test", "data": "example"}'
```

#### 3. **View Your Webhooks**

```bash
# View all received webhooks (uses cookie)
curl -b cookies.txt http://localhost:8080/logs

# Check your usage and limits (uses cookie)
curl -b cookies.txt http://localhost:8080/status
```

### Detailed Usage Examples

#### **Testing Stripe Webhooks**
```bash
# Simulate Stripe payment webhook
curl -b cookies.txt -X POST http://localhost:8080/hooks \
  -H "Content-Type: application/json" \
  -H "Stripe-Signature: t=1234567890,v1=abc123def456" \
  -d '{
    "id": "evt_1234567890",
    "object": "event",
    "type": "payment_intent.succeeded",
    "data": {
      "object": {
        "id": "pi_1234567890",
        "amount": 2000,
        "currency": "usd",
        "status": "succeeded"
      }
    }
  }'
```

#### **Testing GitHub Webhooks**
```bash
# Simulate GitHub push webhook
curl -b cookies.txt -X POST http://localhost:8080/hooks \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-GitHub-Delivery: 12345678-1234-1234-1234-123456789012" \
  -d '{
    "ref": "refs/heads/main",
    "commits": [
      {
        "id": "abc123def456",
        "message": "Fix webhook handling",
        "author": {
          "name": "John Doe",
          "email": "john@example.com"
        }
      }
    ]
  }'
```

#### **Using in Your Application Code**

**Node.js Example:**
```javascript
const axios = require('axios');

// Your webhook endpoint
const webhookUrl = 'http://localhost:8080/hooks/your-token';

// Send webhook from your app
async function sendWebhook(eventData) {
  try {
    await axios.post(webhookUrl, {
      event: 'user.action',
      data: eventData,
      timestamp: new Date().toISOString()
    });
    console.log('Webhook sent successfully');
  } catch (error) {
    console.error('Webhook failed:', error.message);
  }
}
```

**Python Example:**
```python
import requests
import json
from datetime import datetime

webhook_url = 'http://localhost:8080/hooks/your-token'

def send_webhook(event_data):
    payload = {
        'event': 'user.action',
        'data': event_data,
        'timestamp': datetime.utcnow().isoformat() + 'Z'
    }
    
    try:
        response = requests.post(webhook_url, json=payload)
        response.raise_for_status()
        print('Webhook sent successfully')
    except requests.exceptions.RequestException as e:
        print(f'Webhook failed: {e}')
```

### **Rate Limits & Usage**

#### **Anonymous Users**
- **Limit**: 50 requests per 24 hours
- **Reset**: Automatic after 24 hours
- **Token**: Temporary, stored in browser cookie

#### **GitHub Users**
- **Limit**: 500 requests per 24 hours  
- **Reset**: Automatic after 24 hours
- **Token**: Persistent across sessions
- **Login**: Visit `/auth/github` to authenticate

#### **Check Your Usage**
```bash
# View current usage and remaining requests (uses cookie)
curl -b cookies.txt http://localhost:8080/status

# Response example:
{
  "token": "your-token",
  "requests_used": 15,
  "requests_remaining": 485,
  "limit": 500,
  "ttl": "18h 45m",
  "owner": "your-github-username",
  "privileged": true
}
```

#### **Reset Your Data**
```bash
# Clear all webhooks and reset usage counter (uses cookie)
curl -b cookies.txt -X POST http://localhost:8080/reset

# Note: This generates a new token and clears all stored webhooks
```

### **Managing Individual Webhooks**

```bash
# Delete a specific webhook by ID
curl -b cookies.txt -X DELETE http://localhost:8080/logs/webhook-id-here

# Get webhook ID from the logs response
curl -b cookies.txt http://localhost:8080/logs
```

### **Authentication & Sessions**

#### **Check Login Status**
```bash
# See if you're logged in with GitHub
curl -b cookies.txt http://localhost:8080/me

# Response if logged in:
{
  "logged_in": true,
  "username": "your-github-username"
}
```

#### **Logout**
```bash
# Logout and get a new anonymous token
curl -b cookies.txt http://localhost:8080/logout
```

### **Webhook Data Format**

Each received webhook is stored with this structure:
```json
{
  "id": "f6f8b2a3-4c5d-6e7f-8901-234567890abc",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "X-Custom-Header": "custom-value",
    "User-Agent": "MyApp/1.0"
  },
  "body": "{\"event\": \"user.signup\", \"user_id\": 12345}",
  "timestamp": "2025-06-23T14:30:45.123Z"
}
```


---

## API Overview

| Endpoint                    | Description                                    |
| --------------------------- | ---------------------------------------------- |
| `GET /create`               | Assigns a new anonymous token in cookie        |
| `POST /hooks`               | Submit a webhook (uses cookie token)           |
| `POST /hooks/:token`        | Submit a webhook to a specific token           |
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
├── backend/
│   ├── main.go                    # Main Go server entry point
│   ├── go.mod                     # Go module dependencies
│   ├── go.sum                     # Go module checksums
│   ├── Dockerfile                 # Backend Docker configuration
│   ├── docker-compose.yml        # Local development setup
│   ├── docs/
│   │   ├── api-spec.yaml         # OpenAPI/Swagger specification
│   │   └── usage.md              # API usage documentation
│   └── internal/
│       ├── handlers/             # HTTP route handlers
│       │   ├── webhook.go        # Webhook receiving logic
│       │   ├── oauth.go          # GitHub OAuth handlers
│       │   ├── status.go         # Token status endpoints
│       │   ├── reset.go          # Token reset functionality
│       │   └── docs.go           # Swagger documentation serving
│       ├── redis/                # Redis client setup
│       │   └── client.go         # Redis connection management
│       ├── auth/                 # OAuth2 authentication
│       │   └── github.go         # GitHub OAuth configuration
│       ├── config/               # Application configuration
│       │   └── config.go         # Config constants and settings
│       └── models/               # Data models
│           └── model.go          # Webhook payload structures
├── frontend/
│   ├── index.html                # Main HTML template
│   ├── package.json              # Node.js dependencies
│   ├── package-lock.json         # Dependency lock file
│   ├── vite.config.ts            # Vite build configuration
│   ├── tailwind.config.js        # Tailwind CSS configuration
│   ├── tsconfig.json             # TypeScript configuration
│   ├── public/                   # Static assets
│   └── src/
│       ├── main.jsx              # React application entry point
│       ├── index.css             # Global styles
│       ├── components/           # React components
│       │   ├── Header.jsx        # Navigation header
│       │   ├── LogList.jsx       # Webhook logs list
│       │   ├── LogDetails.jsx    # Individual log viewer
│       │   ├── TokenStatus.jsx   # Token info display
│       │   └── TestWebhookForm.jsx # Webhook testing form
│       └── pages/
│           └── Dashboard.jsx     # Main dashboard page
├── Dockerfile                    # Multi-stage Docker build
├── railway.toml                  # Railway deployment configuration
├── docker-compose.yml            # Full-stack development setup
├── .env.example                  # Environment variables template
└── README.md                     # Project documentation
```

---
