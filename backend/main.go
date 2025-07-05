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

	// Dashboard is homepage
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
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

	// Error handling
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving frontend fallback for: %s", r.URL.Path)
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})

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

	// Serve dist folder
	fs := http.StripPrefix("/", http.FileServer(http.Dir("./frontend/dist")))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Frontend fallback hit for path:", r.URL.Path)
		fs.ServeHTTP(w, r)
	}))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
