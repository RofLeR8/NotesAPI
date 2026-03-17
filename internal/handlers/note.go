package handlers

import (
	"NotesAPI/internal/middleware"
	"NotesAPI/internal/models"
	"NotesAPI/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	repo *repository.NoteRepository
}

func NewNoteHandler(repo *repository.NoteRepository) *NoteHandler {
	return &NoteHandler{repo: repo}
}

type CreateNoteInput struct {
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"`
}

// Создать заметку
func (h *NoteHandler) CreateNote(c *gin.Context) {
	userID, exists := middleware.UserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
	}
	var input CreateNoteInput
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	note := &models.Note{
		Title:   input.Title,
		Content: input.Content,
		UserID:  uint(userID),
	}
	if err := h.repo.Create(note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)

}

// Получить все заметки пользователя
func (h *NoteHandler) GetNotes(c *gin.Context) {
	userID, exists := middleware.UserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
	}
	notes, err := h.repo.GetAllByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notes"})
		return
	}
	c.JSON(http.StatusOK, notes)
}

// Получить одну заметку
func (h *NoteHandler) GetNote(c *gin.Context) {
	userID, exists := middleware.UserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
	}
	noteIDStr := c.Param("id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	note, err := h.repo.GetByIDAndUserID(uint(noteID), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

type UpdateNoteInput struct {
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"`
}

// Обновить заметку
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	userID, exists := middleware.UserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
	}
	noteIDStr := c.Param("id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	note, err := h.repo.GetByIDAndUserID(uint(noteID), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	var input UpdateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note.Title = input.Title
	note.Content = input.Content

	if err := h.repo.Update(note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// Удалить заметку
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	userID, exists := middleware.UserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not found in context"})
	}
	noteIDStr := c.Param("id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	_, err = h.repo.GetByIDAndUserID(uint(noteID), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	if err := h.repo.Delete(uint(noteID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note deleted"})
}
