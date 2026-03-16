package models

import "time"

type AdminWithdrawal struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AdminID       uint      `gorm:"not null;index" json:"admin_id"`
	BankCode      string    `gorm:"type:varchar(20);not null" json:"bank_code"`
	AccountNumber string    `gorm:"type:varchar(50);not null" json:"account_number"`
	AccountName   string    `gorm:"type:varchar(255)" json:"account_name"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	AdminFee      float64   `gorm:"type:decimal(15,2);not null;default:2000" json:"admin_fee"`
	FinalAmount   float64   `gorm:"type:decimal(15,2);not null" json:"final_amount"`
	OrderID       string    `gorm:"type:varchar(191);not null;uniqueIndex" json:"order_id"`
	SessionID     string    `gorm:"type:varchar(255)" json:"session_id"`
	Status        string    `gorm:"type:enum('Success','Pending','Failed');not null;default:'Pending'" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (AdminWithdrawal) TableName() string {
	return "admin_withdrawals"
}
