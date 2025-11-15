package users

import (
	"net/http"
	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

// BinaryMember untuk detail anggota binary
type BinaryMember struct {
	UserID   uint    `json:"user_id"`
	Name     string  `json:"name"`
	Number   string  `json:"number"`
	Omset    float64 `json:"omset"`
	Position string  `json:"position"` // "left" atau "right"
}

// BinaryStructureResponse untuk response struktur binary
type BinaryStructureResponse struct {
	Root    BinaryMember   `json:"root"`    // User sendiri (level 0)
	Level1  []BinaryMember `json:"level1"`  // 2 anggota (left, right)
	Level2  []BinaryMember `json:"level2"`  // 4 anggota
	Level3  []BinaryMember `json:"level3"`  // 8 anggota
	OmsetLeft   float64 `json:"omset_left"`
	OmsetRight  float64 `json:"omset_right"`
	TotalOmset  float64 `json:"total_omset"`
}

// OmsetResponse untuk response omset
type OmsetResponse struct {
	OmsetLeft   float64 `json:"omset_left"`
	OmsetRight  float64 `json:"omset_right"`
	TotalOmset  float64 `json:"total_omset"`
	Level1Count int     `json:"level1_count"` // Jumlah member di level 1
	Level2Count int     `json:"level2_count"` // Jumlah member di level 2
	Level3Count int     `json:"level3_count"` // Jumlah member di level 3
}

// GET /api/users/binary/structure
func GetBinaryStructureHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB

	// Get user info (root)
	var rootUser models.User
	if err := db.Select("id, name, number").Where("id = ?", uid).First(&rootUser).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Calculate root omset (hanya dari investasi root sendiri, tanpa downline untuk display)
	rootOmset, _ := utils.CalculateUserOmsetOnly(uid)

	// Get binary node
	var binaryNode models.BinaryNode
	if err := db.Where("user_id = ?", uid).First(&binaryNode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User belum punya binary node, return empty structure
			response := BinaryStructureResponse{
				Root: BinaryMember{
					UserID: rootUser.ID,
					Name:   rootUser.Name,
					Number: rootUser.Number,
					Omset:  rootOmset,
				},
				Level1:      []BinaryMember{},
				Level2:      []BinaryMember{},
				Level3:      []BinaryMember{},
				OmsetLeft:   0,
				OmsetRight:  0,
				TotalOmset:  rootOmset,
			}
			utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
				Success: true,
				Message: "Successfully",
				Data:    response,
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Calculate total omset
	omsetLeft, omsetRight, totalOmset, _ := utils.CalculateOmset(uid)

	// Build structure per level dengan prioritas omset terbesar
	// Level 1: top 2 dengan omset terbesar
	// Level 2: top 4 dengan omset terbesar
	// Level 3: top 8 dengan omset terbesar
	level1Members := utils.GetTopMembersByOmset(uid, 1, 2)
	level2Members := utils.GetTopMembersByOmset(uid, 2, 4)
	level3Members := utils.GetTopMembersByOmset(uid, 3, 8)

	// Convert to BinaryMember
	level1 := make([]BinaryMember, len(level1Members))
	for i, m := range level1Members {
		level1[i] = BinaryMember{
			UserID:   m.UserID,
			Name:     m.Name,
			Number:   m.Number,
			Omset:    m.Omset,
			Position: m.Position,
		}
	}
	level2 := make([]BinaryMember, len(level2Members))
	for i, m := range level2Members {
		level2[i] = BinaryMember{
			UserID:   m.UserID,
			Name:     m.Name,
			Number:   m.Number,
			Omset:    m.Omset,
			Position: m.Position,
		}
	}
	level3 := make([]BinaryMember, len(level3Members))
	for i, m := range level3Members {
		level3[i] = BinaryMember{
			UserID:   m.UserID,
			Name:     m.Name,
			Number:   m.Number,
			Omset:    m.Omset,
			Position: m.Position,
		}
	}

	response := BinaryStructureResponse{
		Root: BinaryMember{
			UserID: rootUser.ID,
			Name:   rootUser.Name,
			Number: rootUser.Number,
			Omset:  rootOmset,
		},
		Level1:      level1,
		Level2:      level2,
		Level3:      level3,
		OmsetLeft:   omsetLeft,
		OmsetRight:  omsetRight,
		TotalOmset:  totalOmset,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

// getMembersAtSpecificLevel mengumpulkan anggota di level tertentu saja (bukan semua level sampai target)
func getMembersAtSpecificLevel(leftID, rightID *uint, db *gorm.DB, targetLevel int) []BinaryMember {
	return getMembersAtLevelRecursive(leftID, rightID, db, 1, targetLevel)
}

// getMembersAtLevelRecursive helper rekursif untuk mengumpulkan anggota di level tertentu
func getMembersAtLevelRecursive(leftID, rightID *uint, db *gorm.DB, currentLevel, targetLevel int) []BinaryMember {
	if currentLevel > targetLevel {
		return []BinaryMember{}
	}

	var members []BinaryMember

	// Jika sudah sampai target level, ambil anggota di level ini saja
	if currentLevel == targetLevel {
		// Process left side
		if leftID != nil {
			var leftUser models.User
			if err := db.Select("id, name, number").Where("id = ?", *leftID).First(&leftUser).Error; err == nil {
				// Calculate omset for this member (hanya dari investasi user tersebut, tanpa downline)
				memberOmset, _ := utils.CalculateUserOmsetOnly(leftUser.ID)
				members = append(members, BinaryMember{
					UserID:   leftUser.ID,
					Name:     leftUser.Name,
					Number:   leftUser.Number,
					Omset:    memberOmset,
					Position: "left",
				})
			}
		}

		// Process right side
		if rightID != nil {
			var rightUser models.User
			if err := db.Select("id, name, number").Where("id = ?", *rightID).First(&rightUser).Error; err == nil {
				// Calculate omset for this member (hanya dari investasi user tersebut, tanpa downline)
				memberOmset, _ := utils.CalculateUserOmsetOnly(rightUser.ID)
				members = append(members, BinaryMember{
					UserID:   rightUser.ID,
					Name:     rightUser.Name,
					Number:   rightUser.Number,
					Omset:    memberOmset,
					Position: "right",
				})
			}
		}
		return members
	}

	// Jika belum sampai target level, recurse ke level berikutnya
	// Process left side
	if leftID != nil {
		var leftNode models.BinaryNode
		if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
			leftMembers := getMembersAtLevelRecursive(leftNode.LeftID, leftNode.RightID, db, currentLevel+1, targetLevel)
			members = append(members, leftMembers...)
		}
	}

	// Process right side
	if rightID != nil {
		var rightNode models.BinaryNode
		if err := db.Where("user_id = ?", *rightID).First(&rightNode).Error; err == nil {
			rightMembers := getMembersAtLevelRecursive(rightNode.LeftID, rightNode.RightID, db, currentLevel+1, targetLevel)
			members = append(members, rightMembers...)
		}
	}

	return members
}

// GET /api/users/binary/omset
func GetBinaryOmsetHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	// Hitung omset dari level 1-3
	omsetLeft, omsetRight, totalOmset, err := utils.CalculateOmset(uid)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Hitung jumlah member per level
	db := database.DB
	var binaryNode models.BinaryNode
	if err := db.Where("user_id = ?", uid).First(&binaryNode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User belum punya binary node
			utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
				Success: true,
				Message: "Successfully",
				Data: OmsetResponse{
					OmsetLeft:   0,
					OmsetRight:  0,
					TotalOmset:  0,
					Level1Count: 0,
					Level2Count: 0,
					Level3Count: 0,
				},
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	level1Count := countMembersAtSpecificLevel(binaryNode.LeftID, binaryNode.RightID, db, 1)
	level2Count := countMembersAtSpecificLevel(binaryNode.LeftID, binaryNode.RightID, db, 2)
	level3Count := countMembersAtSpecificLevel(binaryNode.LeftID, binaryNode.RightID, db, 3)

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: OmsetResponse{
			OmsetLeft:   omsetLeft,
			OmsetRight:  omsetRight,
			TotalOmset:  totalOmset,
			Level1Count: level1Count,
			Level2Count: level2Count,
			Level3Count: level3Count,
		},
	})
}

