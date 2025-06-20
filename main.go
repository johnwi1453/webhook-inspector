package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"webhook-inspector/internal/redis"
)

func main() {
	redis.InitRedis()

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
