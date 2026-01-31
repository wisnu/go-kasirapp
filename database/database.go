package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"kasir-api/config"

	_ "github.com/lib/pq"
)

func Connect(cfg config.DBConfig) (*sql.DB, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, strconv.Itoa(cfg.Port), cfg.Name, cfg.SSLMode,
	)
	db, db_err := sql.Open("postgres", dsn)
	if db_err != nil {
		return nil, db_err
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
