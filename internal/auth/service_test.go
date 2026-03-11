package auth_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/IvanLouren/GoSplit/internal/auth"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("gosplit_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}
	defer testDB.Close()

	if err := runMigrations(testDB); err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}

	os.Setenv("JWT_SECRET", "test-secret")

	os.Exit(m.Run())

}

func runMigrations(db *sql.DB) error {
	migration, err := os.ReadFile("../../migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}
	_, err = db.Exec(string(migration))
	return err
}

func TestRegister(t *testing.T) {
	service := auth.NewService(testDB)

	user, err := service.Register("User 1", "user1@test.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %s", err)
	}

	if user.ID == uuid.Nil {
		t.Errorf("expected user ID to be set, got nil")
	}

	if user.Name != "User 1" {
		t.Errorf("expected Name 'User 1', got %s", user.Name)
	}

	if user.Email != "user1@test.com" {
		t.Errorf("expected Name 'user1@test.com', got %s", user.Email)
	}

	if user.Password == "password123" {
		t.Errorf("expected password to be hashed, got plain text")
	}
}

func TestLogin(t *testing.T) {
	service := auth.NewService(testDB)

	_, err := service.Register("User 2", "user2@test.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %s", err)
	}

	token, err := service.Login("user2@test.com", "password123")
	if err != nil {
		t.Fatalf("failed to log user: %s", err)
	}

	if token == "" {
		t.Errorf("expected a JWT token, got empty string")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	service := auth.NewService(testDB)

	_, err := service.Register("User 3", "user3@test.com", "password123")
	if err != nil {
		t.Fatalf("failed to register user: %s", err)
	}
	_, err = service.Login("user3@test.com", "password12")
	if err == nil {
		t.Fatalf("expected error for wrong password, got nil")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	service := auth.NewService(testDB)

	_, err := service.Login("user0101@test.com", "password123")
	if err == nil {
		t.Fatalf("expected error for nonexistent user, got nil")
	}
}
