package models

import "time"

type User struct {
	ID        uint      `gorm:"uniqueIndex;not null" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	Notes     []Note    `gorm:"foreignKey:UserID" json:"notes,omitempty"`
}
