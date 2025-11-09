package models

import "time"

type User struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:100;not null" json:"name"`
	Number           string    `gorm:"size:20;uniqueIndex;not null" json:"number"`
	Password         string    `gorm:"size:255;not null" json:"-"`
	ReffCode         string    `gorm:"size:20;uniqueIndex;not null" json:"reff_code"`
	ReffBy           *uint     `gorm:"column:reff_by" json:"reff_by"`
	Balance          float64   `gorm:"type:decimal(15,2);default:0" json:"balance"`
	Income           float64   `gorm:"type:decimal(15,2);default:0" json:"income"`
	Level            *uint     `gorm:"column:level;default:0" json:"level"`
	TotalInvest      float64   `gorm:"column:total_invest;type:decimal(15,2);default:0" json:"total_invest"`
	TotalInvestVIP   float64   `gorm:"column:total_invest_vip;type:decimal(15,2);default:0" json:"total_invest_vip"`
	SpinTicket       *uint     `gorm:"column:spin_ticket;default:0" json:"spin_ticket"`
	Status           string    `gorm:"type:enum('Active','Inactive','Suspend');default:'Active'" json:"status"`
	InvestmentStatus string    `gorm:"type:enum('Active','Inactive');default:'Inactive'" json:"investment_status"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

func (User) TableName() string {
	return "users"
}
