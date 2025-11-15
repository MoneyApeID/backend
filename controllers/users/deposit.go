package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

type linkQuResponse struct {
	Amount         float64 `json:"amount"`
	Expired        string  `json:"expired"`
	Status         string  `json:"status"`
	ResponseCode   string  `json:"response_code"`
	ResponseDesc   string  `json:"response_desc"`
	ImageQRIS      string  `json:"imageqris"`
	QRISText       string  `json:"qris_text"`
	VirtualAccount string  `json:"virtual_account"`
	PartnerReff    string  `json:"partner_reff"`
}

var bankCodeMap = map[string]string{
	"BCA":     "014",
	"BRI":     "002",
	"BNI":     "009",
	"MANDIRI": "008",
	"PERMATA": "013",
	"BNC":     "490",
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

	if method != "QRIS" && method != "BANK" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Silahkan pilih metode pembayaran"})
		return
	}
	if method == "BANK" {
		if _, ok := bankCodeMap[channel]; !ok {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank tidak valid"})
			return
		}
	}

	amount := utils.RoundFloat(req.Amount, 2)
	if amount <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Jumlah isi ulang tidak valid"})
		return
	}

	linkQuConfig, err := getLinkQuConfig()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: err.Error()})
		return
	}

	db := database.DB
	var user models.User
	if err := db.Where("id = ?", uid).First(&user).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
		return
	}

	orderID := utils.GenerateOrderID(uid)
	now := time.Now().In(linkQuConfig.location)
	expiredAt := now.Add(15 * time.Minute)
	expiredStr := expiredAt.Format("20060102150405")

	customerPhone := normalizePhone(user.Number)
	customerEmail := fmt.Sprintf("%s@gmail.com", strings.TrimSpace(user.Number))
	customerID := strings.TrimSpace(user.Name) + strings.TrimSpace(user.Number)
	if strings.TrimSpace(customerID) == "" {
		customerID = fmt.Sprintf("user-%d", uid)
	}

	body := map[string]interface{}{
		"amount":         int64(utils.RoundFloat(amount, 0)),
		"partner_reff":   orderID,
		"customer_id":    customerID,
		"customer_name":  strings.TrimSpace(user.Name),
		"expired":        expiredStr,
		"username":       linkQuConfig.username,
		"pin":            linkQuConfig.pin,
		"customer_phone": customerPhone,
		"customer_email": customerEmail,
		"url_callback":   linkQuConfig.callbackURL,
	}

	var endpoint string
	if method == "QRIS" {
		endpoint = "/linkqu-partner/transaction/create/qris"
	} else {
		endpoint = "/linkqu-partner/transaction/create/va"
		body["bank_code"] = bankCodeMap[channel]
		body["remark"] = fmt.Sprintf("Deposit Rp %.0f", amount)
	}

	resp, err := callLinkQu(linkQuConfig, endpoint, body)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Terjadi kesalahan pada sisi pembayaran, Tim kami akan segera menangani masalah tersebut"})
		return
	}

	expiredTime, err := parseLinkQuTime(resp.Expired, linkQuConfig.location)
	if err != nil {
		expiredTime = expiredAt
	}

	var paymentCode string
	if method == "QRIS" {
		paymentCode = strings.TrimSpace(resp.QRISText)
	} else {
		paymentCode = strings.TrimSpace(resp.VirtualAccount)
	}
	if paymentCode == "" {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan pada sisi pembayaran, Tim kami akan segera menangani masalah tersebut"})
		return
	}

	deposit := models.Deposit{
		UserID:        uid,
		Amount:        amount,
		OrderID:       orderID,
		PaymentMethod: method,
		Status:        "Pending",
		ExpiredAt:     expiredTime,
	}
	deposit.PaymentCode = &paymentCode
	if method == "BANK" {
		deposit.PaymentChannel = &channel
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// Create deposit
		if err := tx.Create(&deposit).Error; err != nil {
			return err
		}

		// Create transaction dengan status Pending
		message := "Isi Ulang saldo menggunakan " + method
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
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}

		return nil
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
		"expired_at":   deposit.ExpiredAt.Format(time.RFC3339),
		"status":       deposit.Status,
	}
	if method == "QRIS" && resp.ImageQRIS != "" {
		responseData["image_qris"] = resp.ImageQRIS
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
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: resp})
}

// GET /api/users/deposits
func ListDepositsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Optional status filter
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	db := database.DB
	var deposits []models.Deposit
	query := db.Where("user_id = ?", uid)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&deposits).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Map deposits to response format
	type depositDTO struct {
		ID            uint    `json:"id"`
		OrderID       string  `json:"order_id"`
		Amount        float64 `json:"amount"`
		PaymentMethod string  `json:"payment_method"`
		PaymentChannel *string `json:"payment_channel,omitempty"`
		Status        string  `json:"status"`
		ExpiredAt     string  `json:"expired_at"`
		CreatedAt     string  `json:"created_at"`
		UpdatedAt     string  `json:"updated_at"`
	}

	items := make([]depositDTO, 0, len(deposits))
	for _, d := range deposits {
		items = append(items, depositDTO{
			ID:            d.ID,
			OrderID:       d.OrderID,
			Amount:        d.Amount,
			PaymentMethod: d.PaymentMethod,
			PaymentChannel: d.PaymentChannel,
			Status:        d.Status,
			ExpiredAt:     d.ExpiredAt.UTC().Format(time.RFC3339),
			CreatedAt:     d.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:     d.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"deposits": items},
	})
}

