package models

import "time"

type Deposit struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null" json:"user_id"`
	User           *User     `gorm:"foreignKey:UserID" json:"-"`
	Amount         float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	OrderID        string    `gorm:"type:varchar(191);uniqueIndex;not null" json:"order_id"`
	PaymentMethod  string    `gorm:"type:enum('QRIS','BANK');not null" json:"payment_method"`
	PaymentChannel *string   `gorm:"type:enum('BCA','BRI','BNI','MANDIRI','PERMATA','BNC')" json:"payment_channel,omitempty"`
	PaymentCode    *string   `gorm:"type:text" json:"payment_code,omitempty"`
	Status         string    `gorm:"type:enum('Success','Pending','Failed');default:'Pending'" json:"status"`
	ExpiredAt      time.Time `gorm:"not null" json:"expired_at"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

func (Deposit) TableName() string {
	return "deposits"
}
