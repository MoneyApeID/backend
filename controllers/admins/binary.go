package admins

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// BinaryStructureAdminResponse untuk response struktur binary admin
type BinaryStructureAdminResponse struct {
	UserID      uint   `json:"user_id"`
	UserName    string `json:"user_name"`
	UserNumber  string `json:"user_number"`
	LeftID      *uint  `json:"left_id"`
	RightID     *uint  `json:"right_id"`
	LeftName    string `json:"left_name,omitempty"`
	RightName   string `json:"right_name,omitempty"`
	OmsetLeft   float64 `json:"omset_left"`
	OmsetRight  float64 `json:"omset_right"`
	TotalOmset  float64 `json:"total_omset"`
	Level1Count int     `json:"level1_count"`
	Level2Count int     `json:"level2_count"`
	Level3Count int     `json:"level3_count"`
}

// GET /api/admin/binary
// List semua user yang memiliki binary node
func GetBinaryStructureAdminHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Get all users that have binary nodes
	var binaryNodes []models.BinaryNode
	if err := db.Find(&binaryNodes).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Get unique user IDs
	userIDMap := make(map[uint]bool)
	for _, node := range binaryNodes {
		userIDMap[node.UserID] = true
	}

	// Convert to slice
	userIDs := make([]uint, 0, len(userIDMap))
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	// Get users
	var users []models.User
	if len(userIDs) > 0 {
		if err := db.Select("id, name, number").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
			return
		}
	}

	// Build response
	type UserListResponse struct {
		UserID     uint   `json:"user_id"`
		UserName   string `json:"user_name"`
		UserNumber string `json:"user_number"`
	}

	response := make([]UserListResponse, 0, len(users))
	for _, user := range users {
		response = append(response, UserListResponse{
			UserID:     user.ID,
			UserName:   user.Name,
			UserNumber: user.Number,
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"users": response},
	})
}

// BinaryMemberAdmin untuk response binary structure admin
type BinaryMemberAdmin struct {
	UserID   uint    `json:"user_id"`
	Name     string  `json:"name"`
	Number   string  `json:"number"`
	Omset    float64 `json:"omset"`
	Position string  `json:"position"`
}

// BinaryStructureResponse untuk response binary structure admin
type BinaryStructureResponse struct {
	Root       BinaryMemberAdmin   `json:"root"`
	Level1     []BinaryMemberAdmin `json:"level1"`
	Level2     []BinaryMemberAdmin `json:"level2"`
	Level3     []BinaryMemberAdmin `json:"level3"`
	OmsetLeft  float64 `json:"omset_left"`
	OmsetRight float64 `json:"omset_right"`
	TotalOmset float64 `json:"total_omset"`
}

