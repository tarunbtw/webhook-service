package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)


type DB struct {
	Conn *sql.DB
}

// New opens the SQLite file and creates tables if they don't exist
func New(path string) *DB {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}

	if err := conn.Ping(); err != nil {
		log.Fatal("failed to ping db:", err)
	}

	d := &DB{Conn: conn}
	d.migrate()
	return d
}

func (d *DB) migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS endpoints (
		id          TEXT PRIMARY KEY,
		url         TEXT NOT NULL,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS webhooks (
		id          TEXT PRIMARY KEY,
		payload     TEXT NOT NULL,
		status      TEXT NOT NULL DEFAULT 'pending',
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS delivery_attempts (
		id           TEXT PRIMARY KEY,
		webhook_id   TEXT NOT NULL,
		endpoint_id  TEXT NOT NULL,
		status_code  INTEGER,
		error        TEXT,
		attempted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (webhook_id)  REFERENCES webhooks(id),
		FOREIGN KEY (endpoint_id) REFERENCES endpoints(id)
	);
	`

	_, err := d.Conn.Exec(schema)
	if err != nil {
		log.Fatal("migration failed:", err)
	}

	log.Println("database ready")
}