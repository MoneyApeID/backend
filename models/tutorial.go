package models

import "time"

type Tutorial struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Image     string    `gorm:"type:text;not null" json:"image"` // Filename only (e.g., "uuid.jpg"), not full URL. Frontend will construct the full URL.
	Link      string    `gorm:"type:text;not null" json:"link"`
	Status    string    `gorm:"type:enum('Active','Inactive');default:'Active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Tutorial) TableName() string {
	return "tutorials"
}

