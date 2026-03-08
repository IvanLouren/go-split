package expenses_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/IvanLouren/GoSplit/internal/expenses"
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

func TestCreateExpense(t *testing.T) {
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

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	// splits
	splits := []expenses.SplitInput{
		{UserID: parsedUserID, Amount: 90.00},
	}

	service := expenses.NewService(testDB)
	expense, err := service.CreateExpense(parsedGroupID, parsedUserID, "Dinner", 90.00, splits)
	if err != nil {
		t.Fatalf("failed to create expense: %s", err)
	}

	// assert
	if expense.ID == uuid.Nil {
		t.Errorf("expected expense ID to be set, got nil")
	}
	if expense.Description != "Dinner" {
		t.Errorf("expected description 'Dinner', got %s", expense.Description)
	}
	if expense.Amount != 90.00 {
		t.Errorf("expected amount 90, got %f", expense.Amount)
	}
	if expense.GroupID != parsedGroupID {
		t.Errorf("expected groupID %s, got %s", parsedGroupID, expense.GroupID)
	}
	if expense.PaidBy != parsedUserID {
		t.Errorf("expected paidBy %s, got %s", parsedUserID, expense.PaidBy)
	}
}

func TestGetExpenses(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 2", "user2@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := expenses.NewService(testDB)
	splits := []expenses.SplitInput{
		{UserID: parsedUserID, Amount: 90.00},
	}
	_, err = service.CreateExpense(parsedGroupID, parsedUserID, "Dinner", 90.00, splits)
	if err != nil {
		t.Fatalf("failed to create expense: %s", err)
	}

	result, err := service.GetExpenses(parsedGroupID)
	if err != nil {
		t.Fatalf("failed to get expenses: %s", err)
	}

	if len(result) == 0 {
		t.Fatalf("expected at least 1 expense, got 0")
	}
	if result[0].Description != "Dinner" {
		t.Errorf("expected description 'Dinner', got %s", result[0].Description)
	}
	if result[0].Amount != 90.00 {
		t.Errorf("expected amount 90, got %f", result[0].Amount)
	}
}

func TestGetExpense(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 3", "user3@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedUserID, _ := uuid.Parse(userID)
	parsedGroupID, _ := uuid.Parse(groupID)

	service := expenses.NewService(testDB)
	splits := []expenses.SplitInput{
		{UserID: parsedUserID, Amount: 90.00},
	}
	expense, err := service.CreateExpense(parsedGroupID, parsedUserID, "Dinner", 90.00, splits)
	if err != nil {
		t.Fatalf("failed to create expense: %s", err)
	}

	result, err := service.GetExpense(expense.ID)
	if err != nil {
		t.Fatalf("failed to get expense: %s", err)
	}

	if result.ID != expense.ID {
		t.Errorf("expected expense ID %s, got %s", expense.ID, result.ID)
	}
	if result.Description != "Dinner" {
		t.Errorf("expected description 'Dinner', got %s", result.Description)
	}
}

func TestDeleteExpense(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 4", "user4@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := expenses.NewService(testDB)
	splits := []expenses.SplitInput{
		{UserID: parsedUserID, Amount: 90.00},
	}
	expense, err := service.CreateExpense(parsedGroupID, parsedUserID, "Dinner", 90.00, splits)
	if err != nil {
		t.Fatalf("failed to create expense: %s", err)
	}
	err = service.DeleteExpense(expense.ID)
	if err != nil {
		t.Fatalf("failed to delete expense: %s", err)
	}

	// assert expense no longer exists
	result, err := service.GetExpenses(parsedGroupID)
	if err != nil {
		t.Fatalf("failed to get expenses: %s", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 expenses after delete, got %d", len(result))
	}
}
