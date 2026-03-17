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
)

func setupNoteHandlerWithAuth(t *testing.T) (*NoteHandler, string, *models.User) {
	repo := repository.NewNoteRepository(handlerDB)
	handler := NewNoteHandler(repo)
	
	// Create a test user
	testUser := &models.User{
		Email:    "testuser@example.com",
		Password: "hashedpassword",
	}
	if err := handlerDB.Create(testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	token := generateTestToken(t, testUser.ID)
	return handler, token, testUser
}

func TestNoteHandler_CreateNote(t *testing.T) {
	teardown := setupHandlerTestDB(t)
	defer teardown()

	handler, token, testUser := setupNoteHandlerWithAuth(t)

	tests := []struct {
		name           string
		input          map[string]interface{}
		token          string
		setContextUser bool
		expectedStatus int
	}{
		{
			name: "successful create note",
			input: map[string]interface{}{
				"title":   "Test Note",
				"content": "Test content",
			},
			token:          token,
			setContextUser: true,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing title",
			input: map[string]interface{}{
				"content": "Test content",
			},
			token:          token,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing content",
			input: map[string]interface{}{
				"title": "Test Note",
			},
			token:          token,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing token",
			input: map[string]interface{}{
				"title":   "Test Note",
				"content": "Test content",
			},
			token:          "",
			setContextUser: false,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid token",
			input: map[string]interface{}{
				"title":   "Test Note",
				"content": "Test content",
			},
			token:          "invalid-token",
			setContextUser: false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPost, "/notes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Set userID in context (simulating JWT middleware)
			if tt.setContextUser {
				ctx.Set("userID", testUser.ID)
			}

			handler.CreateNote(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("NoteHandler.CreateNote() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNoteHandler_GetNotes(t *testing.T) {
	teardown := setupHandlerTestDB(t)
	defer teardown()

	handler, token, _ := setupNoteHandlerWithAuth(t)

	user1 := &models.User{
		Email:    "getnotesuser1@example.com",
		Password: "hashedpassword",
	}
	user2 := &models.User{
		Email:    "getnotesuser2@example.com",
		Password: "hashedpassword",
	}
	handlerDB.Create(user1)
	handlerDB.Create(user2)

	notes := []*models.Note{
		{Title: "Note 1", Content: "Content 1", UserID: user1.ID},
		{Title: "Note 2", Content: "Content 2", UserID: user1.ID},
	}
	for _, note := range notes {
		handlerDB.Create(note)
	}

	tests := []struct {
		name           string
		token          string
		userID         uint
		setContextUser bool
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get notes for user with notes",
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "get notes for user without notes",
			token:          generateTestToken(t, user2.ID),
			userID:         user2.ID,
			setContextUser: true,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "missing token",
			token:          "",
			userID:         user1.ID,
			setContextUser: false,
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Set userID in context (simulating JWT middleware)
			if tt.setContextUser {
				ctx.Set("userID", tt.userID)
			}

			handler.GetNotes(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("NoteHandler.GetNotes() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response []models.Note
				json.Unmarshal(w.Body.Bytes(), &response)
				if len(response) != tt.expectedCount {
					t.Errorf("NoteHandler.GetNotes() count = %v, want %v", len(response), tt.expectedCount)
				}
			}
		})
	}
}

func TestNoteHandler_GetNote(t *testing.T) {
	teardown := setupHandlerTestDB(t)
	defer teardown()

	handler, token, _ := setupNoteHandlerWithAuth(t)

	user1 := &models.User{
		Email:    "getnoteuser1@example.com",
		Password: "hashedpassword",
	}
	user2 := &models.User{
		Email:    "getnoteuser2@example.com",
		Password: "hashedpassword",
	}
	handlerDB.Create(user1)
	handlerDB.Create(user2)

	testNote := &models.Note{
		Title:   "Test Note",
		Content: "Test Content",
		UserID:  user1.ID,
	}
	handlerDB.Create(testNote)

	tests := []struct {
		name           string
		noteID         string
		token          string
		userID         uint
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "successful get note",
			noteID:         strconv.FormatUint(uint64(testNote.ID), 10),
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "note not owned by user",
			noteID:         strconv.FormatUint(uint64(testNote.ID), 10),
			token:          generateTestToken(t, user2.ID),
			userID:         user2.ID,
			setContextUser: true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid note id format",
			noteID:         "invalid",
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existing note",
			noteID:         "99999",
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing token",
			noteID:         strconv.FormatUint(uint64(testNote.ID), 10),
			token:          "",
			userID:         user1.ID,
			setContextUser: false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/notes/"+tt.noteID, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.noteID}}

			// Set userID in context (simulating JWT middleware)
			if tt.setContextUser {
				ctx.Set("userID", tt.userID)
			}

			handler.GetNote(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("NoteHandler.GetNote() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNoteHandler_UpdateNote(t *testing.T) {
	teardown := setupHandlerTestDB(t)
	defer teardown()

	handler, token, _ := setupNoteHandlerWithAuth(t)

	user1 := &models.User{
		Email:    "updatenoteuser1@example.com",
		Password: "hashedpassword",
	}
	user2 := &models.User{
		Email:    "updatenoteuser2@example.com",
		Password: "hashedpassword",
	}
	handlerDB.Create(user1)
	handlerDB.Create(user2)

	testNote := &models.Note{
		Title:   "Original Title",
		Content: "Original Content",
		UserID:  user1.ID,
	}
	handlerDB.Create(testNote)

	tests := []struct {
		name           string
		noteID         string
		token          string
		userID         uint
		setContextUser bool
		input          map[string]interface{}
		expectedStatus int
	}{
		{
			name:   "successful update note",
			noteID: strconv.FormatUint(uint64(testNote.ID), 10),
			token:  token,
			userID: user1.ID,
			setContextUser: true,
			input: map[string]interface{}{
				"title":   "Updated Title",
				"content": "Updated Content",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "note not owned by user",
			noteID: strconv.FormatUint(uint64(testNote.ID), 10),
			token:  generateTestToken(t, user2.ID),
			userID: user2.ID,
			setContextUser: true,
			input: map[string]interface{}{
				"title":   "Updated Title",
				"content": "Updated Content",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "missing title",
			noteID: strconv.FormatUint(uint64(testNote.ID), 10),
			token:  token,
			userID: user1.ID,
			setContextUser: true,
			input: map[string]interface{}{
				"content": "Updated Content",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid note id format",
			noteID: "invalid",
			token:  token,
			userID: user1.ID,
			setContextUser: true,
			input: map[string]interface{}{
				"title":   "Updated Title",
				"content": "Updated Content",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "non-existing note",
			noteID: "99999",
			token:  token,
			userID: user1.ID,
			setContextUser: true,
			input: map[string]interface{}{
				"title":   "Updated Title",
				"content": "Updated Content",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing token",
			noteID:         strconv.FormatUint(uint64(testNote.ID), 10),
			token:          "",
			userID:         user1.ID,
			setContextUser: false,
			input:          map[string]interface{}{"title": "Updated", "content": "Updated"},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest(http.MethodPut, "/notes/"+tt.noteID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.noteID}}

			// Set userID in context (simulating JWT middleware)
			if tt.setContextUser {
				ctx.Set("userID", tt.userID)
			}

			handler.UpdateNote(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("NoteHandler.UpdateNote() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNoteHandler_DeleteNote(t *testing.T) {
	teardown := setupHandlerTestDB(t)
	defer teardown()

	handler, token, _ := setupNoteHandlerWithAuth(t)

	user1 := &models.User{
		Email:    "deletenoteuser1@example.com",
		Password: "hashedpassword",
	}
	user2 := &models.User{
		Email:    "deletenoteuser2@example.com",
		Password: "hashedpassword",
	}
	handlerDB.Create(user1)
	handlerDB.Create(user2)

	testNote := &models.Note{
		Title:   "Note to delete",
		Content: "Content to delete",
		UserID:  user1.ID,
	}
	handlerDB.Create(testNote)

	tests := []struct {
		name           string
		noteID         string
		token          string
		userID         uint
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "successful delete note",
			noteID:         strconv.FormatUint(uint64(testNote.ID), 10),
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "note not owned by user",
			noteID:         "99998",
			token:          generateTestToken(t, user2.ID),
			userID:         user2.ID,
			setContextUser: true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid note id format",
			noteID:         "invalid",
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existing note",
			noteID:         "99999",
			token:          token,
			userID:         user1.ID,
			setContextUser: true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing token",
			noteID:         strconv.FormatUint(uint64(testNote.ID), 10),
			token:          "",
			userID:         user1.ID,
			setContextUser: false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, "/notes/"+tt.noteID, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.noteID}}

			// Set userID in context (simulating JWT middleware)
			if tt.setContextUser {
				ctx.Set("userID", tt.userID)
			}

			handler.DeleteNote(ctx)

			if w.Code != tt.expectedStatus {
				t.Errorf("NoteHandler.DeleteNote() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
