# webhook-delivery-service

A minimal webhook delivery service written in Go. Accepts incoming events, persists them, and delivers them to registered endpoints with automatic retries, a full audit log, and a dead letter queue.

Built to solve a real problem: when a downstream server is unavailable, webhooks fired directly are lost silently. This service guarantees delivery by storing events before attempting any network call.

Similar in scope to [Hookdeck](https://hookdeck.com) and [Svix](https://svix.com).

---

## How it works

Incoming webhooks are stored to SQLite immediately on receipt. A background worker polls for pending webhooks every 5 seconds and attempts HTTP delivery to all registered endpoints. Failed deliveries are retried with exponential backoff. Every attempt is logged with its status code and timestamp. Webhooks that exhaust all retries are marked failed and held in a dead letter queue, where they can be inspected and replayed at any time.

```
POST /webhooks
      |
      v
  SQLite (status: pending)
      |
      v
  Worker goroutine
      |
      v
  HTTP POST to registered endpoints
      |
      +-- 2xx --> status: delivered, attempt logged
      |
      +-- error --> retry (0s, 10s, 30s)
                        |
                        +-- still failing --> status: failed
                                                  |
                                                  v
                                            dead letter queue
                                                  |
                                                  v
                                        POST /webhooks/:id/replay
                                                  |
                                                  v
                                            status: pending (retried)
```

The server and worker run as a single process in production. In `cmd/worker/` there is a standalone worker binary that can be run separately for horizontal scaling in environments that support multiple services.

---

## API

### Register an endpoint

```
POST /endpoints
Content-Type: application/json

{"url": "https://your-server.com/webhook"}
```

### Send a webhook

```
POST /webhooks
Content-Type: application/json

{"event": "form.submitted", "data": {}}
```

Response is `202 Accepted`. The payload is stored and queued immediately.

### List all webhooks

```
GET /webhooks/all
```

### List failed webhooks (dead letter queue)

```
GET /webhooks/failed
```

Returns all webhooks that exhausted their retry attempts. These are your dead letters — nothing was silently dropped.

### Inspect delivery attempts

```
GET /webhooks/:id/attempts
```

Returns every delivery attempt for a webhook with timestamp, HTTP status code, and error message.

### Replay a failed webhook

```
POST /webhooks/:id/replay
```

Resets a failed webhook back to `pending`. The worker picks it up within 5 seconds and reattempts delivery with the full retry cycle.

---

## Running locally

```bash
git clone https://github.com/tarunbtw/webhook-delivery-service
cd webhook-delivery-service
go run cmd/server/main.go
```

Server starts on `:8080`. Dashboard available at `http://localhost:8080`.

To test delivery, use [webhook.site](https://webhook.site) to get a free receiver URL, register it as an endpoint, and send a webhook. To test the dead letter queue, register a URL that does not exist and watch the worker exhaust retries, then replay it after fixing the endpoint.

---

## Docker

```bash
docker build -f Dockerfile.server -t webhook-server .
docker run -p 8080:8080 -e DB_PATH=/app/data/webhook.db webhook-server
```

Or with Compose:

```bash
docker-compose up
```

---

## Stack

- Go 1.26
- SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)
- No frameworks — standard library only for HTTP
- Deployed on Render via Docker

### Notes

Uses SQLite with ephemeral storage on free-tier hosting. On restart, the database is cleared. For production use, swap the storage layer to PostgreSQL and use a persistent volume.

---

## Project structure

```
cmd/
  server/   HTTP server + embedded worker goroutine
  worker/   standalone worker binary for separate-process deployments
internal/
  db/       schema, migrations, query functions
```