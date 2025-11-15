package models

import "time"

// Reward menyimpan definisi reward yang tersedia
type Reward struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	OmsetTarget float64   `gorm:"type:decimal(15,2);not null" json:"omset_target"` // Target omset untuk mendapatkan reward
	RewardDesc  string    `gorm:"type:text" json:"reward_desc"`                    // Deskripsi reward (untuk manual distribution)
	Duration    int       `gorm:"not null" json:"duration"`                        // Durasi dalam hari
	IsAccumulative bool    `gorm:"default:0" json:"is_accumulative"`               // true = akumulasi, false = reset
	Status      string    `gorm:"type:enum('Active','Inactive');default:'Active'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Reward) TableName() string {
	return "rewards"
}

// RewardProgress menyimpan progress reward untuk setiap user
type RewardProgress struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	User         *User      `gorm:"foreignKey:UserID" json:"-"`
	RewardID     uint      `gorm:"not null;index" json:"reward_id"`
	Reward       *Reward   `gorm:"foreignKey:RewardID" json:"-"`
	OmsetLeft    float64   `gorm:"type:decimal(15,2);default:0" json:"omset_left"`   // Omset dari sisi kiri
	OmsetRight   float64   `gorm:"type:decimal(15,2);default:0" json:"omset_right"`   // Omset dari sisi kanan
	TotalOmset   float64   `gorm:"type:decimal(15,2);default:0" json:"total_omset"`   // Total omset (kiri + kanan)
	IsCompleted  bool      `gorm:"default:0" json:"is_completed"`                    // Apakah sudah mencapai target
	IsClaimed    bool      `gorm:"default:0" json:"is_claimed"`                      // Apakah sudah di-claim (manual)
	StartedAt    time.Time `gorm:"not null" json:"started_at"`                        // Kapan periode dimulai
	ExpiresAt     *time.Time `gorm:"index" json:"expires_at"`                         // Kapan periode berakhir (untuk reset)
	LastResetAt   *time.Time `json:"last_reset_at"`                                   // Kapan terakhir di-reset
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (RewardProgress) TableName() string {
	return "reward_progress"
}

