package admins

import (
	"encoding/json"
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type ForumResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	UserName    string    `json:"username"`
	Phone       string    `json:"phone"`
	Reward      float64   `json:"reward"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// GET /api/admin/forums
func GetForumsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Parse query parameters
	status := r.URL.Query().Get("status")
	IDStr := r.URL.Query().Get("id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	// Build query
	query := db.Table("forums").
		Select("forums.*, users.name as user_name, users.number as phone").
		Joins("LEFT JOIN users ON forums.user_id = users.id")

	// Apply filters
	if status != "" {
		query = query.Where("forums.status = ?", status)
	}
	if IDStr != "" {
		ID, err := strconv.Atoi(IDStr)
		if err == nil {
			query = query.Where("forums.id = ?", ID)
		}
	}
	if startDate != "" {
		query = query.Where("forums.created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("forums.created_at <= ?", endDate)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("users.name LIKE ? OR users.number LIKE ?", like, like)
	}

	// Parse pagination parameters
	page := 1
	limit := 20
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := (page - 1) * limit

	// Execute query
	type ForumWithUserName struct {
		models.Forum
		UserName string `gorm:"column:user_name"`
		Phone    string `gorm:"column:phone"`
	}
	var forums []ForumWithUserName
	if err := query.Order("forums.created_at DESC").Offset(offset).Limit(limit).Find(&forums).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Transform to response format
	var response []ForumResponse
	for _, f := range forums {
		response = append(response, ForumResponse{
			ID:          f.ID,
			UserID:      f.UserID,
			UserName:    f.UserName,
			Phone:       f.Phone,
			Reward:      f.Reward,
			Description: f.Description,
			Image:       f.Image,
			Status:      f.Status,
			CreatedAt:   f.CreatedAt,
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

type ApproveForumRequest struct {
	Reward float64 `json:"reward"`
}

// PUT /api/admin/forums/:id/approve
func ApproveForumHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumID := vars["id"]

	var req ApproveForumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	db := database.DB

	err := db.Transaction(func(tx *gorm.DB) error {
		// Get forum
		var forum models.Forum
		if err := tx.First(&forum, forumID).Error; err != nil {
			return err
		}

		if forum.Status != "Pending" {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Forum sudah diproses sebelumnya",
			})
			return nil
		}

		// Update forum status and reward
		forum.Status = "Accepted"
		forum.Reward = req.Reward
		if err := tx.Save(&forum).Error; err != nil {
			return err
		}

		// Add reward to user balance
		if err := tx.Model(&models.User{}).
			Where("id = ?", forum.UserID).
			UpdateColumn("income", gorm.Expr("income + ?", req.Reward)).Error; err != nil {
			return err
		}

		// Create bonus transaction
		msg := "Hadiah Forum Post"
		trx := models.Transaction{
			UserID:          forum.UserID,
			Amount:          req.Reward,
			Charge:          0,
			OrderID:         utils.GenerateOrderID(forum.UserID),
			TransactionFlow: "debit",
			TransactionType: "bonus",
			Message:         &msg,
			Status:          "Success",
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Forum telah berhasil disetujui dan hadiah telah diberikan",
		Data: map[string]interface{}{
			"id":     forumID,
			"status": "Accepted",
		},
	})
}

// PUT /api/admin/forums/:id/reject
func RejectForumHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumID := vars["id"]

	db := database.DB

	var forum models.Forum
	if err := db.First(&forum, forumID).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
			Success: false,
			Message: "Forum tidak ditemukan",
		})
		return
	}

	if forum.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Forum sudah di proses sebelumnya",
		})
		return
	}

	// Update forum status
	forum.Status = "Rejected"
	if err := db.Save(&forum).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Forum telah berhasil ditolak",
		Data: map[string]interface{}{
			"id":     forumID,
			"status": "Rejected",
		},
	})
}
