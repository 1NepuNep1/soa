package models

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	CreatorID   uint32         `gorm:"not null" json:"creatorId"`
	IsPrivate   bool           `gorm:"default:false" json:"isPrivate"`
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}
