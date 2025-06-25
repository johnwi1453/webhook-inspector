package main

import (
	"log"
	"net/http"
	"os"
	"webhook-inspector/internal/handlers"
	"webhook-inspector/internal/redis"

	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Load .env values before anything else
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	redis.InitRedis()
	log.Println("GITHUB_CLIENT_ID =", os.Getenv("GITHUB_CLIENT_ID"))

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

	// Get token status
	r.Get("/status", handlers.GetTokenStatus)

	// Reset current token
	r.Post("/reset", handlers.ResetToken)

	// Login via Github
	r.Get("/auth/github", handlers.GitHubLogin)
	r.Get("/auth/github/callback", handlers.GitHubCallback)

	// Get current logged-in user
	r.Get("/me", handlers.GetCurrentUser)

	// Get info about logged-in user's token
	r.Get("/token", handlers.GetWebhookToken)

	// Logout user
	r.Get("/logout", handlers.Logout)

	// Error handling
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("No matching route for %s %s\n", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
