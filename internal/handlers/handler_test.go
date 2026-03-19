package handlers

import (
	"NotesAPI/internal/models"
	"NotesAPI/internal/utils"
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

var testJWTConfig = utils.JWTConfig{
	SecretKey:      "test-secret-key-for-testing-only",
	AccessExpHours: 24,
}

func setupTestDB(t *testing.T) func() {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	dsn := "host=" + host + " port=" + port.Port() + " user=test password=test dbname=test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Note{}); err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to migrate database: %v", err)
	}

	testDB = db

	return func() {
		container.Terminate(ctx)
	}
}

func generateTestToken(t *testing.T, userID uint) string {
	token, err := utils.GenerateToken(userID, testJWTConfig)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	return token
}
