package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"project/database"
	"project/models"
	"project/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WithdrawalRequest struct {
	Amount        float64 `json:"amount"`
	BankAccountID uint    `json:"bank_account_id"`
}

func WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	var req WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid JSON"})
		return
	}

	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	// Load settings
	db := database.DB
	var setting models.Setting
	if err := db.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}

	// Validate amount
	if req.Amount < setting.MinWithdraw {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: fmt.Sprintf("Minimal penarikan adalah Rp%.0f", setting.MinWithdraw)})
		return
	}
	if req.Amount > setting.MaxWithdraw {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: fmt.Sprintf("Maksimal penarikan adalah Rp%.0f", setting.MaxWithdraw)})
		return
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	hour := now.Hour()
	if hour < 12 || hour >= 17 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Penarikan hanya dapat dilakukan pada pukul 12:00 - 17:00 WIB"})
		return
	}

	if now.Weekday() == time.Sunday {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Penarikan hanya dapat dilakukan pada hari Senin sampai Sabtu"})
		return
	}

	// Check if user has already made a withdrawal today
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)
	var todayWithdrawals int64
	if err := db.Model(&models.Withdrawal{}).Where("user_id = ? AND created_at BETWEEN ? AND ?", uid, startOfDay, endOfDay).Count(&todayWithdrawals).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}
	if todayWithdrawals > 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Anda hanya dapat melakukan 1 kali penarikan dalam sehari"})
		return
	}

	// Load bank account owned by user
	var acc models.BankAccount
	if err := db.Preload("Bank").Where("id = ? AND user_id = ?", req.BankAccountID, uid).First(&acc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Rekening tujuan tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}
	if acc.Bank == nil || acc.Bank.Status != "Active" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Layanan bank ini sedang dalam pemeliharaan"})
		return
	}

	// Compute charge and final amount
	charge := round2(req.Amount * (setting.WithdrawCharge / 100.0))
	finalAmount := req.Amount - charge
	orderID := utils.GenerateOrderID(uid)

	// Sentinel error for insufficient balance
	var errInsufficientBalance = errors.New("insufficient_balance")

	var wd models.Withdrawal
	if err := db.Transaction(func(tx *gorm.DB) error {
		// Lock user row for update and validate balance
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, uid).Error; err != nil {
			return err
		}
		if user.Income < req.Amount {
			return errInsufficientBalance
		}
		newBalance := round2(user.Income - req.Amount)
		if err := tx.Model(&user).Update("income", newBalance).Error; err != nil {
			return err
		}

		// Create withdrawal pending
		wd = models.Withdrawal{
			UserID:        uid,
			BankAccountID: acc.ID,
			Amount:        req.Amount,
			Charge:        charge,
			FinalAmount:   finalAmount,
			OrderID:       orderID,
			Status:        "Pending",
		}
		if err := tx.Create(&wd).Error; err != nil {
			return err
		}

		// Create corresponding debit transaction (Pending)
		msg := fmt.Sprintf("Penarikan ke %s %s", acc.Bank.Name, MaskAccountNumber(acc.AccountNumber))
		trx := models.Transaction{
			UserID:          uid,
			Amount:          req.Amount,
			Charge:          charge,
			OrderID:         orderID,
			TransactionFlow: "credit",
			TransactionType: "withdrawal",
			Message:         &msg,
			Status:          "Pending",
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, errInsufficientBalance) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Saldo tidak mencukupi"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}

	resp := map[string]interface{}{
		"withdrawal": map[string]interface{}{
			"id":             wd.ID,
			"order_id":       wd.OrderID,
			"amount":         wd.Amount,
			"charge":         wd.Charge,
			"final_amount":   wd.FinalAmount,
			"bank_name":      acc.Bank.Name,
			"account_name":   acc.AccountName,
			"account_number": MaskAccountNumber(acc.AccountNumber),
			"status":         wd.Status,
			"created_at":     wd.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}
	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Permintaan penarikan berhasil diproses",
		Data:    resp,
	})
}

// GET /api/users/withdrawal
func ListWithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	// Pagination: page + limit
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 50 {
		limit = 25
	}
	offset := (page - 1) * limit

	db := database.DB
	var withdrawals []models.Withdrawal
	if err := db.Where("user_id = ?", uid).Order("id DESC").Limit(limit).Offset(offset).Find(&withdrawals).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed to retrieve withdrawal data"})
		return
	}
	var resp []map[string]interface{}
	for _, wd := range withdrawals {
		var acc models.BankAccount
		var bank models.Bank
		db.First(&acc, wd.BankAccountID)
		db.First(&bank, acc.BankID)
		resp = append(resp, map[string]interface{}{
			"amount":          wd.Amount,
			"charge":          wd.Charge,
			"final_amount":    wd.FinalAmount,
			"order_id":        wd.OrderID,
			"status":          wd.Status,
			"withdrawal_time": wd.CreatedAt.Format("2006-01-02 15:04:05"),
			"account_name":    acc.AccountName,
			"account_number":  acc.AccountNumber,
			"bank_name":       bank.Name,
		})
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    resp,
	})
}

// Helpers

func CalculateWithdrawalCharge(amount float64) float64 {
	percent := getWithdrawalChargePercent()
	return round2(amount * (percent / 100.0))
}

func getWithdrawalChargePercent() float64 {
	s := os.Getenv("WITHDRAWAL_CHARGE_PERCENT")
	if s == "" {
		return 10.0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 10.0
	}
	return v
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

func MaskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 6 {
		return accountNumber
	}
	return accountNumber[:3] + "****" + accountNumber[len(accountNumber)-3:]
}
