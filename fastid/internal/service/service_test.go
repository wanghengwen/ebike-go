package service

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"luopingtech-fastid-register/internal/model"
	"luopingtech-fastid-register/internal/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a MySQL test database connection using environment variables
// or default local credentials. The tables are auto-migrated and cleaned before each test.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := envOrDefault("TEST_DB_HOST", "127.0.0.1")
	port := envOrDefault("TEST_DB_PORT", "3306")
	user := envOrDefault("TEST_DB_USER", "fastid")
	pass := envOrDefault("TEST_DB_PASS", "testfastid321#")
	name := envOrDefault("TEST_DB_NAME", "fastid_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping test: cannot connect to MySQL at %s:%s/%s: %v", host, port, name, err)
	}

	// Migrate tables
	if err := db.AutoMigrate(&model.App{}, &model.Machine{}); err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Clean tables before test
	db.Exec("DELETE FROM t_machine")
	db.Exec("DELETE FROM t_app")

	return db
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestGetMachineID_ValidationError_EmptyAppName(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	_, err := svc.GetMachineID("", "", "", "192.168.1.1:8080")
	if err == nil {
		t.Fatal("expected error for empty appName")
	}
	if err.Error() != "appName 和 machineUuid 都不能为空" {
		t.Errorf("error = %q, want %q", err.Error(), "appName 和 machineUuid 都不能为空")
	}
}

func TestGetMachineID_ValidationError_EmptyMachineUUID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	_, err := svc.GetMachineID("", "", "testApp", "")
	if err == nil {
		t.Fatal("expected error for empty machineUUID")
	}
	if err.Error() != "appName 和 machineUuid 都不能为空" {
		t.Errorf("error = %q, want %q", err.Error(), "appName 和 machineUuid 都不能为空")
	}
}

func TestGetMachineID_ValidationError_WhitespaceOnly(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	_, err := svc.GetMachineID("", "", "  ", "  ")
	if err == nil {
		t.Fatal("expected error for whitespace-only params")
	}
}

func TestGetMachineID_FirstMachine(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// First machine should get ID = 1 (IFNULL(MAX(null), 0) + 1 = 1)
	id, err := svc.GetMachineID("", "", "testApp", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("machineID = %d, want 1 (first machine)", id)
	}
}

func TestGetMachineID_SecondMachine(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// First machine
	id1, err := svc.GetMachineID("", "", "testApp", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second machine (different UUID)
	id2, err := svc.GetMachineID("", "", "testApp", "192.168.1.2:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 != 1 {
		t.Errorf("first machineID = %d, want 1", id1)
	}
	if id2 != 2 {
		t.Errorf("second machineID = %d, want 2", id2)
	}
}

func TestGetMachineID_DuplicateRegistration(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// Register first
	id1, err := svc.GetMachineID("", "", "testApp", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Register same machine again - should return the same ID
	id2, err := svc.GetMachineID("", "", "testApp", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 != id2 {
		t.Errorf("duplicate registration returned different IDs: %d vs %d", id1, id2)
	}
}

func TestGetMachineID_DefaultNamespaceAndGroup(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// Register with empty namespace and groupID
	_, err := svc.GetMachineID("", "", "testApp", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the app was created with defaults
	var app model.App
	result := db.Where("name = ?", "testApp").First(&app)
	if result.Error != nil {
		t.Fatalf("failed to find app: %v", result.Error)
	}
	if app.Namespace != "PUBLIC" {
		t.Errorf("namespace = %q, want %q", app.Namespace, "PUBLIC")
	}
	if app.GroupID != "PUBLIC" {
		t.Errorf("groupID = %q, want %q", app.GroupID, "PUBLIC")
	}
}

func TestGetMachineID_CustomNamespaceAndGroup(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	_, err := svc.GetMachineID("prod", "group1", "testApp", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var app model.App
	result := db.Where("name = ? AND namespace = ?", "testApp", "prod").First(&app)
	if result.Error != nil {
		t.Fatalf("failed to find app: %v", result.Error)
	}
	if app.GroupID != "group1" {
		t.Errorf("groupID = %q, want %q", app.GroupID, "group1")
	}
}

func TestGetMachineID_DifferentApps(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// Register machine for app1
	id1, err := svc.GetMachineID("", "", "app1", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Register machine for app2 - should start from 1 independently
	id2, err := svc.GetMachineID("", "", "app2", "192.168.1.1:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 != 1 {
		t.Errorf("app1 first machineID = %d, want 1", id1)
	}
	if id2 != 1 {
		t.Errorf("app2 first machineID = %d, want 1 (independent per app)", id2)
	}
}

func TestGetMachineID_MultipleSequentialRegistrations(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	for i := 1; i <= 10; i++ {
		uuid := fmt.Sprintf("machine-%d", i)
		id, err := svc.GetMachineID("", "", "testApp", uuid)
		if err != nil {
			t.Fatalf("unexpected error at iteration %d: %v", i, err)
		}
		if id != i {
			t.Errorf("iteration %d: machineID = %d, want %d", i, id, i)
		}
	}
}

func TestListApps(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// Create some apps via registration
	_, _ = svc.GetMachineID("", "", "app1", "m1")
	_, _ = svc.GetMachineID("", "", "app2", "m2")

	apps, err := svc.ListApps()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Errorf("app count = %d, want 2", len(apps))
	}
}

func TestListMachines(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	// Register machines
	_, _ = svc.GetMachineID("", "", "testApp", "m1")
	_, _ = svc.GetMachineID("", "", "testApp", "m2")
	_, _ = svc.GetMachineID("", "", "testApp", "m3")

	// Find app
	var app model.App
	db.Where("name = ?", "testApp").First(&app)

	machines, err := svc.ListMachines(app.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(machines) != 3 {
		t.Errorf("machine count = %d, want 3", len(machines))
	}
}

// Verify the error type matches Java's IllegalArgumentException message
func TestGetMachineID_ErrorMessageMatchesJava(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewFastIdRepo(db)
	svc := NewFastIdService(repo)

	tests := []struct {
		name        string
		appName     string
		machineUUID string
		wantErr     bool
		errMsg      string
	}{
		{"both empty", "", "", true, "appName 和 machineUuid 都不能为空"},
		{"appName empty", "", "uuid", true, "appName 和 machineUuid 都不能为空"},
		{"machineUuid empty", "app", "", true, "appName 和 machineUuid 都不能为空"},
		{"both present", "app", "uuid", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetMachineID("", "", tt.appName, tt.machineUUID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !errors.Is(err, err) || err.Error() != tt.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
