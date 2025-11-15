package admins

import (
	"errors"
	"net/http"
	"strconv"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DepositResponse untuk response list deposits
type DepositResponse struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	UserName      string  `json:"user_name"`
	Phone         string  `json:"phone"`
	Amount        float64 `json:"amount"`
	OrderID       string  `json:"order_id"`
	PaymentMethod string  `json:"payment_method"`
	PaymentChannel *string `json:"payment_channel,omitempty"`
	PaymentCode   *string `json:"payment_code,omitempty"`
	Status        string  `json:"status"`
	ExpiredAt     string  `json:"expired_at"`
	CreatedAt     string  `json:"created_at"`
}

// GET /api/admin/deposits
func GetDeposits(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")
	orderID := r.URL.Query().Get("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Start query
	db := database.DB
	query := db.Model(&models.Deposit{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if orderID != "" {
		query = query.Where("order_id LIKE ?", "%"+orderID+"%")
	}

	// Get deposits
	var deposits []models.Deposit
	query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&deposits)

	// Get total count for pagination
	var total int64
	countQuery := db.Model(&models.Deposit{})
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if orderID != "" {
		countQuery = countQuery.Where("order_id LIKE ?", "%"+orderID+"%")
	}
	countQuery.Count(&total)

	// Prepare user IDs to fetch names in batch
	userIDsSet := make(map[uint]struct{})
	for _, dep := range deposits {
		userIDsSet[dep.UserID] = struct{}{}
	}
	var userIDs []uint
	for id := range userIDsSet {
		userIDs = append(userIDs, id)
	}

	// Fetch users and build a map[id]user
	usersByID := make(map[uint]models.User, len(userIDs))
	if len(userIDs) > 0 {
		var users []models.User
		db.Select("id, name, number").Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			usersByID[u.ID] = u
		}
	}

	// Transform to response format
	var response []DepositResponse
	for _, dep := range deposits {
		user := usersByID[dep.UserID]
		response = append(response, DepositResponse{
			ID:            dep.ID,
			UserID:        dep.UserID,
			UserName:      user.Name,
			Phone:         user.Number,
			Amount:        dep.Amount,
			OrderID:       dep.OrderID,
			PaymentMethod: dep.PaymentMethod,
			PaymentChannel: dep.PaymentChannel,
			PaymentCode:   dep.PaymentCode,
			Status:        dep.Status,
			ExpiredAt:     dep.ExpiredAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:     dep.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: map[string]interface{}{
			"deposits": response,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		},
	})
}

// PUT /api/admin/deposits/{id}/approve
func ApproveDeposit(w http.ResponseWriter, r *http.Request) {
	// Get deposit ID from path variable
	vars := mux.Vars(r)
	depositIDStr := vars["id"]
	if depositIDStr == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID deposit tidak valid"})
		return
	}

	depositID, err := strconv.ParseUint(depositIDStr, 10, 64)
	if err != nil || depositID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID deposit tidak valid"})
		return
	}

	db := database.DB

	// Get deposit
	var deposit models.Deposit
	if err := db.Where("id = ?", uint(depositID)).First(&deposit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Deposit tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Check if already processed
	if deposit.Status == "Success" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Deposit sudah berstatus Success"})
		return
	}

	// Update deposit status, add balance, update transaction, and give spin tickets
	err = db.Transaction(func(tx *gorm.DB) error {
		// Update deposit status
		if err := tx.Model(&models.Deposit{}).Where("id = ?", deposit.ID).Update("status", "Success").Error; err != nil {
			return err
		}

		// Get user with lock
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, balance, spin_ticket").Where("id = ?", deposit.UserID).First(&user).Error; err != nil {
			return err
		}

		// Add balance
		newBalance := utils.RoundFloat(user.Balance+deposit.Amount, 2)
		if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("balance", newBalance).Error; err != nil {
			return err
		}

		// Update transaction status (find by order_id)
		var transaction models.Transaction
		if err := tx.Where("order_id = ? AND transaction_type = ?", deposit.OrderID, "deposit").First(&transaction).Error; err == nil {
			// Update transaction status
			if err := tx.Model(&models.Transaction{}).Where("id = ?", transaction.ID).Update("status", "Success").Error; err != nil {
				return err
			}
		}

		// Bonus spin ticket berdasarkan jumlah deposit
		// 100k-499k → 1 ticket, 500k+ → 2 tickets
		if deposit.Amount >= 100000 {
			var spinTicketsToAdd uint = 1
			if deposit.Amount >= 500000 {
				spinTicketsToAdd = 2
			}

			// Update spin ticket
			if user.SpinTicket == nil {
				// Jika spin_ticket masih nil, set langsung
				if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("spin_ticket", spinTicketsToAdd).Error; err != nil {
					return err
				}
			} else {
				// Jika sudah ada, tambahkan
				if err := tx.Model(&models.User{}).Where("id = ?", user.ID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket + ?", spinTicketsToAdd)).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memproses deposit"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Deposit berhasil disetujui",
		Data: map[string]interface{}{
			"deposit_id": deposit.ID,
			"status":     "Success",
		},
	})
}

// PUT /api/admin/deposits/{id}/reject
func RejectDeposit(w http.ResponseWriter, r *http.Request) {
	// Get deposit ID from path variable
	vars := mux.Vars(r)
	depositIDStr := vars["id"]
	if depositIDStr == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID deposit tidak valid"})
		return
	}

	depositID, err := strconv.ParseUint(depositIDStr, 10, 64)
	if err != nil || depositID == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID deposit tidak valid"})
		return
	}

	db := database.DB

	// Get deposit
	var deposit models.Deposit
	if err := db.Where("id = ?", uint(depositID)).First(&deposit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Deposit tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Check if already processed
	if deposit.Status == "Failed" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Deposit sudah berstatus Failed"})
		return
	}

	// Update deposit status to Failed
	if err := db.Model(&models.Deposit{}).Where("id = ?", deposit.ID).Update("status", "Failed").Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengupdate status"})
		return
	}

	// Update transaction status to Failed if exists
	var transaction models.Transaction
	if err := db.Where("order_id = ? AND transaction_type = ?", deposit.OrderID, "deposit").First(&transaction).Error; err == nil {
		db.Model(&models.Transaction{}).Where("id = ?", transaction.ID).Update("status", "Failed")
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Deposit berhasil ditolak",
		Data: map[string]interface{}{
			"deposit_id": deposit.ID,
			"status":     "Failed",
		},
	})
}

