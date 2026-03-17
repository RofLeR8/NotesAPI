package repository

import (
	"NotesAPI/internal/models"
	"errors"

	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// Создать новую заметку
func (r *NoteRepository) Create(note *models.Note) error {
	return r.db.Create(note).Error
}

// Найти заметку по ID
func (r *NoteRepository) GetByID(id uint) (*models.Note, error) {
	var note models.Note
	err := r.db.First(&note, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("note not found")
		}
		return nil, err
	}
	return &note, nil
}

// Получить все заметки ползователя по ID
func (r *NoteRepository) GetAllByUserID(userID uint) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Where("user_id = ?", userID).Find(&notes).Error
	return notes, err
}

// Обновить заметку
func (r *NoteRepository) Update(note *models.Note) error {
	return r.db.Save(note).Error
}

// Удалить заметку
func (r *NoteRepository) Delete(id uint) error {
	return r.db.Delete(&models.Note{}, id).Error
}

// Найти заметку и проверить принадлежность пользователю
func (r *NoteRepository) GetByIDAndUserID(id, userID uint) (*models.Note, error) {
	var note models.Note
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("note not found")
		}
		return nil, err
	}
	return &note, nil
}
