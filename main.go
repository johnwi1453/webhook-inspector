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

	// Home page
	r.Get("/", handlers.Home)

	// Get health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Post a new webhook
	r.Route("/api", func(r chi.Router) {
		r.Route("/hooks", func(r chi.Router) {
			r.Post("/", handlers.HandleWebhook)
			r.Post("/{token}", handlers.HandleWebhook)
		})
	})

	// Create a webhook token for new users
	r.Get("/create", handlers.CreateSession)

	// Get logs
	r.Get("/logs", handlers.GetWebhookLogs)
	r.Get("/logs/{token}", handlers.GetWebhookLogs)

	// Error handling
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("⚠️  No matching route for %s %s\n", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
