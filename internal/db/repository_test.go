package db

import (
	"context"
	"testing"
)

// Negative / Error-path unit tests for db/repository.go

// Error/negative cases tested against properly initialized but constrained connection.
func TestRepository_NegativePaths(t *testing.T) {
	// A DB instance with no connection or a closed connection
	dbMock := &DB{Conn: nil}

	err := dbMock.CreateCompany(context.Background(), &Company{})
	if err == nil {
		t.Fatal("Expected error when DB connection is nil, got nil")
	}

	_, err = dbMock.GetCompanyByID(context.Background(), -1)
	if err == nil {
		t.Fatal("Expected error on invalid ID with nil connection, got nil")
	}

	_, err = dbMock.GetContactByEmail(context.Background(), "")
	if err == nil {
		t.Fatal("Expected error on missing DB/empty email, got nil")
	}

	err = dbMock.UpdateDealState(context.Background(), 1, "Fake_State")
	if err == nil {
		t.Fatal("Expected error on fake state update without connection, got nil")
	}
}
