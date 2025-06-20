package main

import (
	"log"
	"net/http"
	"webhook-inspector/internal/handlers"
	"webhook-inspector/internal/redis"

	"github.com/go-chi/chi/v5"
)

func main() {
	redis.InitRedis()

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/hooks", func(r chi.Router) {
			r.Post("/{token}", handlers.HandleWebhook)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("⚠️  No matching route for %s %s\n", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
