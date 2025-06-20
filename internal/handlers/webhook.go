package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	fmt.Println("Received webhook for token:", token)
	fmt.Println("Body:", string(body))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}
