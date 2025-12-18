package postgres_test

import (
	"testing"

	"cgap/internal/postgres"
)

func TestNew_InvalidDSN(t *testing.T) {
	_, err := postgres.New("invalid-dsn")
	if err == nil {
		t.Error("Expected error for invalid DSN")
	}
}

func TestNew_EmptyDSN(t *testing.T) {
	_, err := postgres.New("")
	if err == nil {
		t.Error("Expected error for empty DSN")
	}
}
