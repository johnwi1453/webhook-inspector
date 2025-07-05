# Stage 1 — Build React frontend
FROM node:20 AS frontend
WORKDIR /app
COPY frontend ./frontend
RUN cd frontend && npm install && npm run build

# Stage 2 — Build Go backend
FROM golang:1.24 AS builder
WORKDIR /app
COPY backend ./backend
COPY --from=frontend /app/frontend/dist ./frontend/dist
WORKDIR /app/backend
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook-inspector ./main.go

# Stage 3 — Final image
FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /app/backend/webhook-inspector .
COPY --from=builder /app/frontend/dist ./frontend/dist
COPY --from=builder /app/backend/docs ./docs

USER nonroot:nonroot
EXPOSE 8080
CMD ["/webhook-inspector"]
