package handlers

import (
	"NotesAPI/internal/models"
	"NotesAPI/internal/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestUserHandler_Register(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := repository.NewUserRepository(testDB)
	handler := NewUserHandler(repo, testJWTConfig)

	tests := []struct {
		name           string
		input          map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful registration",
			input: map[string]interface{}{
				"email":    "newuser@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "duplicate email",
			input: map[string]interface{}{
				"email":    "existing@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "invalid email format",
			input: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			input: map[string]interface{}{
				"email":    "shortpass@example.com",
				"password": "12345",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "duplicate email" {
				testDB.Create(&models.User{
					Email:    "existing@example.com",
					Password: "hashedpassword",
				})
			}

			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			handler.Register(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("UserHandler.Register() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response TokenResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.Token == "" {
					t.Error("Expected token in response")
				}
				if response.User.ID == 0 {
					t.Error("Expected user ID in response")
				}
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := repository.NewUserRepository(testDB)
	handler := NewUserHandler(repo, testJWTConfig)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	testUser := &models.User{
		Email:    "loginuser@example.com",
		Password: string(hashedPassword),
	}
	testDB.Create(testUser)

	tests := []struct {
		name           string
		input          map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful login",
			input: map[string]interface{}{
				"email":    "loginuser@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email",
			input: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid password",
			input: map[string]interface{}{
				"email":    "loginuser@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid email format",
			input: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			handler.Login(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("UserHandler.Login() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response TokenResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.Token == "" {
					t.Error("Expected token in response")
				}
				if response.User.Email != "loginuser@example.com" {
					t.Errorf("Expected user email loginuser@example.com, got %s", response.User.Email)
				}
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := repository.NewUserRepository(testDB)
	handler := NewUserHandler(repo, testJWTConfig)

	testUser := &models.User{
		Email:    "getuser@example.com",
		Password: "hashedpassword",
	}
	testDB.Create(testUser)

	tests := []struct {
		name           string
		userID         string
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "successful get user",
			userID:         strconv.FormatUint(uint64(testUser.ID), 10),
			setContextUser: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user id format",
			userID:         "invalid",
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "user not found",
			userID:         "99999",
			setContextUser: true,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.userID}}

			// Simulate JWT middleware setting userID in context
			if tt.setContextUser {
				ctx.Set("userID", testUser.ID)
			}

			handler.GetUser(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("UserHandler.GetUser() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
