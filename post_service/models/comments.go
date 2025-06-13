package models

import (
	"time"
)

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	PostID    uint32 `gorm:"not null"`
	ClientID  uint32 `gorm:"not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
}
