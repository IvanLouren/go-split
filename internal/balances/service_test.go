package balances_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/IvanLouren/GoSplit/internal/balances"
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

func TestGetBalances(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 1", "user1@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	var expenseID string
	err = testDB.QueryRow(`INSERT INTO expenses (group_id, paid_by, description, amount) VALUES ($1, $2, $3, $4) RETURNING id`,
		groupID, userID, "Dinner", 90.00).Scan(&expenseID)
	if err != nil {
		t.Fatalf("failed to insert expense: %s", err)
	}

	var user2ID string
	err = testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 2", "user2@test.com", "hashedpassword").Scan(&user2ID)
	if err != nil {
		t.Fatalf("failed to insert user2: %s", err)
	}

	_, err = testDB.Exec(`INSERT INTO expense_splits (expense_id, user_id, amount) VALUES ($1, $2, $3)`,
		expenseID, user2ID, 45.00)
	if err != nil {
		t.Fatalf("failed to insert split: %s", err)
	}

	// parse groupID to uuid
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	// call GetBalances
	service := balances.NewService(testDB)
	result, err := service.GetBalances(parsedGroupID)
	if err != nil {
		t.Fatalf("failed to get balances: %s", err)
	}

	// assert 2 balances returned
	if len(result) != 2 {
		t.Fatalf("expected 2 balances, got %d", len(result))
	}

	// find each user's balance
	for _, b := range result {
		switch b.UserID.String() {
		case userID:
			if b.Balance != 90.00 {
				t.Errorf("expected test user balance 90, got %f", b.Balance)
			}
		case user2ID:
			if b.Balance != -45.00 {
				t.Errorf("expected user2 balance -45, got %f", b.Balance)
			}
		default:
			t.Errorf("unexpected user in balances: %s", b.UserID)
		}
	}
}
