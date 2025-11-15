package models

import "time"

type Setting struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"type:text;not null" json:"name"`
	Company        string    `gorm:"type:text;not null" json:"company"`
	Popup          string    `gorm:"type:text" json:"popup"` // Filename only (e.g., "popup.png"), not full URL. Frontend will construct the full URL.
	PopupTitle     string    `gorm:"type:varchar(255)" json:"popup_title"`
	MinWithdraw    float64   `gorm:"type:decimal(15,2);not null" json:"min_withdraw"`
	MaxWithdraw    float64   `gorm:"type:decimal(15,2);not null" json:"max_withdraw"`
	WithdrawCharge float64   `gorm:"type:decimal(15,2);not null" json:"withdraw_charge"`
	AutoWithdraw   bool      `gorm:"default:0" json:"auto_withdraw"`
	Maintenance    bool      `gorm:"default:0" json:"maintenance"`
	ClosedRegister bool      `gorm:"default:0" json:"closed_register"`
	LinkCS         string    `gorm:"type:text;not null" json:"link_cs"`
	LinkGroup      string    `gorm:"type:text;not null" json:"link_group"`
	LinkApp        string    `gorm:"type:text;not null" json:"link_app"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}
