package models

import "time"

type Forum struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	WithdrawalID *uint     `gorm:"column:withdrawal_id" json:"withdrawal_id,omitempty"`
	Reward       float64   `gorm:"type:decimal(15,2);default:0" json:"reward"`
	Description  string    `gorm:"type:varchar(60);not null" json:"description"`
	Image        string    `gorm:"type:varchar(255);not null" json:"image"`
	Status       string    `gorm:"type:enum('Accepted','Pending','Rejected');default:'Pending'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
