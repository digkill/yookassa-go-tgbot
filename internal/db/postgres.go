package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var DB *sql.DB

func Init(dsn string) {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("DB connection error:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("DB ping error:", err)
	}

	log.Println("✅ Connected to PostgreSQL")
}
