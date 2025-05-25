package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

var Conn *sql.DB

func ConnectDB() {
	dsn := os.Getenv("CLICKHOUSE_DSN")
	if dsn == "" {
		log.Fatal("CLICKHOUSE_DSN must be set")
	}

	conn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatalf("failed to open ClickHouse via database/sql: %v", err)
	}

	conn.SetConnMaxLifetime(time.Minute * 5)

	if err := conn.Ping(); err != nil {
		log.Fatalf("failed to ping ClickHouse: %v", err)
	}

	Conn = conn
	log.Println("✓ Connected to ClickHouse (via database/sql)")
}
