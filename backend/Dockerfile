FROM golang:1.24.4

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN ls -la /app && go build -v -o /app/server main.go

EXPOSE 8080
CMD ["/app/server"]
