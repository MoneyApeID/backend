package admins

import (
	"encoding/json"
	"net/http"
	"strconv"
	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// GET /api/admin/rewards
// List semua rewards
func ListRewardsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	var rewards []models.Reward
	if err := db.Order("omset_target ASC").Find(&rewards).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"rewards": rewards},
	})
}

// POST /api/admin/rewards
// Create new reward
func CreateRewardHandler(w http.ResponseWriter, r *http.Request) {
	type CreateRewardRequest struct {
		Name           string  `json:"name"`
		OmsetTarget    float64 `json:"omset_target"`
		RewardDesc     *string `json:"reward_desc,omitempty"`
		Duration       int     `json:"duration"`
		IsAccumulative bool    `json:"is_accumulative"`
		Status         string  `json:"status"` // "Active" or "Inactive"
	}

	var req CreateRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Request tidak valid"})
		return
	}

	if req.Name == "" || req.OmsetTarget <= 0 || req.Duration <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Name, omset_target, dan duration harus diisi"})
		return
	}

	if req.Status != "Active" && req.Status != "Inactive" {
		req.Status = "Active"
	}

	db := database.DB

	rewardDesc := ""
	if req.RewardDesc != nil {
		rewardDesc = *req.RewardDesc
	}

	reward := models.Reward{
		Name:           req.Name,
		OmsetTarget:    req.OmsetTarget,
		RewardDesc:     rewardDesc,
		Duration:       req.Duration,
		IsAccumulative: req.IsAccumulative,
		Status:         req.Status,
	}

	if err := db.Create(&reward).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat membuat reward"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Reward berhasil dibuat",
		Data:    reward,
	})
}

// PUT /api/admin/rewards
// Update reward
func UpdateRewardHandler(w http.ResponseWriter, r *http.Request) {
	type UpdateRewardRequest struct {
		ID             uint    `json:"id"`
		Name           *string `json:"name,omitempty"`
		OmsetTarget    *float64 `json:"omset_target,omitempty"`
		RewardDesc     *string `json:"reward_desc,omitempty"`
		Duration       *int    `json:"duration,omitempty"`
		IsAccumulative *bool   `json:"is_accumulative,omitempty"`
		Status         *string `json:"status,omitempty"` // "Active" or "Inactive"
	}

	var req UpdateRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Request tidak valid"})
		return
	}

	if req.ID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID harus diisi"})
		return
	}

	db := database.DB

	var reward models.Reward
	if err := db.Where("id = ?", req.ID).First(&reward).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Reward tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Update fields
	if req.Name != nil {
		reward.Name = *req.Name
	}
	if req.OmsetTarget != nil {
		reward.OmsetTarget = *req.OmsetTarget
	}
	if req.RewardDesc != nil {
		reward.RewardDesc = *req.RewardDesc
	}
	if req.Duration != nil {
		reward.Duration = *req.Duration
	}
	if req.IsAccumulative != nil {
		reward.IsAccumulative = *req.IsAccumulative
	}
	if req.Status != nil {
		if *req.Status == "Active" || *req.Status == "Inactive" {
			reward.Status = *req.Status
		}
	}

	if err := db.Save(&reward).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat mengupdate reward"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Reward berhasil di-update",
		Data:    reward,
	})
}

// DELETE /api/admin/rewards/{id}
// Delete reward
func DeleteRewardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID tidak valid"})
		return
	}

	db := database.DB

	// Check if reward exists
	var reward models.Reward
	if err := db.Where("id = ?", id).First(&reward).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Reward tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Check if there are any reward progress using this reward
	var count int64
	if err := db.Model(&models.RewardProgress{}).Where("reward_id = ?", id).Count(&count).Error; err == nil {
		if count > 0 {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Reward tidak dapat dihapus karena masih digunakan oleh reward progress"})
			return
		}
	}

	// Delete reward
	if err := db.Delete(&reward).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat menghapus reward"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Reward berhasil dihapus",
	})
}

