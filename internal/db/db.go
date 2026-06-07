package db

import (
	"database/sql"
	"fmt"
	"time"
)

// DB handles database interactions.
type DB struct {
	Conn *sql.DB
}

// NewDB creates a new database instance.
func NewDB(dataSourceName string) (*DB, error) {
	conn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pooling for production resilience
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{Conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}
