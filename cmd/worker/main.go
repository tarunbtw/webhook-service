package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tarunbtw/webhook-service/internal/db"
)

func main() {
	log.SetOutput(os.Stdout)

	database := db.New("webhook.db")
	client := &http.Client{Timeout: 10 * time.Second}

	log.Println("worker started, watching for pending webhooks...")

	for {
		deliver(database, client)
		time.Sleep(5 * time.Second)
	}
}

func deliver(database *db.DB, client *http.Client) {
	webhooks, err := database.GetPendingWebhooks()
	if err != nil {
		log.Println("error fetching webhooks:", err)
		return
	}

	if len(webhooks) == 0 {
		return
	}

	endpoints, err := database.GetAllEndpoints()
	if err != nil {
		log.Println("error fetching endpoints:", err)
		return
	}

	for _, webhook := range webhooks {
		allDelivered := true

		for _, endpoint := range endpoints {
			success := attemptDelivery(database, client, webhook, endpoint)
			if !success {
				allDelivered = false
			}
		}

		if allDelivered {
			database.UpdateWebhookStatus(webhook.ID, "delivered")
			log.Printf("webhook %s delivered to all endpoints\n", webhook.ID)
		} else {
			database.UpdateWebhookStatus(webhook.ID, "failed")
			log.Printf("webhook %s failed delivery\n", webhook.ID)
		}
	}
}

func attemptDelivery(database *db.DB, client *http.Client, webhook db.Webhook, endpoint db.Endpoint) bool {
	var lastErr string
	var lastStatus int

	// exponential backoff: try 3 times
	delays := []time.Duration{0, 10 * time.Second, 30 * time.Second}

	for attempt, delay := range delays {
		if delay > 0 {
			log.Printf("retrying webhook %s to %s in %v (attempt %d)\n", webhook.ID, endpoint.URL, delay, attempt+1)
			time.Sleep(delay)
		}

		resp, err := client.Post(endpoint.URL, "application/json", bytes.NewBufferString(webhook.Payload))

		if err != nil {
			lastErr = err.Error()
			lastStatus = 0
			log.Printf("attempt %d failed: %s\n", attempt+1, lastErr)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			database.LogAttempt(webhook.ID, endpoint.ID, resp.StatusCode, "")
			return true
		}

		lastStatus = resp.StatusCode
		lastErr = "non-2xx response"
		log.Printf("attempt %d got status %d\n", attempt+1, lastStatus)
	}

	// all 3 attempts failed
	database.LogAttempt(webhook.ID, endpoint.ID, lastStatus, lastErr)
	return false
}