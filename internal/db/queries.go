package db

import (
	"log"
	"time"

	"github.com/google/uuid"
)

func (d *DB) CreateEndpoint(url string) (*Endpoint, error) {
	e := &Endpoint{
		ID:        uuid.NewString(),
		URL:       url,
		CreatedAt: time.Now(),
	}
	_, err := d.Conn.Exec(
		`INSERT INTO endpoints (id, url) VALUES (?, ?)`,
		e.ID, e.URL,
	)
	return e, err
}

func (d *DB) CreateWebhook(payload string) (*Webhook, error) {
	w := &Webhook{
		ID:      uuid.NewString(),
		Payload: payload,
		Status:  "pending",
	}
	_, err := d.Conn.Exec(
		`INSERT INTO webhooks (id, payload, status) VALUES (?, ?, ?)`,
		w.ID, w.Payload, w.Status,
	)
	return w, err
}

func (d *DB) GetPendingWebhooks() ([]Webhook, error) {
	rows, err := d.Conn.Query(
		`SELECT id, payload, status FROM webhooks WHERE status = 'pending'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(&w.ID, &w.Payload, &w.Status); err != nil {
			log.Println("scan error:", err)
			continue
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (d *DB) GetAllEndpoints() ([]Endpoint, error) {
	rows, err := d.Conn.Query(`SELECT id, url FROM endpoints`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []Endpoint
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.URL); err != nil {
			continue
		}
		endpoints = append(endpoints, e)
	}
	return endpoints, nil
}

func (d *DB) GetAllWebhooks() ([]Webhook, error) {
	rows, err := d.Conn.Query(
		`SELECT id, payload, status, created_at FROM webhooks ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var w Webhook
		rows.Scan(&w.ID, &w.Payload, &w.Status, &w.CreatedAt)
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (d *DB) GetFailedWebhooks() ([]Webhook, error) {
	rows, err := d.Conn.Query(
		`SELECT id, payload, status, created_at FROM webhooks WHERE status = 'failed'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var w Webhook
		rows.Scan(&w.ID, &w.Payload, &w.Status, &w.CreatedAt)
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (d *DB) LogAttempt(webhookID, endpointID string, statusCode int, errMsg string) error {
	_, err := d.Conn.Exec(
		`INSERT INTO delivery_attempts (id, webhook_id, endpoint_id, status_code, error)
		 VALUES (?, ?, ?, ?, ?)`,
		uuid.NewString(), webhookID, endpointID, statusCode, errMsg,
	)
	return err
}

func (d *DB) UpdateWebhookStatus(id, status string) error {
	_, err := d.Conn.Exec(
		`UPDATE webhooks SET status = ? WHERE id = ?`,
		status, id,
	)
	return err
}

func (d *DB) GetAttemptsForWebhook(webhookID string) ([]DeliveryAttempt, error) {
	rows, err := d.Conn.Query(
		`SELECT id, webhook_id, endpoint_id, status_code, error, attempted_at
		 FROM delivery_attempts WHERE webhook_id = ?`,
		webhookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []DeliveryAttempt
	for rows.Next() {
		var a DeliveryAttempt
		rows.Scan(&a.ID, &a.WebhookID, &a.EndpointID, &a.StatusCode, &a.Error, &a.AttemptedAt)
		attempts = append(attempts, a)
	}
	return attempts, nil
}