package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tarunbtw/webhook-service/internal/db"
)

type Server struct {
	db *db.DB
}

func main() {
	log.SetOutput(os.Stdout)

	database := db.New("webhook.db")

	s := &Server{db: database}

	http.HandleFunc("/endpoints", s.handleEndpoints)
	http.HandleFunc("/webhooks", s.handleWebhooks)
	http.HandleFunc("/webhooks/", s.handleWebhookAttempts)

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *Server) handleEndpoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		http.Error(w, "invalid body, need {url}", http.StatusBadRequest)
		return
	}

	endpoint, err := s.db.CreateEndpoint(body.URL)
	if err != nil {
		http.Error(w, "failed to create endpoint", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(endpoint)
}

func (s *Server) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(body)
	webhook, err := s.db.CreateWebhook(string(payload))
	if err != nil {
		http.Error(w, "failed to store webhook", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(webhook)
}

func (s *Server) handleWebhookAttempts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL is /webhooks/{id}/attempts
	// r.URL.Path gives us the full path, we split it
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[2] != "attempts" {
		http.Error(w, "use /webhooks/{id}/attempts", http.StatusBadRequest)
		return
	}

	webhookID := parts[1]
	attempts, err := s.db.GetAttemptsForWebhook(webhookID)
	if err != nil {
		http.Error(w, "failed to fetch attempts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attempts)
}