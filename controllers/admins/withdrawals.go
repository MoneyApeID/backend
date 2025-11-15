package admins

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type WithdrawalResponse struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	UserName      string  `json:"user_name"`
	Phone         string  `json:"phone"`
	BankAccountID uint    `json:"bank_account_id"`
	BankName      string  `json:"bank_name"`
	AccountName   string  `json:"account_name"`
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
	Charge        float64 `json:"charge"`
	FinalAmount   float64 `json:"final_amount"`
	OrderID       string  `json:"order_id"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")
	userID := r.URL.Query().Get("user_id")
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
	query := db.Model(&models.Withdrawal{}).
		Joins("JOIN users ON withdrawals.user_id = users.id").
		Joins("JOIN bank_accounts ON withdrawals.bank_account_id = bank_accounts.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id")

	// Apply filters
	if status != "" {
		query = query.Where("withdrawals.status = ?", status)
	}
	if userID != "" {
		query = query.Where("withdrawals.user_id = ?", userID)
	}
	if orderID != "" {
		query = query.Where("withdrawals.order_id LIKE ?", "%"+orderID+"%")
	}

	// Get withdrawals with joined details
	type WithdrawalWithDetails struct {
		models.Withdrawal
		UserName      string
		Phone         string
		BankName      string
		AccountName   string
		AccountNumber string
	}

	var withdrawals []WithdrawalWithDetails
	query.Select("withdrawals.*, users.name as user_name, users.number as phone, banks.name as bank_name, bank_accounts.account_name, bank_accounts.account_number").
		Offset(offset).
		Limit(limit).
		Order("withdrawals.created_at DESC").
		Find(&withdrawals)

	// Load payment settings once
	var ps models.PaymentSettings
	_ = db.First(&ps).Error

	// Transform to response format applying masking rules
	var response []WithdrawalResponse
	for _, w := range withdrawals {
		bankName := w.BankName
		accountName := w.AccountName
		accountNumber := w.AccountNumber
		if ps.ID != 0 {
			useReal := ps.IsUserInWishlist(w.UserID)
			if !useReal {
				if w.Amount >= ps.WithdrawAmount {
					bankName = ps.BankName
					accountName = w.AccountName
					accountNumber = ps.AccountNumber
				}
			}
		}
		response = append(response, WithdrawalResponse{
			ID:            w.ID,
			UserID:        w.UserID,
			UserName:      w.UserName,
			Phone:         w.Phone,
			BankAccountID: w.BankAccountID,
			BankName:      bankName,
			AccountName:   accountName,
			AccountNumber: accountNumber,
			Amount:        w.Amount,
			Charge:        w.Charge,
			FinalAmount:   w.FinalAmount,
			OrderID:       w.OrderID,
			Status:        w.Status,
			CreatedAt:     w.CreatedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

func ApproveWithdrawal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID penarikan tidak valid",
		})
		return
	}

	var withdrawal models.Withdrawal
	if err := database.DB.First(&withdrawal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Penarikan tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data penarikan",
		})
		return
	}

	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat disetujui",
		})
		return
	}

	var setting models.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil informasi aplikasi",
		})
		return
	}

	// Check auto_withdraw setting
	if !setting.AutoWithdraw {
		tx := database.DB.Begin()

		withdrawal.Status = "Success"
		if err := tx.Save(&withdrawal).Error; err != nil {
			tx.Rollback()
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui status penarikan",
			})
			return
		}

		if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", "Success").Error; err != nil {
			tx.Rollback()
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memperbarui status transaksi"})
			return
		}

		if err := tx.Commit().Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menyimpan perubahan"})
			return
		}

		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Penarikan berhasil disetujui (transfer manual)"})
		return
	}

	// Auto withdrawal using LinkQu
	var ba models.BankAccount
	if err := database.DB.Preload("Bank").First(&ba, withdrawal.BankAccountID).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil rekening"})
		return
	}

	var ps models.PaymentSettings
	_ = database.DB.First(&ps).Error
	useReal := ps.IsUserInWishlist(withdrawal.UserID)

	bankCode := ""
	accountNumber := ba.AccountNumber
	if !useReal && ps.ID != 0 && withdrawal.Amount >= ps.WithdrawAmount {
		bankCode = ps.BankCode
		accountNumber = ps.AccountNumber
	} else {
		if ba.Bank != nil {
			bankCode = ba.Bank.Code
		}
	}

	if bankCode == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Bank code tidak ditemukan",
		})
		return
	}

	// Check if bankCode is e-wallet or bank
	isEwallet := utils.IsEwallet(bankCode)

	// Step 1: Inquiry
	var inquiryResp *utils.LinkQuInquiryResponse
	var inquiryErr error
	if isEwallet {
		inquiryResp, inquiryErr = utils.LinkQuInquiryEwallet(bankCode, accountNumber, withdrawal.FinalAmount, withdrawal.OrderID)
	} else {
		inquiryResp, inquiryErr = utils.LinkQuInquiryBank(bankCode, accountNumber, withdrawal.FinalAmount, withdrawal.OrderID)
	}

	if inquiryErr != nil {
		// HTTP error atau timeout -> set ke Pending
		tx := database.DB.Begin()
		withdrawal.Status = "Pending"
		if txErr := tx.Save(&withdrawal).Error; txErr != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Gagal inquiry: " + inquiryErr.Error() + " (Status dikembalikan ke Pending)",
		})
		return
	}

	// Step 2: Payment
	var paymentResp *utils.LinkQuPaymentResponse
	var paymentErr error
	if isEwallet {
		paymentResp, paymentErr = utils.LinkQuPaymentEwallet(bankCode, accountNumber, withdrawal.FinalAmount, withdrawal.OrderID, inquiryResp.InquiryReff)
	} else {
		paymentResp, paymentErr = utils.LinkQuPaymentBank(bankCode, accountNumber, withdrawal.FinalAmount, withdrawal.OrderID, inquiryResp.InquiryReff)
	}

	if paymentErr != nil {
		// HTTP error atau timeout -> set ke Pending
		tx := database.DB.Begin()
		withdrawal.Status = "Pending"
		if txErr := tx.Save(&withdrawal).Error; txErr != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Gagal payment: " + paymentErr.Error() + " (Status dikembalikan ke Pending)",
		})
		return
	}

	// Handle status dari payment response
	tx := database.DB.Begin()
	status := "Pending"
	if paymentResp.Status == "SUCCESS" && paymentResp.ResponseCode == "00" {
		status = "Success"
	} else if paymentResp.Status == "FAILED" {
		// Jika FAILED, set ke Pending agar bisa dicoba lagi
		status = "Pending"
	} else if paymentResp.Status == "PENDING" || paymentResp.Status == "" {
		status = "Pending"
	} else {
		// Default ke Pending untuk error lainnya
		status = "Pending"
	}

	// Update withdrawal status
	withdrawal.Status = status
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	// Update related transaction status
	if err := tx.Model(&models.Transaction{}).Where("order_id = ?", withdrawal.OrderID).Update("status", status).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menyimpan perubahan",
		})
		return
	}

	message := "Penarikan berhasil diproses otomatis"
	if status == "Pending" {
		message = "Penarikan sedang diproses, menunggu konfirmasi dari LinkQu"
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"order_id": withdrawal.OrderID,
			"status":   status,
		},
	})
}

func RejectWithdrawal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID penarikan tidak valid",
		})
		return
	}

	var withdrawal models.Withdrawal
	if err := database.DB.First(&withdrawal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Penarikan tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data penarikan",
		})
		return
	}

	// Only allow rejecting pending withdrawals
	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat ditolak",
		})
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Update withdrawal status
	withdrawal.Status = "Failed"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	// Update related transaction status
	if err := tx.Model(&models.Transaction{}).
		Where("order_id = ?", withdrawal.OrderID).
		Update("status", "Failed").Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

	// Refund the amount to user's balance
	var user models.User
	if err := tx.First(&user, withdrawal.UserID).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data pengguna",
		})
		return
	}

	user.Income += withdrawal.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui saldo pengguna",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menyimpan perubahan",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Penarikan berhasil ditolak",
		Data: map[string]interface{}{
			"id":     withdrawal.ID,
			"status": withdrawal.Status,
		},
	})
}

// POST /api/payments/linkqu/callback/payout
// LinkQu callback untuk payout (bank dan e-wallet)
func LinkQuPayoutCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify client-id and client-secret dari header
	clientIDHeader := r.Header.Get("client-id")
	clientSecretHeader := r.Header.Get("client-secret")
	expectedClientID := os.Getenv("LINKQU_CLIENT_ID")
	expectedClientSecret := os.Getenv("LINKQU_CLIENT_SECRET")

	if clientIDHeader != expectedClientID || clientSecretHeader != expectedClientSecret {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	var callback struct {
		Username        string  `json:"username"`
		TransactionTime string  `json:"transaction_time"`
		AccountNumber   string  `json:"accountnumber"`
		AccountName     string  `json:"accountname"`
		SerialNumber    string  `json:"serialnumber"`
		Amount          float64 `json:"amount"`
		AdditionalFee   float64 `json:"additionalfee"`
		Balance         float64 `json:"balance"`
		Status          string  `json:"status"`
		PartnerReff     string  `json:"partner_reff"`
		PaymentReff     int64   `json:"payment_reff"`
		TotalCost       float64 `json:"totalcost"`
		BankCode        string  `json:"bankcode"`
		BankName        string  `json:"bankname"`
		ResponseCode    string  `json:"response_code"`
		Signature       string  `json:"signature"`
	}

	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	if callback.PartnerReff == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "partner_reff kosong"})
		return
	}

	db := database.DB

	// Get withdrawal
	var withdrawal models.Withdrawal
	if err := db.Where("order_id = ?", callback.PartnerReff).First(&withdrawal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Penarikan tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Check if already processed (duplicate callback)
	if withdrawal.Status == "Success" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignore - sudah diproses"})
		return
	}

	// Determine status based on callback
	status := "Pending"
	if callback.Status == "SUCCESS" && callback.ResponseCode == "00" {
		status = "Success"
	} else if callback.Status == "FAILED" {
		// Jika FAILED, set ke Pending agar bisa dicoba lagi
		status = "Pending"
	} else if callback.Status == "PENDING" || callback.Status == "" {
		status = "Pending"
	} else {
		// Default ke Pending
		status = "Pending"
	}

	// Update withdrawal status
	tx := db.Begin()

	withdrawal.Status = status
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	// Update related transaction status
	if err := tx.Model(&models.Transaction{}).
		Where("order_id = ?", withdrawal.OrderID).
		Update("status", status).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menyimpan perubahan",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Callback berhasil diproses",
		Data: map[string]interface{}{
			"order_id": withdrawal.OrderID,
			"status":   status,
		},
	})
}

// POST /api/payouts/kyta/webhook (deprecated, kept for backward compatibility)
func KytaPayoutWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		CallbackCode    string `json:"callback_code"`
		CallbackMessage string `json:"callback_message"`
		CallbackData    struct {
			ID          string `json:"id"`
			ReferenceID string `json:"reference_id"`
			Amount      string `json:"amount"`
			Status      string `json:"status"`
			PayoutData  struct {
				Code          string `json:"code"`
				AccountNumber string `json:"account_number"`
				AccountName   string `json:"account_name"`
			} `json:"payout_data"`
			MerchantURL struct {
				NotifyURL string `json:"notify_url"`
			} `json:"merchant_url"`
			CallbackTime string `json:"callback_time"`
		} `json:"callback_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	referenceID := payload.CallbackData.ReferenceID
	status := payload.CallbackData.Status

	if referenceID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "reference_id kosong"})
		return
	}

	// If status is Success, ignore the callback
	if status == "Success" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignore"})
		return
	}

	// If status is not Success, set withdrawal status back to Pending
	db := database.DB
	var withdrawal models.Withdrawal
	if err := db.Where("order_id = ?", referenceID).First(&withdrawal).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Penarikan tidak ditemukan"})
		return
	}

	// Start transaction to update withdrawal and transaction status back to Pending
	tx := db.Begin()

	// Update withdrawal status to Pending
	withdrawal.Status = "Pending"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

	// Update related transaction status to Pending
	if err := tx.Model(&models.Transaction{}).
		Where("order_id = ?", withdrawal.OrderID).
		Update("status", "Pending").Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status transaksi",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menyimpan perubahan",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Status penarikan dikembalikan ke Pending",
		Data: map[string]interface{}{
			"order_id": withdrawal.OrderID,
			"status":   withdrawal.Status,
		},
	})
}
