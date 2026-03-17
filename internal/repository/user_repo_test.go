package repository

import (
	"NotesAPI/internal/models"
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

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

func TestUserRepository_Create(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(testDB)

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "successful create",
			user: &models.User{
				Email:    "test@example.com",
				Password: "hashedpassword123",
			},
			wantErr: false,
		},
		{
			name: "duplicate email should fail",
			user: &models.User{
				Email:    "test@example.com",
				Password: "anotherpassword",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(testDB)

	user := &models.User{
		Email:    "getbyemail@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "existing user",
			email:   "getbyemail@example.com",
			wantErr: false,
		},
		{
			name:    "non-existing user",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := repo.GetByEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && foundUser == nil {
				t.Error("UserRepository.GetByEmail() returned nil user")
			}
			if !tt.wantErr && foundUser.Email != tt.email {
				t.Errorf("UserRepository.GetByEmail() email = %v, want %v", foundUser.Email, tt.email)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(testDB)

	user := &models.User{
		Email:    "getbyid@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && foundUser == nil {
				t.Error("UserRepository.GetByID() returned nil user")
			}
			if !tt.wantErr && foundUser.ID != tt.id {
				t.Errorf("UserRepository.GetByID() id = %v, want %v", foundUser.ID, tt.id)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(testDB)

	user := &models.User{
		Email:    "update@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name      string
		updateFn  func(*models.User) *models.User
		wantErr   bool
		verifyFn  func(*models.User) bool
	}{
		{
			name: "update email",
			updateFn: func(u *models.User) *models.User {
				u.Email = "updated@example.com"
				return u
			},
			wantErr: false,
			verifyFn: func(u *models.User) bool {
				return u.Email == "updated@example.com"
			},
		},
		{
			name: "update password",
			updateFn: func(u *models.User) *models.User {
				u.Password = "newhashedpassword"
				return u
			},
			wantErr: false,
			verifyFn: func(u *models.User) bool {
				return u.Password == "newhashedpassword"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedUser := tt.updateFn(user)
			err := repo.Update(updatedUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				fetchedUser, _ := repo.GetByID(user.ID)
				if !tt.verifyFn(fetchedUser) {
					t.Error("UserRepository.Update() failed to update user correctly")
				}
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(testDB)

	user := &models.User{
		Email:    "delete@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existing user",
			id:      99999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			_, err = repo.GetByID(tt.id)
			if err == nil && tt.id == user.ID {
				t.Error("UserRepository.Delete() user still exists after deletion")
			}
		})
	}
}
