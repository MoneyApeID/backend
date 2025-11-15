package utils

import (
	"project/database"
	"project/models"
	"time"

	"gorm.io/gorm"
)

// InitializeRewardProgress menginisialisasi reward progress untuk user
// Dipanggil saat user pertama kali memiliki investasi aktif
func InitializeRewardProgress(userID uint) error {
	db := database.DB

	// Get all active rewards
	var rewards []models.Reward
	if err := db.Where("status = ?", "Active").Find(&rewards).Error; err != nil {
		return err
	}

	now := time.Now()

	// Create reward progress for each reward
	for _, reward := range rewards {
		// Check if progress already exists
		var existing models.RewardProgress
		if err := db.Where("user_id = ? AND reward_id = ?", userID, reward.ID).First(&existing).Error; err == nil {
			// Already exists, skip
			continue
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// Calculate expires_at based on duration
		var expiresAt *time.Time
		if !reward.IsAccumulative {
			// Reset reward: set expires_at
			exp := now.Add(time.Duration(reward.Duration) * 24 * time.Hour)
			expiresAt = &exp
		}
		// Accumulative reward: expires_at = nil (tidak pernah expire)

		progress := models.RewardProgress{
			UserID:      userID,
			RewardID:    reward.ID,
			OmsetLeft:   0,
			OmsetRight:  0,
			TotalOmset:  0,
			IsCompleted: false,
			IsClaimed:   false,
			StartedAt:   now,
			ExpiresAt:   expiresAt,
		}

		if err := db.Create(&progress).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateRewardProgress mengupdate progress reward untuk user tertentu
// Dipanggil setelah investasi dibuat atau setelah cron daily returns
func UpdateRewardProgress(userID uint) error {
	db := database.DB

	// Hitung omset dari level 1-3
	omsetLeft, omsetRight, totalOmset, err := CalculateOmset(userID)
	if err != nil {
		return err
	}

	// Get all active rewards
	var rewards []models.Reward
	if err := db.Where("status = ?", "Active").Find(&rewards).Error; err != nil {
		return err
	}

	now := time.Now()

	// Update progress for each reward
	for _, reward := range rewards {
		var progress models.RewardProgress
		if err := db.Where("user_id = ? AND reward_id = ?", userID, reward.ID).First(&progress).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Initialize if not exists
				var expiresAt *time.Time
				if !reward.IsAccumulative {
					exp := now.Add(time.Duration(reward.Duration) * 24 * time.Hour)
					expiresAt = &exp
				}
				progress = models.RewardProgress{
					UserID:      userID,
					RewardID:    reward.ID,
					StartedAt:   now,
					ExpiresAt:   expiresAt,
				}
			} else {
				continue
			}
		}

		// Check if expired (for reset rewards)
		if !reward.IsAccumulative && progress.ExpiresAt != nil {
			if now.After(*progress.ExpiresAt) {
				// Reset progress - set omset menjadi 0
				progress.OmsetLeft = 0
				progress.OmsetRight = 0
				progress.TotalOmset = 0
				progress.IsCompleted = false
				progress.IsClaimed = false
				progress.StartedAt = now
				progress.LastResetAt = &now
				// Set new expires_at
				exp := now.Add(time.Duration(reward.Duration) * 24 * time.Hour)
				progress.ExpiresAt = &exp
				// Save reset progress dulu
				if err := db.Save(&progress).Error; err != nil {
					return err
				}
				// Skip update omset karena sudah di-reset
				continue
			}
		}

		// Jika reward sudah di-claim, omset harus tetap 0 sampai periode baru dimulai
		// Setelah claim, omset di-reset menjadi 0 dan IsClaimed = true
		// Ketika UpdateRewardProgress dipanggil, kita perlu memastikan bahwa omset tetap 0
		// sampai IsClaimed di-reset menjadi false (saat expired atau manual reset)
		// Tapi sebenarnya, setelah claim, omset harus mulai dihitung ulang dari aktivitas baru
		// Jadi kita perlu memastikan bahwa omset dihitung dari aktivitas setelah StartedAt
		// Untuk saat ini, kita skip update omset jika IsClaimed = true untuk mempertahankan nilai 0
		// sampai admin atau sistem mereset IsClaimed menjadi false
		if progress.IsClaimed {
			// Reward sudah di-claim, omset harus tetap 0
			// Jangan update omset sampai IsClaimed di-reset menjadi false
			// Ini memastikan bahwa setelah claim, omset tetap 0 sampai periode baru dimulai
			continue
		}

		// Update omset (hanya jika tidak expired atau accumulative, dan belum di-claim)
		progress.OmsetLeft = omsetLeft
		progress.OmsetRight = omsetRight
		progress.TotalOmset = totalOmset

		// Check if completed
		if progress.TotalOmset >= reward.OmsetTarget {
			progress.IsCompleted = true
		}

		// Save progress
		if err := db.Save(&progress).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateAllUsersRewardProgress mengupdate reward progress untuk semua user yang memiliki investasi aktif
// Dipanggil dari cron job
func UpdateAllUsersRewardProgress() error {
	db := database.DB

	// Get all users with active investments
	var users []models.User
	if err := db.Where("investment_status = ?", "Active").Find(&users).Error; err != nil {
		return err
	}

	// Update progress for each user
	for _, user := range users {
		if err := UpdateRewardProgress(user.ID); err != nil {
			// Log error but continue with other users
			continue
		}
	}

	return nil
}

