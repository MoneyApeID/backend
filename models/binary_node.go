package models

import "time"

// BinaryNode menyimpan struktur binary kiri-kanan untuk setiap user
type BinaryNode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"-"`
	LeftID    *uint     `gorm:"column:left_id" json:"left_id"`    // User ID di kiri
	RightID   *uint     `gorm:"column:right_id" json:"right_id"` // User ID di kanan
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (BinaryNode) TableName() string {
	return "binary_nodes"
}

