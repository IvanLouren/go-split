package groups_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/IvanLouren/GoSplit/internal/groups"
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

func TestCreateGroup(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User", "test@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	service := groups.NewService(testDB)

	group, err := service.CreateGroup("Trip to Rome", parsedUserID)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}

	if group.ID == uuid.Nil {
		t.Errorf("expected group ID to be set, got nil")
	}

	if group.Name != "Trip to Rome" {
		t.Errorf("expected group Name 'Trip to Rome', got %s", group.Name)
	}

	if group.CreatedBy != parsedUserID {
		t.Errorf("expected createdBy %s, got %s", parsedUserID, group.CreatedBy)
	}
}

func TestGetGroups(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User", "test2@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	_, err = testDB.Exec(`INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`, groupID, userID) // insert a member so that GetGroups can find it
	if err != nil {
		t.Fatalf("failed to insert group member: %s", err)
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	service := groups.NewService(testDB)
	result, err := service.GetGroups(parsedUserID)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}

	if len(result) == 0 {
		t.Fatalf("failed to get at least 1 group, got %d", len(result))
	}
	if result[0].Name != "Trip to Rome" {
		t.Errorf("expected group Name 'Trip to Rome', got %s", result[0].Name)
	}
}

func TestGetGroup(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 3", "test3@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := groups.NewService(testDB)
	group, err := service.GetGroup(parsedGroupID)
	if err != nil {
		t.Fatalf("failed to get group: %s", err)
	}

	if group.ID != parsedGroupID {
		t.Errorf("expected group ID %s, got %s", parsedGroupID, group.ID)
	}

	if group.Name != "Trip to Rome" {
		t.Errorf("expected group name 'Trip to Rome', got %s", group.Name)
	}
}

func TestUpdateGroup(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 4", "test4@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := groups.NewService(testDB)
	updGroup, err := service.UpdateGroup(parsedGroupID, "New Name")
	if err != nil {
		t.Fatalf("failed to update group: %s", err)
	}

	if updGroup.Name != "New Name" {
		t.Errorf("expected group name 'New Name', got %s", updGroup.Name)
	}

	if updGroup.ID != parsedGroupID {
		t.Errorf("expected group ID %s, got %s", parsedGroupID, updGroup.ID)
	}
}

func TestDeleteGroup(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 5", "test5@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}
	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := groups.NewService(testDB)
	err = service.DeleteGroup(parsedGroupID)
	if err != nil {
		t.Fatalf("failed to delete group: %s", err)
	}
	_, err = service.GetGroup(parsedGroupID)
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	}
}

func TestAddMember(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 6", "test6@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}
	var memberID string
	err = testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 7", "test7@test.com", "hashedpassword").Scan(&memberID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedMemberID, err := uuid.Parse(memberID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := groups.NewService(testDB)
	err = service.AddMember(parsedGroupID, parsedMemberID)
	if err != nil {
		t.Fatalf("failed to add member: %s", err)
	}

	groupMember, err := service.GetGroups(parsedMemberID)
	if err != nil {
		t.Fatalf("failed to get groups: %s", err)
	}
	if len(groupMember) == 0 {
		t.Errorf("expected member to be in at least 1 group, got 0")
	}

}

func TestRemoveMember(t *testing.T) {
	var userID string
	err := testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 8", "test8@test.com", "hashedpassword").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}
	var memberID string
	err = testDB.QueryRow(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		"Test User 9", "test9@test.com", "hashedpassword").Scan(&memberID)
	if err != nil {
		t.Fatalf("failed to insert user: %s", err)
	}

	var groupID string
	err = testDB.QueryRow(`INSERT INTO groups (name, created_by) VALUES ($1, $2) RETURNING id`,
		"Trip to Rome", userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %s", err)
	}

	parsedMemberID, err := uuid.Parse(memberID)
	if err != nil {
		t.Fatalf("failed to parse userID: %s", err)
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		t.Fatalf("failed to parse groupID: %s", err)
	}

	service := groups.NewService(testDB)
	err = service.AddMember(parsedGroupID, parsedMemberID)
	if err != nil {
		t.Fatalf("failed to add member: %s", err)
	}

	err = service.RemoveMember(parsedGroupID, parsedMemberID)
	if err != nil {
		t.Fatalf("failed to remove member: %s", err)
	}

	groupMember, err := service.GetGroups(parsedMemberID)
	if err != nil {
		t.Fatalf("failed to get groups: %s", err)
	}
	if len(groupMember) != 0 {
		t.Errorf("expected member to be 0 group, got %d", len(groupMember))
	}
}
