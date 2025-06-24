# Webhook Inspector: Usage Guide

This guide explains how to use Webhook Inspector to test and debug incoming HTTP requests. You can either use it anonymously or log in with GitHub for elevated access.

---

## Anonymous User Flow

### 1. Create a webhook token

```bash
curl -i http://localhost:8080/create
```

* This sets a `webhook_token` cookie in your browser

### 2. Send a webhook (anonymous session)

```bash
curl -X POST http://localhost:8080/api/hooks \
  -H "Content-Type: application/json" \
  -d '{"message": "hello world"}'
```

* Rate limit: 5 requests
* TTL: 24 hours

### 3. View your webhooks

```bash
curl http://localhost:8080/logs
```

### 4. Check status and usage

```bash
curl http://localhost:8080/status
```

### 5. Reset logs and usage

```bash
curl -X POST http://localhost:8080/reset
```

---

## GitHub Login Flow

### 1. Start login

Go to:

```
http://localhost:8080/auth/github
```

This will redirect you to GitHub OAuth.

### 2. After login, check session

```bash
curl --cookie "session_token=<your_cookie>" http://localhost:8080/me
```

### 3. Get your webhook token

```bash
curl --cookie "session_token=<your_cookie>" http://localhost:8080/token
```

* You'll receive a token associated with your GitHub username

### 4. Send a webhook using token

```bash
curl -X POST http://localhost:8080/api/hooks/<your_token> \
  -H "Content-Type: application/json" \
  -d '{"event": "signup"}'
```

* Rate limit: 500 requests
* TTL: 24 hours

### 5. View logs for your token

```bash
curl http://localhost:8080/logs/<your_token>
```

### 6. Check status

```bash
curl http://localhost:8080/status
```

* Returns token, owner (if any), rate usage, remaining requests, and TTL

---

## Testing Tips

* Use browser for GitHub login and to trigger cookie storage
* Use Postman or curl for raw API testing
* Copy `session_token` from browser cookies to use in Postman requests
* Tokens created via GitHub login are privileged and reusable
* Anonymous tokens are deleted when you call `/create` again

---

## Full Route Reference

| Method | Path                  | Description                               |
| ------ | --------------------- | ----------------------------------------- |
| GET    | /create               | Assigns anonymous token in cookie         |
| POST   | /api/hooks            | Send webhook (uses cookie token)          |
| POST   | /api/hooks/\:token    | Send webhook using direct token           |
| GET    | /logs                 | View logs for cookie-based token          |
| GET    | /logs/\:token         | View logs for specific token              |
| GET    | /status               | Check request quota + TTL                 |
| POST   | /reset                | Delete all data tied to current token     |
| GET    | /auth/github          | Start GitHub login                        |
| GET    | /auth/github/callback | OAuth2 callback                           |
| GET    | /me                   | View current logged-in user (if any)      |
| GET    | /token                | View GitHub user's assigned webhook token |

---

For live testing and UI support, a React frontend is coming soon.