// POST /api/payments/linkqu/callback
func LinkQuCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Validasi header client-id dan client-secret
	headerClientID := strings.TrimSpace(r.Header.Get("client-id"))
	headerClientSecret := strings.TrimSpace(r.Header.Get("client-secret"))
	envClientID := strings.TrimSpace(os.Getenv("LINKQU_CLIENT_ID"))
	envClientSecret := strings.TrimSpace(os.Getenv("LINKQU_CLIENT_SECRET"))

	if headerClientID == "" || headerClientSecret == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Missing client-id or client-secret header"})
		return
	}

	if headerClientID != envClientID || headerClientSecret != envClientSecret {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Invalid client-id or client-secret"})
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	partnerReff := strings.TrimSpace(getString(payload, "partner_reff"))
	status := strings.ToUpper(strings.TrimSpace(getString(payload, "status")))

	if partnerReff == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "partner_reff kosong"})
		return
	}

	db := database.DB
	var deposit models.Deposit
	if err := db.Where("order_id = ?", partnerReff).First(&deposit).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Deposit tidak ditemukan"})
		return
	}

	// Cek duplikasi: jika status sudah sama dengan yang dikirim, ignore
	dbStatus := strings.ToUpper(deposit.Status)
	if dbStatus == status {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignore"})
		return
	}

	// Handle status SUCCESS (hanya proses jika deposit masih Pending)
	if status == "SUCCESS" && deposit.Status == "Pending" {
		if err := db.Transaction(func(tx *gorm.DB) error {

			// Update deposit status
			if err := tx.Model(&models.Deposit{}).Where("id = ?", deposit.ID).Updates(map[string]interface{}{
				"status":       "Success",
			}).Error; err != nil {
				return err
			}

			// Update transaction status (transaction sudah dibuat saat deposit dibuat)
			if err := tx.Model(&models.Transaction{}).Where("order_id = ? AND transaction_type = ?", deposit.OrderID, "deposit").Update("status", "Success").Error; err != nil {
				return err
			}

			// Update user balance
			var user models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, balance, spin_ticket").Where("id = ?", deposit.UserID).First(&user).Error; err != nil {
				return err
			}

			newBalance := utils.RoundFloat(user.Balance+deposit.Amount, 2)
			if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("balance", newBalance).Error; err != nil {
				return err
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
		}); err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memproses callback"})
			return
		}

		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK"})
		return
	}

	// Handle status FAILED
	if status == "FAILED" {
		if deposit.Status == "Pending" {
			_ = db.Model(&models.Deposit{}).Where("id = ?", deposit.ID).Update("status", "Failed")
		}
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK"})
		return
	}

	// Handle status PENDING (tidak perlu update, biarkan tetap Pending)
	if status == "PENDING" {
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "OK"})
		return
	}

	// Status tidak dikenal atau tidak valid
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Ignored"})
}

type linkQuConfig struct {
	baseURL      string
	username     string
	pin          string
	clientID     string
	clientSecret string
	callbackURL  string
	location     *time.Location
}

func getLinkQuConfig() (*linkQuConfig, error) {
	baseURL := strings.TrimRight(os.Getenv("LINKQU_BASE_URL"), "/")
	username := os.Getenv("LINKQU_USERNAME")
	pin := os.Getenv("LINKQU_PIN")
	clientID := os.Getenv("LINKQU_CLIENT_ID")
	clientSecret := os.Getenv("LINKQU_CLIENT_SECRET")
	callbackURL := os.Getenv("LINKQU_CALLBACK_PAYMENT")

	if baseURL == "" || username == "" || pin == "" || clientID == "" || clientSecret == "" || callbackURL == "" {
		return nil, errors.New("Konfigurasi LinkQu belum lengkap")
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("Asia/Jakarta", 7*3600)
	}

	return &linkQuConfig{
		baseURL:      baseURL,
		username:     username,
		pin:          pin,
		clientID:     clientID,
		clientSecret: clientSecret,
		callbackURL:  callbackURL,
		location:     loc,
	}, nil
}

func callLinkQu(cfg *linkQuConfig, endpoint string, body map[string]interface{}) (*linkQuResponse, error) {
	fullURL := cfg.baseURL + endpoint

	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("client-id", cfg.clientID)
	req.Header.Set("client-secret", cfg.clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var linkQuResp linkQuResponse
	if err := json.NewDecoder(resp.Body).Decode(&linkQuResp); err != nil {
		return nil, err
	}

	if linkQuResp.ResponseCode != "00" {
		desc := linkQuResp.ResponseDesc
		if desc == "" {
			desc = "Terjadi kesalahan pada sisi pembayaran, Tim kami akan segera menangani masalah tersebut"
		}
		return nil, errors.New(desc)
	}

	return &linkQuResp, nil
}

func parseLinkQuTime(value string, loc *time.Location) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, errors.New("empty")
	}
	return time.ParseInLocation("20060102150405", value, loc)
}

func normalizePhone(number string) string {
	n := strings.TrimSpace(number)
	if n == "" {
		return "0"
	}
	if strings.HasPrefix(n, "0") {
		return n
	}
	if strings.HasPrefix(n, "+62") {
		return "0" + strings.TrimPrefix(n, "+62")
	}
	if strings.HasPrefix(n, "62") {
		return "0" + strings.TrimPrefix(n, "62")
	}
	return "0" + n
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		switch t := v.(type) {
		case string:
			return t
		case fmt.Stringer:
			return t.String()
		case float64:
			return fmt.Sprintf("%.0f", t)
		case json.Number:
			return t.String()
		}
	}
	return ""
}
