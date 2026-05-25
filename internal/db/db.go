package db

import (
	"database/sql"
	"fmt"
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
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{Conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}
