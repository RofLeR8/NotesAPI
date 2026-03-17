package repository

import (
	"NotesAPI/internal/models"
	"testing"
)

func TestNoteRepository_Create(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewNoteRepository(testDB)

	user := &models.User{
		Email:    "noteuser@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		note    *models.Note
		wantErr bool
	}{
		{
			name: "successful create",
			note: &models.Note{
				Title:   "Test Note",
				Content: "Test content",
				UserID:  user.ID,
			},
			wantErr: false,
		},
		{
			name: "create note with empty content",
			note: &models.Note{
				Title:   "Empty Note",
				Content: "",
				UserID:  user.ID,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.note)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.note.ID == 0 {
				t.Error("NoteRepository.Create() note ID not set after creation")
			}
		})
	}
}

func TestNoteRepository_GetByID(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewNoteRepository(testDB)

	user := &models.User{
		Email:    "getbyiduser@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	note := &models.Note{
		Title:   "Get By ID Note",
		Content: "Content for get by ID test",
		UserID:  user.ID,
	}
	if err := testDB.Create(note).Error; err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "existing note",
			id:      note.ID,
			wantErr: false,
		},
		{
			name:    "non-existing note",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundNote, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteRepository.GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && foundNote == nil {
				t.Error("NoteRepository.GetByID() returned nil note")
			}
			if !tt.wantErr && foundNote.ID != tt.id {
				t.Errorf("NoteRepository.GetByID() id = %v, want %v", foundNote.ID, tt.id)
			}
		})
	}
}

func TestNoteRepository_GetAllByUserID(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewNoteRepository(testDB)

	user1 := &models.User{
		Email:    "user1@example.com",
		Password: "hashedpassword",
	}
	user2 := &models.User{
		Email:    "user2@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}
	if err := testDB.Create(user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	notes := []*models.Note{
		{Title: "Note 1", Content: "Content 1", UserID: user1.ID},
		{Title: "Note 2", Content: "Content 2", UserID: user1.ID},
		{Title: "Note 3", Content: "Content 3", UserID: user1.ID},
	}
	for _, note := range notes {
		if err := testDB.Create(note).Error; err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}
	}

	if err := testDB.Create(&models.Note{Title: "User2 Note", Content: "User2 Content", UserID: user2.ID}).Error; err != nil {
		t.Fatalf("Failed to create test note for user2: %v", err)
	}

	tests := []struct {
		name          string
		userID        uint
		expectedCount int
	}{
		{
			name:          "user with multiple notes",
			userID:        user1.ID,
			expectedCount: 3,
		},
		{
			name:          "user with single note",
			userID:        user2.ID,
			expectedCount: 1,
		},
		{
			name:          "user with no notes",
			userID:        99999,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundNotes, err := repo.GetAllByUserID(tt.userID)
			if err != nil {
				t.Errorf("NoteRepository.GetAllByUserID() error = %v", err)
			}
			if len(foundNotes) != tt.expectedCount {
				t.Errorf("NoteRepository.GetAllByUserID() count = %v, want %v", len(foundNotes), tt.expectedCount)
			}
		})
	}
}

func TestNoteRepository_Update(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewNoteRepository(testDB)

	user := &models.User{
		Email:    "updateuser@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	note := &models.Note{
		Title:   "Original Title",
		Content: "Original Content",
		UserID:  user.ID,
	}
	if err := testDB.Create(note).Error; err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	tests := []struct {
		name      string
		updateFn  func(*models.Note) *models.Note
		wantErr   bool
		verifyFn  func(*models.Note) bool
	}{
		{
			name: "update title",
			updateFn: func(n *models.Note) *models.Note {
				n.Title = "Updated Title"
				return n
			},
			wantErr: false,
			verifyFn: func(n *models.Note) bool {
				return n.Title == "Updated Title"
			},
		},
		{
			name: "update content",
			updateFn: func(n *models.Note) *models.Note {
				n.Content = "Updated Content"
				return n
			},
			wantErr: false,
			verifyFn: func(n *models.Note) bool {
				return n.Content == "Updated Content"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedNote := tt.updateFn(note)
			err := repo.Update(updatedNote)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				fetchedNote, _ := repo.GetByID(note.ID)
				if !tt.verifyFn(fetchedNote) {
					t.Error("NoteRepository.Update() failed to update note correctly")
				}
			}
		})
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewNoteRepository(testDB)

	user := &models.User{
		Email:    "deleteuser@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	note := &models.Note{
		Title:   "Note to delete",
		Content: "Content to delete",
		UserID:  user.ID,
	}
	if err := testDB.Create(note).Error; err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing note",
			id:      note.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existing note",
			id:      99999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			_, err = repo.GetByID(tt.id)
			if err == nil && tt.id == note.ID {
				t.Error("NoteRepository.Delete() note still exists after deletion")
			}
		})
	}
}

func TestNoteRepository_GetByIDAndUserID(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()

	repo := NewNoteRepository(testDB)

	user1 := &models.User{
		Email:    "owner@example.com",
		Password: "hashedpassword",
	}
	user2 := &models.User{
		Email:    "other@example.com",
		Password: "hashedpassword",
	}
	if err := testDB.Create(user1).Error; err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}
	if err := testDB.Create(user2).Error; err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	note := &models.Note{
		Title:   "Owned Note",
		Content: "Owned Content",
		UserID:  user1.ID,
	}
	if err := testDB.Create(note).Error; err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		userID  uint
		wantErr bool
	}{
		{
			name:    "note owned by user",
			id:      note.ID,
			userID:  user1.ID,
			wantErr: false,
		},
		{
			name:    "note not owned by user",
			id:      note.ID,
			userID:  user2.ID,
			wantErr: true,
		},
		{
			name:    "non-existing note",
			id:      99999,
			userID:  user1.ID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundNote, err := repo.GetByIDAndUserID(tt.id, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteRepository.GetByIDAndUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && foundNote == nil {
				t.Error("NoteRepository.GetByIDAndUserID() returned nil note")
			}
			if !tt.wantErr && foundNote.UserID != tt.userID {
				t.Errorf("NoteRepository.GetByIDAndUserID() user mismatch")
			}
		})
	}
}