// countMembersAtSpecificLevel menghitung jumlah member di level tertentu saja (bukan akumulasi)
// Level 1 = 2 anggota (left, right dari root)
// Level 2 = 4 anggota (anak dari level 1)
// Level 3 = 8 anggota (anak dari level 2)
func countMembersAtSpecificLevel(leftID, rightID *uint, db *gorm.DB, targetLevel int) int {
	if targetLevel == 1 {
		// Level 1: hitung left dan right dari root
		count := 0
		if leftID != nil {
			count++
		}
		if rightID != nil {
			count++
		}
		return count
	}

	// Untuk level 2 dan 3, kita perlu recurse ke level sebelumnya
	// Level 2: hitung semua anak dari level 1
	// Level 3: hitung semua anak dari level 2
	return countMembersAtLevelRecursive(leftID, rightID, db, 1, targetLevel)
}

// countMembersAtLevelRecursive helper rekursif untuk menghitung member di level tertentu saja
func countMembersAtLevelRecursive(leftID, rightID *uint, db *gorm.DB, currentLevel, targetLevel int) int {
	if currentLevel >= targetLevel {
		return 0
	}

	// Jika currentLevel + 1 == targetLevel, hitung anak-anak dari node ini
	if currentLevel+1 == targetLevel {
		count := 0
		// Hitung anak dari left
		if leftID != nil {
			var leftNode models.BinaryNode
			if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
				if leftNode.LeftID != nil {
					count++
				}
				if leftNode.RightID != nil {
					count++
				}
			}
		}
		// Hitung anak dari right
		if rightID != nil {
			var rightNode models.BinaryNode
			if err := db.Where("user_id = ?", *rightID).First(&rightNode).Error; err == nil {
				if rightNode.LeftID != nil {
					count++
				}
				if rightNode.RightID != nil {
					count++
				}
			}
		}
		return count
	}

	// Jika belum sampai target level, recurse ke level berikutnya
	count := 0
	if leftID != nil {
		var leftNode models.BinaryNode
		if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
			count += countMembersAtLevelRecursive(leftNode.LeftID, leftNode.RightID, db, currentLevel+1, targetLevel)
		}
	}
	if rightID != nil {
		var rightNode models.BinaryNode
		if err := db.Where("user_id = ?", *rightID).First(&rightNode).Error; err == nil {
			count += countMembersAtLevelRecursive(rightNode.LeftID, rightNode.RightID, db, currentLevel+1, targetLevel)
		}
	}
	return count
}

