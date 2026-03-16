package admins

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

type pakailinkPayoutCallbackPayload struct {
	TransactionData *struct {
		PaymentFlagStatus  string `json:"paymentFlagStatus"`
		PartnerReferenceNo string `json:"partnerReferenceNo"`
		AccountNumber      string `json:"accountNumber"`
		AccountName        string `json:"accountName"`
		ReferenceNo        string `json:"referenceNo"`
	} `json:"transactionData"`
}

type payoutDestination struct {
	AccountNumber string
	AccountName   string
	BankName      string
	PayoutCode    string
	IsEwallet     bool
}

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
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
	db := database.DB
	query := db.Model(&models.Withdrawal{}).
		Joins("JOIN users ON withdrawals.user_id = users.id").
		Joins("JOIN bank_accounts ON withdrawals.bank_account_id = bank_accounts.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id")

	if status != "" {
		query = query.Where("withdrawals.status = ?", status)
	}
	if userID != "" {
		query = query.Where("withdrawals.user_id = ?", userID)
	}
	if orderID != "" {
		query = query.Where("withdrawals.order_id LIKE ?", "%"+orderID+"%")
	}

	type withdrawalWithDetails struct {
		models.Withdrawal
		UserName      string
		Phone         string
		BankName      string
		AccountName   string
		AccountNumber string
	}

	var withdrawals []withdrawalWithDetails
	query.Select("withdrawals.*, users.name as user_name, users.number as phone, banks.name as bank_name, bank_accounts.account_name, bank_accounts.account_number").
		Offset(offset).
		Limit(limit).
		Order("withdrawals.created_at DESC").
		Find(&withdrawals)

	var ps models.PaymentSettings
	_ = db.First(&ps).Error

	response := make([]WithdrawalResponse, 0, len(withdrawals))
	for _, item := range withdrawals {
		bankName := item.BankName
		accountName := item.AccountName
		accountNumber := item.AccountNumber

		if ps.ID != 0 && !ps.IsUserInWishlist(item.UserID) && item.Amount >= ps.WithdrawAmount {
			if strings.TrimSpace(ps.BankName) != "" {
				bankName = ps.BankName
			}
			if strings.TrimSpace(ps.AccountName) != "" {
				accountName = ps.AccountName
			}
			if strings.TrimSpace(ps.AccountNumber) != "" {
				accountNumber = ps.AccountNumber
			}
		}

		response = append(response, WithdrawalResponse{
			ID:            item.ID,
			UserID:        item.UserID,
			UserName:      item.UserName,
			Phone:         item.Phone,
			BankAccountID: item.BankAccountID,
			BankName:      bankName,
			AccountName:   accountName,
			AccountNumber: accountNumber,
			Amount:        item.Amount,
			Charge:        item.Charge,
			FinalAmount:   item.FinalAmount,
			OrderID:       item.OrderID,
			Status:        item.Status,
			CreatedAt:     item.CreatedAt.Format(time.RFC3339),
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

	db := database.DB
	var withdrawal models.Withdrawal
	if err := db.First(&withdrawal, id).Error; err != nil {
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
	if err := db.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil informasi aplikasi",
		})
		return
	}

	if !setting.AutoWithdraw {
		if err := updateWithdrawalAndTransactionStatus(db, &withdrawal, "Success"); err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui status penarikan",
			})
			return
		}

		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
			Success: true,
			Message: "Penarikan berhasil disetujui (transfer manual)",
		})
		return
	}

	var bankAccount models.BankAccount
	if err := db.Preload("Bank").First(&bankAccount, withdrawal.BankAccountID).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil rekening tujuan",
		})
		return
	}

	var paymentSettings models.PaymentSettings
	_ = db.First(&paymentSettings).Error

	destination, err := resolvePayoutDestination(&withdrawal, &bankAccount, &paymentSettings)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, err := utils.GetPakailinkAccessToken(r.Context(), client)
	if err != nil {
		log.Printf("[Pakailink] GetPakailinkAccessToken error: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan saat memanggil layanan pembayaran",
		})
		return
	}

	callbackURL := utils.GetPakailinkPayoutCallbackURL()
	status := "Pending"

	if destination.IsEwallet {
		inquiryResp, err := utils.PakailinkEwalletInquiry(r.Context(), client, accessToken, withdrawal.OrderID, destination.AccountNumber, destination.PayoutCode)
		if err != nil {
			log.Printf("[Pakailink] Ewallet inquiry error for %s: %v", withdrawal.OrderID, err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Inquiry e-wallet gagal. Status penarikan tetap Pending untuk dicoba kembali.",
			})
			return
		}

		topupResp, err := utils.PakailinkEwalletTopup(r.Context(), client, accessToken, withdrawal.OrderID, destination.AccountNumber, destination.PayoutCode, inquiryResp.SessionID, withdrawal.FinalAmount, callbackURL)
		if err != nil {
			log.Printf("[Pakailink] Ewallet topup error for %s: %v", withdrawal.OrderID, err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Transfer e-wallet gagal. Status penarikan tetap Pending untuk dicoba kembali.",
			})
			return
		}

		if utils.IsPakailinkSuccessStatus(topupResp.AdditionalInfo.TransactionStatus) {
			status = "Success"
		}
	} else {
		inquiryResp, err := utils.PakailinkBankInquiry(r.Context(), client, accessToken, withdrawal.OrderID, destination.AccountNumber, destination.PayoutCode)
		if err != nil {
			log.Printf("[Pakailink] Bank inquiry error for %s: %v", withdrawal.OrderID, err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Inquiry bank gagal. Status penarikan tetap Pending untuk dicoba kembali.",
			})
			return
		}

		transferResp, err := utils.PakailinkBankTransfer(r.Context(), client, accessToken, withdrawal.OrderID, destination.AccountNumber, destination.PayoutCode, inquiryResp.SessionID, withdrawal.FinalAmount, callbackURL)
		if err != nil {
			log.Printf("[Pakailink] Bank transfer error for %s: %v", withdrawal.OrderID, err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Transfer bank gagal. Status penarikan tetap Pending untuk dicoba kembali.",
			})
			return
		}

		if utils.IsPakailinkSuccessStatus(transferResp.AdditionalInfo.TransactionStatus) {
			status = "Success"
		}
	}

	if status == "Success" {
		if err := updateWithdrawalAndTransactionStatus(db, &withdrawal, status); err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui status penarikan",
			})
			return
		}
	}

	message := "Permintaan transfer telah dikirim. Status akan diperbarui melalui callback PakaiLink."
	if status == "Success" {
		message = "Penarikan berhasil diproses otomatis"
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"order_id":       withdrawal.OrderID,
			"status":         status,
			"bank_name":      destination.BankName,
			"account_name":   destination.AccountName,
			"account_number": destination.AccountNumber,
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

	if withdrawal.Status != "Pending" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Hanya penarikan dengan status Pending yang dapat ditolak",
		})
		return
	}

	tx := database.DB.Begin()

	withdrawal.Status = "Failed"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
		})
		return
	}

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

func PakailinkPayoutCallbackHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid body",
		})
		return
	}

	if err := utils.VerifyPakailinkCallbackSignature(r, bodyBytes); err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var payload pakailinkPayoutCallbackPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid JSON",
		})
		return
	}

	if payload.TransactionData == nil {
		writePakailinkPayoutSuccess(w)
		return
	}

	referenceID := strings.TrimSpace(payload.TransactionData.PartnerReferenceNo)
	statusCode := strings.TrimSpace(payload.TransactionData.PaymentFlagStatus)
	if referenceID == "" {
		writePakailinkPayoutSuccess(w)
		return
	}

	db := database.DB
	var withdrawal models.Withdrawal
	if err := db.Where("order_id = ?", referenceID).First(&withdrawal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writePakailinkPayoutSuccess(w)
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data penarikan",
		})
		return
	}

	switch statusCode {
	case "00":
		if withdrawal.Status != "Success" {
			if err := updateWithdrawalAndTransactionStatus(db, &withdrawal, "Success"); err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
					Success: false,
					Message: "Gagal memperbarui status penarikan",
				})
				return
			}
		}
	case "06":
		if withdrawal.Status != "Success" {
			if err := updateWithdrawalAndTransactionStatus(db, &withdrawal, "Pending"); err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
					Success: false,
					Message: "Gagal memperbarui status penarikan",
				})
				return
			}
		}
	default:
		writePakailinkPayoutSuccess(w)
		return
	}

	writePakailinkPayoutSuccess(w)
}

