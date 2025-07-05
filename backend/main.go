package main

import (
	"fmt"
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

	r.Route("/api", func(r chi.Router) {
		// Webhooks
		r.Route("/hooks", func(r chi.Router) {
			r.Post("/", handlers.HandleWebhook)
			r.Post("/{token}", handlers.HandleWebhook)
		})

		// Token mgmt
		r.Get("/create", handlers.CreateSession)
		r.Get("/logs", handlers.GetWebhookLogs)
		r.Get("/status", handlers.GetTokenStatus)
		r.Post("/reset", handlers.ResetToken)
		r.Delete("/logs/{id}", handlers.DeleteWebhook)

		// Auth
		r.Get("/auth/github", handlers.GitHubLogin)
		r.Get("/auth/github/callback", handlers.GitHubCallback)
		r.Get("/me", handlers.GetCurrentUser)
		r.Get("/logout", handlers.Logout)
	})

	// GitHub OAuth callback (needs to be outside /api for GitHub redirect)
	r.Get("/auth/github/callback", handlers.GitHubCallback)

	// Logout route (also handle outside /api for direct access)
	r.Get("/logout", handlers.Logout)

	// Dashboard routes - serve the React SPA
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})

	// Get health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Swagger UI for API documentation
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	r.Get("/docs/", handlers.SwaggerUI)
	r.Get("/docs/api-spec.yaml", handlers.SwaggerSpec)

	// Debug endpoint to check files
	r.Get("/debug/files", func(w http.ResponseWriter, r *http.Request) {
		files, err := os.ReadDir("./frontend/dist")
		if err != nil {
			http.Error(w, "Could not read dist", 500)
			return
		}
		for _, f := range files {
			fmt.Fprintln(w, f.Name())
		}
	})

	// Serve static files from frontend/dist
	fs := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/assets/*", fs)
	r.Handle("/vite.svg", fs)

	// Serve index.html for all other routes (SPA fallback)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving frontend fallback for: %s", r.URL.Path)
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
