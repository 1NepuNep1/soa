package models

import (
	"time"
)

type User struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	Login        string      `gorm:"unique;not null" json:"login"`
	PasswordHash string      `gorm:"not null" json:"-"`
	Email        string      `gorm:"unique;not null" json:"email"`
	Status       string      `gorm:"default:'active'" json:"status"`
	Profile      UserProfile `gorm:"constraint:OnDelete:CASCADE" json:"profile"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type UserProfile struct {
	ProfileID   uint       `gorm:"primaryKey" json:"profileId"`
	UserID      uint       `gorm:"uniqueIndex" json:"userId"`
	FirstName   string     `json:"firstName"`
	LastName    string     `json:"lastName"`
	BirthDate   *time.Time `json:"birthDate,omitempty"`
	PhoneNumber string     `json:"phoneNumber"`
	Bio         string     `json:"bio"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Session struct {
	SessionID uint      `gorm:"primaryKey" json:"sessionId"`
	UserID    uint      `gorm:"index" json:"userId"`
	Token     string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
