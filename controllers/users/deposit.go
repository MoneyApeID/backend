package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateDepositRequest struct {
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentChannel string  `json:"payment_channel"`
}

type pakailinkDepositCallbackPayload struct {
	TransactionData *struct {
		PartnerReferenceNo string `json:"partnerReferenceNo"`
		CallbackType       string `json:"callbackType"`
		PaymentFlagStatus  string `json:"paymentFlagStatus"`
		PaymentFlagReason  *struct {
			English   string `json:"english"`
			Indonesia string `json:"indonesia"`
		} `json:"paymentFlagReason"`
	} `json:"transactionData"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	CallbackType               string `json:"callbackType"`
	LatestTransactionStatus    string `json:"latestTransactionStatus"`
	TransactionStatusDesc      string `json:"transactionStatusDesc"`
}

// POST /api/users/deposits
func CreateDepositHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid JSON"})
		return
	}

	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	method := strings.ToUpper(strings.TrimSpace(req.PaymentMethod))
	channel := strings.ToUpper(strings.TrimSpace(req.PaymentChannel))
	amount := utils.RoundFloat(req.Amount, 2)

	if method != "QRIS" && method != "BANK" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Silahkan pilih metode pembayaran"})
		return
	}
	if amount <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Jumlah isi ulang tidak valid"})
		return
	}
	// Read min deposit from admin settings
	var setting models.Setting
	if err := database.DB.First(&setting).Error; err == nil && setting.MinDeposit > 0 {
		if amount < setting.MinDeposit {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: fmt.Sprintf("Jumlah deposit minimal Rp %.0f", setting.MinDeposit)})
			return
		}
	}
	if method == "QRIS" {
		if amount > 10000000 {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Jumlah pembayaran QRIS maksimal Rp 10.000.000"})
			return
		}
	} else {
		if amount > 100000000 {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Jumlah pembayaran BANK maksimal Rp 100.000.000"})
			return
		}
		if !utils.IsSupportedVABankChannel(channel) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank tidak valid"})
			return
		}
	}

	db := database.DB
	var user models.User
	if err := db.Where("id = ?", uid).First(&user).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
		return
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	accessToken, err := utils.GetPakailinkAccessToken(r.Context(), httpClient)
	if err != nil {
		log.Printf("[Pakailink] GetPakailinkAccessToken error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat memanggil layanan pembayaran"})
		return
	}

	orderID := utils.GenerateOrderID(uid)
	var paymentCode string
	var expiredAt time.Time

	if method == "QRIS" {
		qrResp, err := utils.CreatePakailinkQRIS(r.Context(), httpClient, accessToken, orderID, amount)
		if err != nil {
			log.Printf("[Pakailink] CreatePakailinkQRIS error: %v", err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat memanggil layanan pembayaran"})
			return
		}
		paymentCode = strings.TrimSpace(qrResp.QRContent)
		expiredAt = parsePakailinkTimestampOrDefault(qrResp.ValidityPeriod, time.Now().Add(24*time.Hour))
	} else {
		customerNo := fmt.Sprintf("%d%010d", uid, time.Now().UnixNano()%10000000000)
		userName := strings.TrimSpace(user.Name)
		if userName == "" {
			userName = "Pelanggan"
		}
		vaName := fmt.Sprintf("%s - Money Rich", userName)
		bankCode := utils.GetVABankCode(channel)

		vaResp, err := utils.CreatePakailinkVA(r.Context(), httpClient, accessToken, orderID, customerNo, vaName, amount, bankCode)
		if err != nil {
			log.Printf("[Pakailink] CreatePakailinkVA error: %v", err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Terjadi kesalahan saat memanggil layanan pembayaran"})
			return
		}
		paymentCode = strings.TrimSpace(vaResp.VirtualAccountData.VirtualAccountNo)
		expiredAt = parsePakailinkTimestampOrDefault(vaResp.VirtualAccountData.ExpiredDate, time.Now().Add(24*time.Hour))
	}

	if paymentCode == "" {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Kode pembayaran tidak tersedia"})
		return
	}

	deposit := models.Deposit{
		UserID:        uid,
		Amount:        amount,
		OrderID:       orderID,
		PaymentMethod: method,
		Status:        "Pending",
		ExpiredAt:     expiredAt,
	}
	deposit.PaymentCode = &paymentCode
	if method == "BANK" {
		deposit.PaymentChannel = &channel
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&deposit).Error; err != nil {
			return err
		}

		message := "Isi ulang saldo menggunakan " + method
		trx := models.Transaction{
			UserID:          uid,
			Amount:          amount,
			Charge:          0,
			OrderID:         orderID,
			TransactionFlow: "debit",
			TransactionType: "deposit",
			Message:         &message,
			Status:          "Pending",
		}
		return tx.Create(&trx).Error
	}); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat pembayaran"})
		return
	}

	responseData := map[string]interface{}{
		"order_id":       deposit.OrderID,
		"amount":         deposit.Amount,
		"payment_method": deposit.PaymentMethod,
		"payment_channel": func() interface{} {
			if deposit.PaymentChannel == nil {
				return nil
			}
			return *deposit.PaymentChannel
		}(),
		"payment_code": paymentCode,
		"expired_at":   deposit.ExpiredAt.UTC().Format(time.RFC3339),
		"status":       deposit.Status,
	}
	if method == "QRIS" {
		responseData["payment_url"] = buildQRRenderURL(paymentCode)
	}

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{Success: true, Message: "Isi ulang berhasil dibuat", Data: responseData})
}

// GET /api/users/payment/{order_id}
func GetDepositDetailsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var orderID string
	if len(parts) >= 3 {
		orderID = parts[len(parts)-1]
	}
	if orderID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Order ID tidak valid"})
		return
	}

	db := database.DB
	var deposit models.Deposit
	if err := db.Where("order_id = ? AND user_id = ?", orderID, uid).First(&deposit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Data isi ulang tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	_ = refreshDepositStatusFromPakailink(r.Context(), db, &deposit)
	_ = db.Where("id = ?", deposit.ID).First(&deposit).Error

	resp := map[string]interface{}{
		"order_id":       deposit.OrderID,
		"amount":         deposit.Amount,
		"payment_method": deposit.PaymentMethod,
		"payment_code": func() interface{} {
			if deposit.PaymentCode == nil {
				return nil
			}
			return *deposit.PaymentCode
		}(),
		"payment_channel": func() interface{} {
			if deposit.PaymentChannel == nil {
				return nil
			}
			return *deposit.PaymentChannel
		}(),
		"status":     deposit.Status,
		"expired_at": deposit.ExpiredAt.UTC().Format(time.RFC3339),
	}

	if deposit.PaymentMethod == "QRIS" && deposit.PaymentCode != nil && *deposit.PaymentCode != "" {
		resp["payment_url"] = buildQRRenderURL(*deposit.PaymentCode)
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: resp})
}

// GET /api/users/deposits
func ListDepositsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

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
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	db := database.DB
	query := db.Model(&models.Deposit{}).Where("user_id = ?", uid)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	var deposits []models.Deposit
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&deposits).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Mark expired local records without forcing a full inquiry on list endpoints.
	now := time.Now()
	for i := range deposits {
		if deposits[i].Status == "Pending" && (deposits[i].ExpiredAt.Before(now) || deposits[i].ExpiredAt.Equal(now)) {
			_ = db.Model(&models.Deposit{}).
				Where("id = ? AND status = ?", deposits[i].ID, "Pending").
				Update("status", "Expired").Error
			deposits[i].Status = "Expired"
		}
	}

	type depositDTO struct {
		ID             uint    `json:"id"`
		OrderID        string  `json:"order_id"`
		Amount         float64 `json:"amount"`
		PaymentMethod  string  `json:"payment_method"`
		PaymentChannel *string `json:"payment_channel,omitempty"`
		Status         string  `json:"status"`
		ExpiredAt      string  `json:"expired_at"`
		CreatedAt      string  `json:"created_at"`
		UpdatedAt      string  `json:"updated_at"`
	}

	items := make([]depositDTO, 0, len(deposits))
	for _, deposit := range deposits {
		items = append(items, depositDTO{
			ID:             deposit.ID,
			OrderID:        deposit.OrderID,
			Amount:         deposit.Amount,
			PaymentMethod:  deposit.PaymentMethod,
			PaymentChannel: deposit.PaymentChannel,
			Status:         deposit.Status,
			ExpiredAt:      deposit.ExpiredAt.UTC().Format(time.RFC3339),
			CreatedAt:      deposit.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:      deposit.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: map[string]interface{}{
			"deposits": items,
			"total":    total,
		},
	})
}

// POST /api/callback/payments
func PakailinkCallbackHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid body"})
		return
	}

	if err := utils.VerifyPakailinkCallbackSignature(r, bodyBytes); err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: err.Error()})
		return
	}

	var payload pakailinkDepositCallbackPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	var callbackType, partnerReferenceNo, transactionStatus, transactionStatusDesc string
	if payload.TransactionData != nil {
		callbackType = strings.TrimSpace(payload.TransactionData.CallbackType)
		partnerReferenceNo = strings.TrimSpace(payload.TransactionData.PartnerReferenceNo)
		transactionStatus = strings.TrimSpace(payload.TransactionData.PaymentFlagStatus)
		if payload.TransactionData.PaymentFlagReason != nil {
			transactionStatusDesc = strings.TrimSpace(firstNonEmptyString(
				payload.TransactionData.PaymentFlagReason.English,
				payload.TransactionData.PaymentFlagReason.Indonesia,
			))
		}
	} else {
		callbackType = strings.TrimSpace(payload.CallbackType)
		partnerReferenceNo = strings.TrimSpace(payload.OriginalPartnerReferenceNo)
		transactionStatus = strings.TrimSpace(payload.LatestTransactionStatus)
		transactionStatusDesc = strings.TrimSpace(payload.TransactionStatusDesc)
	}

	if callbackType != "" &&
		!strings.EqualFold(callbackType, "payment") &&
		!strings.EqualFold(callbackType, "settlement") {
		writePakailinkPaymentCallbackSuccess(w)
		return
	}
	if partnerReferenceNo == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "partnerReferenceNo kosong"})
		return
	}

	db := database.DB
	var deposit models.Deposit
	if err := db.Where("order_id = ?", partnerReferenceNo).First(&deposit).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Deposit tidak ditemukan"})
		return
	}

	log.Printf(
		"[Pakailink] payment callback order=%s callback_type=%s status=%s desc=%s",
		partnerReferenceNo,
		callbackType,
		transactionStatus,
		transactionStatusDesc,
	)

	switch {
	case utils.IsPakailinkSuccessState(transactionStatus, transactionStatusDesc):
		if err := db.Transaction(func(tx *gorm.DB) error {
			return processDepositSuccess(tx, &deposit)
		}); err != nil {
			log.Printf("[Pakailink] callback success processing error: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memproses callback"})
			return
		}
	case utils.IsPakailinkPendingState(transactionStatus, transactionStatusDesc):
		// Keep current status as-is.
	case utils.IsPakailinkFailedState(transactionStatus, transactionStatusDesc):
		if deposit.Status != "Success" {
			_ = db.Model(&models.Deposit{}).
				Where("id = ? AND status <> ?", deposit.ID, "Success").
				Update("status", "Failed").Error
			_ = db.Model(&models.Transaction{}).
				Where("order_id = ? AND transaction_type = ?", deposit.OrderID, "deposit").
				Update("status", "Failed").Error
		}
	default:
		log.Printf("[Pakailink] keeping deposit pending for unclassified callback status order=%s status=%s desc=%s", partnerReferenceNo, transactionStatus, transactionStatusDesc)
	}

	writePakailinkPaymentCallbackSuccess(w)
}

func refreshDepositStatusFromPakailink(ctx context.Context, db *gorm.DB, deposit *models.Deposit) error {
	if deposit == nil {
		return nil
	}
	if deposit.Status == "Success" {
		return nil
	}
	if deposit.ExpiredAt.Before(time.Now()) || deposit.ExpiredAt.Equal(time.Now()) {
		if deposit.Status == "Pending" {
			_ = db.Model(&models.Deposit{}).Where("id = ? AND status = ?", deposit.ID, "Pending").Update("status", "Expired").Error
			deposit.Status = "Expired"
		}
		return nil
	}
	if deposit.Status != "Pending" {
		return nil
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	accessToken, err := utils.GetPakailinkAccessToken(ctx, httpClient)
	if err != nil {
		return err
	}

	var statusValue string
	var statusDesc string

	if deposit.PaymentMethod == "BANK" {
		statusResp, err := utils.InquiryPakailinkVAStatus(ctx, httpClient, accessToken, deposit.OrderID)
		if err != nil {
			return err
		}
		statusValue = strings.TrimSpace(statusResp.LatestTransactionStatus)
		statusDesc = strings.TrimSpace(statusResp.TransactionStatusDesc)
	} else if deposit.PaymentMethod == "QRIS" {
		statusResp, err := utils.InquiryPakailinkQRStatus(ctx, httpClient, accessToken, deposit.OrderID)
		if err != nil {
			return err
		}
		statusValue = strings.TrimSpace(statusResp.LatestTransactionStatus)
		statusDesc = strings.TrimSpace(statusResp.TransactionStatusDesc)
	}

	log.Printf(
		"[Pakailink] inquiry deposit order=%s method=%s status=%s desc=%s",
		deposit.OrderID,
		deposit.PaymentMethod,
		statusValue,
		statusDesc,
	)

	switch {
	case utils.IsPakailinkSuccessState(statusValue, statusDesc):
		return db.Transaction(func(tx *gorm.DB) error {
			return processDepositSuccess(tx, deposit)
		})
	case utils.IsPakailinkFailedState(statusValue, statusDesc):
		if err := db.Model(&models.Deposit{}).
			Where("id = ? AND status = ?", deposit.ID, "Pending").
			Update("status", "Failed").Error; err == nil {
			deposit.Status = "Failed"
		}
		_ = db.Model(&models.Transaction{}).
			Where("order_id = ? AND transaction_type = ?", deposit.OrderID, "deposit").
			Update("status", "Failed").Error
		return nil
	case utils.IsPakailinkPendingState(statusValue, statusDesc):
		return nil
	default:
		log.Printf("[Pakailink] unclassified inquiry status kept pending order=%s status=%s desc=%s", deposit.OrderID, statusValue, statusDesc)
		return nil
	}
}

func processDepositSuccess(tx *gorm.DB, deposit *models.Deposit) error {
	var lockedDeposit models.Deposit
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", deposit.ID).
		First(&lockedDeposit).Error; err != nil {
		return err
	}

	if lockedDeposit.Status == "Success" {
		*deposit = lockedDeposit
		return nil
	}

	if err := tx.Model(&models.Deposit{}).Where("id = ?", lockedDeposit.ID).Update("status", "Success").Error; err != nil {
		return err
	}
	if err := tx.Model(&models.Transaction{}).
		Where("order_id = ? AND transaction_type = ?", lockedDeposit.OrderID, "deposit").
		Update("status", "Success").Error; err != nil {
		return err
	}

	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id, balance, spin_ticket").
		Where("id = ?", lockedDeposit.UserID).
		First(&user).Error; err != nil {
		return err
	}

	newBalance := utils.RoundFloat(user.Balance+lockedDeposit.Amount, 2)
	if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("balance", newBalance).Error; err != nil {
		return err
	}

	if lockedDeposit.Amount >= 100000 {
		spinTicketsToAdd := uint(1)
		if lockedDeposit.Amount >= 500000 {
			spinTicketsToAdd = 2
		}

		if user.SpinTicket == nil {
			if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("spin_ticket", spinTicketsToAdd).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.User{}).Where("id = ?", user.ID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket + ?", spinTicketsToAdd)).Error; err != nil {
				return err
			}
		}
	}

	lockedDeposit.Status = "Success"
	*deposit = lockedDeposit
	return nil
}

func parsePakailinkTimestampOrDefault(value string, fallback time.Time) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000-07:00",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return fallback
}

func buildQRRenderURL(payload string) string {
	return "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" + neturl.QueryEscape(payload)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func writePakailinkPaymentCallbackSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"responseCode":"2002800","responseMessage":"Successful"}`))
}
