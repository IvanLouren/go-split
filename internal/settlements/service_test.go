package settlements_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/IvanLouren/GoSplit/internal/settlements"
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

func TestCreateSettlement(t *testing.T) {
	var paidByID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 1", "user1@test.com", "hashedpassword").Scan(&paidByID)
	if err != nil {
		t.Fatalf("failed to insert paidBy user: %s", err)
	}

	var paidToID string
	err = testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 2", "user2@test.com", "hashedpassword").Scan(&paidToID)
	if err != nil {
		t.Fatalf("failed to insert paidTo user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", paidByID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	parsedPaidByID, err := uuid.Parse(paidByID)
	if err != nil {
		t.Fatalf("failed to parse paidByID: %s", err)
	}

	parsedPaidToID, err := uuid.Parse(paidToID)
	if err != nil {
		t.Fatalf("failed to parse paidToID: %s", err)
	}

	service := settlements.NewService(testDB)
	settlement, err := service.CreateSettlement(parsedGroupID, parsedPaidByID, parsedPaidToID, 45.00)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}

	if settlement.ID == uuid.Nil {
		t.Errorf("expected settlement ID to be set, got nil")
	}
	if settlement.GroupID != parsedGroupID {
		t.Errorf("expected groupID %s, got %s", parsedGroupID, settlement.GroupID)
	}
	if settlement.PaidBy != parsedPaidByID {
		t.Errorf("expected paidBy %s, got %s", parsedPaidByID, settlement.PaidBy)
	}
	if settlement.PaidTo != parsedPaidToID {
		t.Errorf("expected paidTo %s, got %s", parsedPaidToID, settlement.PaidTo)
	}
	if settlement.Amount != 45.00 {
		t.Errorf("expected amount 45.00, got %f", settlement.Amount)
	}
}

func TestGetSettlements(t *testing.T) {
	var paidByID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 3", "user3@test.com", "hashedpassword").Scan(&paidByID)
	if err != nil {
		t.Fatalf("failed to insert paidBy user: %s", err)
	}

	var paidToID string
	err = testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"User 4", "user4@test.com", "hashedpassword").Scan(&paidToID)
	if err != nil {
		t.Fatalf("failed to insert paidTo user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Paris", paidByID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	parsedPaidByID, err := uuid.Parse(paidByID)
	if err != nil {
		t.Fatalf("failed to parse paidByID: %s", err)
	}

	parsedPaidToID, err := uuid.Parse(paidToID)
	if err != nil {
		t.Fatalf("failed to parse paidToID: %s", err)
	}

	service := settlements.NewService(testDB)
	_, err = service.CreateSettlement(parsedGroupID, parsedPaidByID, parsedPaidToID, 45.00)
	if err != nil {
		t.Fatalf("failed to create settlement: %s", err)
	}

	result, err := service.GetSettlements(parsedGroupID)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}

	if len(result) == 0 {
		t.Fatalf("expected at least 1 settlement, got 0")
	}
	if result[0].Amount != 45.00 {
		t.Errorf("expected amount 45.00, got %f", result[0].Amount)
	}
	if result[0].GroupID != parsedGroupID {
		t.Errorf("expected groupID %s, got %s", parsedGroupID, result[0].GroupID)
	}
}
