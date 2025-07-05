# Stage 1 — Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app

# Copy everything
COPY . .

# Download deps
RUN go mod download

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook-inspector ./main.go

# Stage 2 — Slim production image
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/webhook-inspector .
COPY --from=builder /app/docs ./docs

# Run as non-root
USER nonroot:nonroot

EXPOSE 8080

CMD ["/webhook-inspector"]