// GET /api/admin/binary/details/{id}
// Get detail binary structure untuk user tertentu (sama seperti user endpoint)
func GetBinaryDetailsAdminHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "User ID tidak valid"})
		return
	}

	db := database.DB

	// Get user info (root)
	var rootUser models.User
	if err := db.Select("id, name, number").Where("id = ?", userID).First(&rootUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "User tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Calculate root omset (hanya dari investasi root sendiri, tanpa downline untuk display)
	rootOmset, _ := utils.CalculateUserOmsetOnly(uint(userID))

	// Get binary node
	var binaryNode models.BinaryNode
	if err := db.Where("user_id = ?", userID).First(&binaryNode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User belum punya binary node, return empty structure
			response := BinaryStructureResponse{
				Root: BinaryMemberAdmin{
					UserID: rootUser.ID,
					Name:   rootUser.Name,
					Number: rootUser.Number,
					Omset:  rootOmset,
				},
				Level1:      []BinaryMemberAdmin{},
				Level2:      []BinaryMemberAdmin{},
				Level3:      []BinaryMemberAdmin{},
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
	omsetLeft, omsetRight, totalOmset, _ := utils.CalculateOmset(uint(userID))

	// Build structure per level dengan prioritas omset terbesar
	level1Members := utils.GetTopMembersByOmset(uint(userID), 1, 2)
	level2Members := utils.GetTopMembersByOmset(uint(userID), 2, 4)
	level3Members := utils.GetTopMembersByOmset(uint(userID), 3, 8)

	level1 := make([]BinaryMemberAdmin, len(level1Members))
	for i, m := range level1Members {
		level1[i] = BinaryMemberAdmin{
			UserID:   m.UserID,
			Name:     m.Name,
			Number:   m.Number,
			Omset:    m.Omset,
			Position: m.Position,
		}
	}
	level2 := make([]BinaryMemberAdmin, len(level2Members))
	for i, m := range level2Members {
		level2[i] = BinaryMemberAdmin{
			UserID:   m.UserID,
			Name:     m.Name,
			Number:   m.Number,
			Omset:    m.Omset,
			Position: m.Position,
		}
	}
	level3 := make([]BinaryMemberAdmin, len(level3Members))
	for i, m := range level3Members {
		level3[i] = BinaryMemberAdmin{
			UserID:   m.UserID,
			Name:     m.Name,
			Number:   m.Number,
			Omset:    m.Omset,
			Position: m.Position,
		}
	}

	response := BinaryStructureResponse{
		Root: BinaryMemberAdmin{
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

// POST /api/admin/binary/claim
// Claim reward - reset omset menjadi 0 dan reset progress untuk user dan reward tertentu
func ClaimRewardAdminHandler(w http.ResponseWriter, r *http.Request) {
	type ClaimRequest struct {
		UserID   uint `json:"user_id"`
		RewardID uint `json:"reward_id"`
	}

	var req ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Request tidak valid"})
		return
	}

	if req.UserID == 0 || req.RewardID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "User ID dan Reward ID harus diisi"})
		return
	}

	db := database.DB

	// Check if reward progress exists
	var progress models.RewardProgress
	if err := db.Where("user_id = ? AND reward_id = ?", req.UserID, req.RewardID).First(&progress).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Reward progress tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Reset omset dan progress
	now := time.Now()
	progress.OmsetLeft = 0
	progress.OmsetRight = 0
	progress.TotalOmset = 0
	progress.IsCompleted = false
	progress.IsClaimed = true
	progress.StartedAt = now
	progress.LastResetAt = &now

	// Get reward untuk set expires_at
	var reward models.Reward
	if err := db.Where("id = ?", req.RewardID).First(&reward).Error; err == nil {
		if !reward.IsAccumulative {
			exp := now.Add(time.Duration(reward.Duration) * 24 * time.Hour)
			progress.ExpiresAt = &exp
		} else {
			progress.ExpiresAt = nil
		}
	}

	// Save progress dengan omset yang sudah di-reset
	if err := db.Save(&progress).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat menyimpan"})
		return
	}

	// Setelah claim, pastikan omset tetap 0 dengan menggunakan UpdateColumn untuk memastikan nilai 0 tersimpan
	// Ini penting karena UpdateRewardProgress mungkin dipanggil setelah claim
	// Kita juga perlu memastikan bahwa omset tidak di-update sampai ada aktivitas baru
	// dengan cara menggunakan UpdateColumns yang akan memastikan nilai 0 tersimpan
	if err := db.Model(&progress).Where("id = ?", progress.ID).UpdateColumns(map[string]interface{}{
		"omset_left":   0,
		"omset_right":  0,
		"total_omset":  0,
		"is_completed": false,
		"is_claimed":   true,
		"started_at":   now,
		"last_reset_at": now,
	}).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat menyimpan omset"})
		return
	}

	// Reset binary structure: reset left_id dan right_id pada level 1 dari root
	// Ini akan memutuskan koneksi level 1, sehingga level 2 juga tidak terhubung
	var rootBinaryNode models.BinaryNode
	if err := db.Where("user_id = ?", req.UserID).First(&rootBinaryNode).Error; err == nil {
		// Get left_id dan right_id sebelum di-reset (untuk logging atau info)
		leftIDBefore := rootBinaryNode.LeftID
		rightIDBefore := rootBinaryNode.RightID

		// Reset left_id dan right_id menjadi NULL
		if err := db.Model(&rootBinaryNode).UpdateColumns(map[string]interface{}{
			"left_id":  nil,
			"right_id": nil,
		}).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat reset binary structure"})
			return
		}

		// Level 2 otomatis tidak terhubung karena mereka adalah child dari level 1
		// Tidak perlu menghapus binary node level 2, cukup reset koneksi level 1
		// User level 1 dan 2 masih memiliki binary node mereka sendiri, tapi tidak terhubung ke root

		// Optional: Log atau return info tentang user yang di-reset
		_ = leftIDBefore
		_ = rightIDBefore
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Reward berhasil di-claim, omset dan progress telah di-reset, binary structure level 1 telah di-reset",
		Data:    map[string]interface{}{"user_id": req.UserID, "reward_id": req.RewardID},
	})
}