// GET /api/users/rewards
func GetRewardsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB

	// Get all active rewards
	var rewards []models.Reward
	if err := db.Where("status = ?", "Active").Order("omset_target ASC").Find(&rewards).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Update reward progress terlebih dahulu untuk memastikan data terbaru
	_ = utils.UpdateRewardProgress(uid)

	// Get reward progress for this user
	var progressList []models.RewardProgress
	if err := db.Where("user_id = ?", uid).Preload("Reward").Find(&progressList).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Create progress map
	progressMap := make(map[uint]*models.RewardProgress)
	for i := range progressList {
		progressMap[progressList[i].RewardID] = &progressList[i]
	}

	// Build response
	type RewardWithProgress struct {
		ID            uint    `json:"id"`
		Name          string  `json:"name"`
		OmsetTarget   float64 `json:"omset_target"`
		RewardDesc    string  `json:"reward_desc"`
		Duration      int     `json:"duration"`
		IsAccumulative bool   `json:"is_accumulative"`
		OmsetLeft     float64 `json:"omset_left"`
		OmsetRight    float64 `json:"omset_right"`
		TotalOmset    float64 `json:"total_omset"`
		IsCompleted   bool   `json:"is_completed"`
		IsClaimed     bool   `json:"is_claimed"`
		StartedAt     string `json:"started_at,omitempty"`
		ExpiresAt     string `json:"expires_at,omitempty"`
		Progress      float64 `json:"progress"` // Percentage (0-100)
	}

	items := make([]RewardWithProgress, 0, len(rewards))
	for _, reward := range rewards {
		item := RewardWithProgress{
			ID:             reward.ID,
			Name:           reward.Name,
			OmsetTarget:    reward.OmsetTarget,
			RewardDesc:     reward.RewardDesc,
			Duration:       reward.Duration,
			IsAccumulative: reward.IsAccumulative,
			OmsetLeft:      0,
			OmsetRight:     0,
			TotalOmset:     0,
			IsCompleted:    false,
			IsClaimed:      false,
			Progress:       0,
		}

		// Get progress if exists
		if progress, ok := progressMap[reward.ID]; ok {
			item.OmsetLeft = progress.OmsetLeft
			item.OmsetRight = progress.OmsetRight
			item.TotalOmset = progress.TotalOmset
			item.IsCompleted = progress.IsCompleted
			item.IsClaimed = progress.IsClaimed
			item.StartedAt = progress.StartedAt.Format("2006-01-02T15:04:05Z07:00")
			if progress.ExpiresAt != nil {
				item.ExpiresAt = progress.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
			}
		} else {
			// Jika progress belum ada, initialize dulu
			_ = utils.InitializeRewardProgress(uid)
			_ = utils.UpdateRewardProgress(uid)
			// Reload progress
			var newProgress models.RewardProgress
			if err := db.Where("user_id = ? AND reward_id = ?", uid, reward.ID).First(&newProgress).Error; err == nil {
				item.OmsetLeft = newProgress.OmsetLeft
				item.OmsetRight = newProgress.OmsetRight
				item.TotalOmset = newProgress.TotalOmset
				item.IsCompleted = newProgress.IsCompleted
				item.IsClaimed = newProgress.IsClaimed
				item.StartedAt = newProgress.StartedAt.Format("2006-01-02T15:04:05Z07:00")
				if newProgress.ExpiresAt != nil {
					item.ExpiresAt = newProgress.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
				}
			}
		}

		// Calculate progress percentage (0-100)
		if reward.OmsetTarget > 0 {
			item.Progress = (item.TotalOmset / reward.OmsetTarget) * 100
			if item.Progress > 100 {
				item.Progress = 100
			}
			if item.Progress < 0 {
				item.Progress = 0
			}
		} else {
			item.Progress = 0
		}

		items = append(items, item)
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"rewards": items},
	})
}

