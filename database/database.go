package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sql.DB, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(2 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	log.Println("Database connected succesfully")

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// Config should be created by the caller (e.g., viper in main).