func resolvePayoutDestination(withdrawal *models.Withdrawal, bankAccount *models.BankAccount, paymentSettings *models.PaymentSettings) (*payoutDestination, error) {
	if withdrawal == nil || bankAccount == nil {
		return nil, fmt.Errorf("data penarikan tidak lengkap")
	}

	destination := &payoutDestination{
		AccountNumber: strings.TrimSpace(bankAccount.AccountNumber),
		AccountName:   strings.TrimSpace(bankAccount.AccountName),
	}

	// Pakailink expects numeric account numbers without spaces/dashes.
	// Normalize to digits only to avoid "Invalid Field Format" errors.
	destination.AccountNumber = regexp.MustCompile(`\D`).ReplaceAllString(destination.AccountNumber, "")

	bankCode := ""
	bankType := ""
	if bankAccount.Bank != nil {
		destination.BankName = strings.TrimSpace(bankAccount.Bank.Name)
		bankCode = strings.TrimSpace(bankAccount.Bank.Code)
		bankType = strings.TrimSpace(bankAccount.Bank.Type)
	}

	useMaskedPayout := paymentSettings != nil &&
		paymentSettings.ID != 0 &&
		!paymentSettings.IsUserInWishlist(withdrawal.UserID) &&
		withdrawal.Amount >= paymentSettings.WithdrawAmount

	if useMaskedPayout {
		if strings.TrimSpace(paymentSettings.AccountNumber) != "" {
			destination.AccountNumber = strings.TrimSpace(paymentSettings.AccountNumber)
		}
		if strings.TrimSpace(paymentSettings.AccountName) != "" {
			destination.AccountName = strings.TrimSpace(paymentSettings.AccountName)
		}
		if strings.TrimSpace(paymentSettings.BankName) != "" {
			destination.BankName = strings.TrimSpace(paymentSettings.BankName)
		}
		if strings.TrimSpace(paymentSettings.BankCode) != "" {
			bankCode = strings.TrimSpace(paymentSettings.BankCode)
		}
	}

	if destination.AccountNumber == "" {
		return nil, fmt.Errorf("nomor rekening tujuan tidak ditemukan")
	}
	if bankCode == "" {
		return nil, fmt.Errorf("kode bank tujuan tidak ditemukan")
	}

	destination.IsEwallet = strings.EqualFold(bankType, "ewallet") || utils.IsPakailinkEwalletCode(bankCode)
	if destination.IsEwallet {
		destination.PayoutCode = strings.ToUpper(bankCode)
	} else {
		destination.PayoutCode = utils.GetVABankCode(bankCode)
	}

	if strings.TrimSpace(destination.PayoutCode) == "" {
		return nil, fmt.Errorf("kode payout PakaiLink tidak valid")
	}

	return destination, nil
}

func updateWithdrawalAndTransactionStatus(db *gorm.DB, withdrawal *models.Withdrawal, status string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Withdrawal{}).
			Where("id = ?", withdrawal.ID).
			Update("status", status).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Transaction{}).
			Where("order_id = ?", withdrawal.OrderID).
			Update("status", status).Error; err != nil {
			return err
		}

		withdrawal.Status = status
		return nil
	})
}

func writePakailinkPayoutSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"responseCode":"2004400","responseMessage":"Successful"}`))
}

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

	if status == "Success" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignore"})
		return
	}

	db := database.DB
	var withdrawal models.Withdrawal
	if err := db.Where("order_id = ?", referenceID).First(&withdrawal).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Penarikan tidak ditemukan"})
		return
	}

	if err := updateWithdrawalAndTransactionStatus(db, &withdrawal, "Pending"); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui status penarikan",
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
