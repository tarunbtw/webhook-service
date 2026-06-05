package db

import "time"

type Endpoint struct {
	ID        string
	URL       string
	CreatedAt time.Time
}

type Webhook struct {
	ID        string
	Payload   string
	Status    string
	CreatedAt time.Time
}

type DeliveryAttempt struct {
	ID          string
	WebhookID   string
	EndpointID  string
	StatusCode  int
	Error       string
	AttemptedAt time.Time
}