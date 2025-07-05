package main

import (
	"log"
	"net/http"
	"os"
	"webhook-inspector/internal/handlers"
	"webhook-inspector/internal/redis"

	"github.com/go-chi/chi/v5"
)

func main() {
	redis.InitRedis()
	log.Println("GITHUB_CLIENT_ID =", os.Getenv("GITHUB_CLIENT_ID"))

	r := chi.NewRouter()

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

	// Get token status
	r.Get("/status", handlers.GetTokenStatus)

	// Reset current token
	r.Post("/reset", handlers.ResetToken)

	// Delete individual webhook
	r.Delete("/logs/{id}", handlers.DeleteWebhook)

	// Login via Github
	r.Get("/auth/github", handlers.GitHubLogin)
	r.Get("/auth/github/callback", handlers.GitHubCallback)

	// Get current logged-in user
	r.Get("/me", handlers.GetCurrentUser)

	// Logout user
	r.Get("/logout", handlers.Logout)

	// Swagger UI for API documentation
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	r.Get("/docs/", handlers.SwaggerUI)
	r.Get("/docs/api-spec.yaml", handlers.SwaggerSpec)

	// Error handling
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("No matching route for %s %s\n", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	// Serve dist folder
	fs := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", fs) // fallback route to frontend

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
