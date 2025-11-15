package users

import (
	"encoding/json"
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strings"

	"gorm.io/gorm"
)

// GET /api/users/task
func TaskListHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	db := database.DB
	var tasks []models.Task
	if err := db.Where("status = ?", "Active").Order("id ASC").Find(&tasks).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "DB error"})
		return
	}
	// Get claimed tasks
	var claimed []models.UserTask
	db.Where("user_id = ?", uid).Find(&claimed)
	claimedMap := map[uint]bool{}
	for _, ut := range claimed {
		claimedMap[ut.TaskID] = true
	}
	// Calculate active subordinates for each level
	getActiveCount := func(level int) int64 {
		db := database.DB
		var level1 []models.User
		if err := db.Where("reff_by = ?", uid).Find(&level1).Error; err != nil {
			return 0
		}
		if level == 1 {
			n := 0
			for _, u := range level1 {
				if strings.ToLower(u.InvestmentStatus) == "active" {
					n++
				}
			}
			return int64(n)
		}
		// Level 2
		level1IDs := make([]uint, 0, len(level1))
		for _, u := range level1 {
			level1IDs = append(level1IDs, u.ID)
		}
		var level2 []models.User
		if len(level1IDs) > 0 {
			db.Where("reff_by IN ?", level1IDs).Find(&level2)
		}
		if level == 2 {
			n := 0
			for _, u := range level2 {
				if strings.ToLower(u.InvestmentStatus) == "active" {
					n++
				}
			}
			return int64(n)
		}
		// Level 3
		level2IDs := make([]uint, 0, len(level2))
		for _, u := range level2 {
			level2IDs = append(level2IDs, u.ID)
		}
		var level3 []models.User
		if len(level2IDs) > 0 {
			db.Where("reff_by IN ?", level2IDs).Find(&level3)
		}
		n := 0
		for _, u := range level3 {
			if strings.ToLower(u.InvestmentStatus) == "active" {
				n++
			}
		}
		return int64(n)
	}
	var resp []map[string]interface{}
	for _, t := range tasks {
		activeCount := getActiveCount(t.RequiredLevel)
		percent := 0
		if t.RequiredActiveMembers > 0 {
			percent = int((float64(activeCount) / float64(t.RequiredActiveMembers)) * 100)
			if percent > 100 {
				percent = 100
			}
		}
		taken := claimedMap[t.ID]
		lock := activeCount < t.RequiredActiveMembers
		resp = append(resp, map[string]interface{}{
			"id":                       t.ID,
			"name":                     t.Name,
			"reward":                   t.Reward,
			"required_level":           t.RequiredLevel,
			"required_active_members":  t.RequiredActiveMembers,
			"active_subordinate_count": activeCount,
			"taken":                    taken,
			"lock":                     lock,
			"percent":                  percent,
		})
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: resp})
}

// POST /api/users/task/submit
func TaskSubmitHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	var req struct {
		TaskID uint `json:"task_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TaskID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid request"})
		return
	}
	db := database.DB
	var task models.Task
	if err := db.First(&task, req.TaskID).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Task not found"})
		return
	}
	// Check if already claimed
	var userTask models.UserTask
	if err := db.Where("user_id = ? AND task_id = ?", uid, task.ID).First(&userTask).Error; err == nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Tugas sudah pernah diambil"})
		return
	}
	// Check active subordinates
	getActiveCount := func(level int) int64 {
		db := database.DB
		var level1 []models.User
		if err := db.Where("reff_by = ?", uid).Find(&level1).Error; err != nil {
			return 0
		}
		if level == 1 {
			n := 0
			for _, u := range level1 {
				if strings.ToLower(u.InvestmentStatus) == "active" {
					n++
				}
			}
			return int64(n)
		}
		// Level 2
		level1IDs := make([]uint, 0, len(level1))
		for _, u := range level1 {
			level1IDs = append(level1IDs, u.ID)
		}
		var level2 []models.User
		if len(level1IDs) > 0 {
			db.Where("reff_by IN ?", level1IDs).Find(&level2)
		}
		if level == 2 {
			n := 0
			for _, u := range level2 {
				if strings.ToLower(u.InvestmentStatus) == "active" {
					n++
				}
			}
			return int64(n)
		}
		// Level 3
		level2IDs := make([]uint, 0, len(level2))
		for _, u := range level2 {
			level2IDs = append(level2IDs, u.ID)
		}
		var level3 []models.User
		if len(level2IDs) > 0 {
			db.Where("reff_by IN ?", level2IDs).Find(&level3)
		}
		n := 0
		for _, u := range level3 {
			if strings.ToLower(u.InvestmentStatus) == "active" {
				n++
			}
		}
		return int64(n)
	}
	activeCount := getActiveCount(task.RequiredLevel)
	if activeCount < task.RequiredActiveMembers {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Belum memenuhi syarat tugas"})
		return
	}
	// Add reward to user balance
	if err := db.Model(&models.User{}).Where("id = ?", uid).Update("income", gorm.Expr("income + ?", task.Reward)).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed to update balance"})
		return
	}
	// Mark as claimed (let claimed_at use DB default)
	db.Model(&models.UserTask{}).Create(map[string]interface{}{
		"user_id": uid,
		"task_id": task.ID,
	})

	db.Model(&models.Transaction{}).Create(map[string]interface{}{
		"user_id":          uid,
		"amount":           task.Reward,
		"charge":           0,
		"order_id":         utils.GenerateOrderID(uid),
		"transaction_flow": "debit",
		"transaction_type": "bonus",
		"message":          ptrString("Reward tugas: " + task.Name),
		"status":           "Success",
	})
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Hadiah berhasil diselesaikan"})
}

func ptrString(s string) *string {
	return &s
}