// countMembersAtLevel menghitung jumlah member sampai level tertentu
func countMembersAtLevel(leftID, rightID *uint, db *gorm.DB, currentLevel, targetLevel int) int {
	if currentLevel > targetLevel {
		return 0
	}

	count := 0
	if leftID != nil {
		count++
		if currentLevel < targetLevel {
			var leftNode models.BinaryNode
			if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
				count += countMembersAtLevel(leftNode.LeftID, leftNode.RightID, db, currentLevel+1, targetLevel)
			}
		}
	}
	if rightID != nil {
		count++
		if currentLevel < targetLevel {
			var rightNode models.BinaryNode
			if err := db.Where("user_id = ?", *rightID).First(&rightNode).Error; err == nil {
				count += countMembersAtLevel(rightNode.LeftID, rightNode.RightID, db, currentLevel+1, targetLevel)
			}
		}
	}
	return count
}

// GET /api/admin/binary/rewards
// Melihat semua reward progress dari semua user
func GetBinaryRewardsAdminHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Get all reward progress with user and reward info
	var progressList []models.RewardProgress
	if err := db.Preload("User").Preload("Reward").Order("user_id ASC, reward_id ASC").Find(&progressList).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Build response
	type RewardProgressResponse struct {
		ID            uint    `json:"id"`
		UserID        uint    `json:"user_id"`
		UserName      string  `json:"user_name"`
		UserNumber    string  `json:"user_number"`
		RewardID      uint    `json:"reward_id"`
		RewardName    string  `json:"reward_name"`
		OmsetTarget   float64 `json:"omset_target"`
		RewardDesc    string  `json:"reward_desc"`
		OmsetLeft     float64 `json:"omset_left"`
		OmsetRight    float64 `json:"omset_right"`
		TotalOmset    float64 `json:"total_omset"`
		IsCompleted   bool   `json:"is_completed"`
		IsClaimed     bool   `json:"is_claimed"`
		StartedAt     string `json:"started_at"`
		ExpiresAt     string `json:"expires_at,omitempty"`
		Progress      float64 `json:"progress"`
	}

	items := make([]RewardProgressResponse, 0, len(progressList))
	for _, progress := range progressList {
		item := RewardProgressResponse{
			ID:          progress.ID,
			UserID:      progress.UserID,
			RewardID:    progress.RewardID,
			OmsetLeft:   progress.OmsetLeft,
			OmsetRight:  progress.OmsetRight,
			TotalOmset:  progress.TotalOmset,
			IsCompleted: progress.IsCompleted,
			IsClaimed:   progress.IsClaimed,
			StartedAt:   progress.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if progress.User != nil {
			item.UserName = progress.User.Name
			item.UserNumber = progress.User.Number
		}

		if progress.Reward != nil {
			item.RewardName = progress.Reward.Name
			item.OmsetTarget = progress.Reward.OmsetTarget
			item.RewardDesc = progress.Reward.RewardDesc
			// Calculate progress percentage
			if progress.Reward.OmsetTarget > 0 {
				item.Progress = (progress.TotalOmset / progress.Reward.OmsetTarget) * 100
				if item.Progress > 100 {
					item.Progress = 100
				}
			}
		}

		if progress.ExpiresAt != nil {
			item.ExpiresAt = progress.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}

		items = append(items, item)
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"reward_progress": items},
	})
}

