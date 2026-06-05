package main

import (
	"fmt"
	"log"
	"os"
	"github.com/tarunbtw/webhook-service/internal/db"
)

func main() {
	log.SetOutput(os.Stdout)
	d := db.New("webhook.db")
	fmt.Println("server starting, db connections:", d.Conn.Stats().OpenConnections)
}