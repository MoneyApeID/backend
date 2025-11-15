package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"project/database"
	"project/middleware"
	"project/models"
	"project/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Name                 string `json:"name" validate:"required,nameok"`
	Number               string `json:"number" validate:"required,phone8"`
	Password             string `json:"password" validate:"required,pwdmin"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
	ReferralCode         string `json:"referral_code"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := middleware.ValidateJSON(w, r, &req); err != nil {
		return
	}

	// Check if registration is closed
	var appSetting models.Setting
	if err := database.DB.Model(&models.Setting{}).Select("closed_register, name").Take(&appSetting).Error; err == nil && appSetting.ClosedRegister {
		utils.WriteJSON(w, http.StatusForbidden, utils.APIResponse{
			Success: false,
			Message: "Pendaftaran sedang ditutup. Silakan coba lagi nanti.",
			Data:    map[string]interface{}{"closed_register": true, "application": appSetting.Name},
		})
		return
	}

	if err := database.DB.Model(&models.Setting{}).Select("maintenance, name").Take(&appSetting).Error; err == nil && appSetting.Maintenance {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Aplikasi sedang dalam pemeliharaan. Silakan coba lagi nanti.",
			Data:    map[string]interface{}{"maintenance": true, "application": appSetting.Name},
		})
		return
	}

	// Trim inputs
	req.Name = strings.TrimSpace(req.Name)
	req.Number = strings.TrimSpace(req.Number)
	req.ReferralCode = strings.TrimSpace(req.ReferralCode)

	// Validations
	if req.Name == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nama lengkap tidak boleh kosong"})
		return
	}
	if req.Number == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Nomor telepon tidak boleh kosong"})
		return
	}
	if len(req.Password) < 6 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Password minimal 6 karakter"})
		return
	}
	if req.Password != req.PasswordConfirmation {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Password tidak cocok"})
		return
	}
	if req.ReferralCode == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kode referral tidak boleh kosong"})
		return
	}

	db := database.DB

	// Ensure unique number
	var existing models.User
	if err := db.Where("number = ?", req.Number).First(&existing).Error; err == nil {
		utils.WriteJSON(w, http.StatusConflict, utils.APIResponse{Success: false, Message: "Nomor telepon sudah terdaftar"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	// Referral handling
	var reffBy *uint
	if req.ReferralCode != "" {
		var refOwner models.User
		if err := db.Where("reff_code = ?", req.ReferralCode).First(&refOwner).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kode referral tidak valid"})
				return
			}
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
			return
		}
		reffBy = &refOwner.ID
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	// Generate unique referral code
	code, err := generateUniqueReffCode(db, 8)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	level := uint(1)
	newUser := models.User{
		Name:          req.Name,
		Number:        req.Number,
		Password:      string(hashed),
		ReffCode:      code,
		ReffBy:        reffBy,
		Balance:       2000,
		Level:         &level,
		Status:        "Active",
	}

	if err := db.Create(&newUser).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Registrasi gagal, silakan coba lagi", Data: err.Error()})
		return
	}

	// Assign user baru ke binary tree (kiri/kanan) jika ada upline
	if reffBy != nil {
		// Assign binary node (kiri atau kanan)
		if err := utils.AssignBinaryNode(*reffBy, newUser.ID); err != nil {
			// Log error tapi jangan gagalkan registrasi
			// Binary assignment bisa dilakukan nanti jika diperlukan
		}
	}

	newTransaction := models.Transaction{
		UserID:          newUser.ID,
		Amount:          2000,
		Charge:          0,
		OrderID:         utils.GenerateOrderID(newUser.ID),
		TransactionFlow: "debit",
		TransactionType: "bonus",
		Message:         ptrString("Bonus pendaftaran"),
		Status:          "Success",
	}

	if err := db.Create(&newTransaction).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	// Generate access and refresh tokens
	accessToken, err := utils.GenerateAccessToken(newUser.ID, "user")
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat token"})
		return
	}
	refreshJTI, _, err := utils.GenerateRefreshToken(newUser.ID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menyimpan refresh token"})
		return
	}
	signed := accessToken
	exp := time.Now().Add(15 * time.Minute)

	var setting models.Setting
	err = db.Model(&models.Setting{}).
		Select("name, company, popup, popup_title, min_withdraw, max_withdraw, withdraw_charge, link_cs, link_group, link_app").
		Take(&setting).Error
	healthy := true
	if err != nil {
		healthy = false
	}

	var TotalWithdraw float64
	db.Model(&models.Withdrawal{}).
		Where("user_id = ? AND status = ?", newUser.ID, "Success").
		Select("COALESCE(SUM(amount),0)").Scan(&TotalWithdraw)

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Registrasi berhasil, Selamat datang!",
		Data: map[string]interface{}{
			"access_token":  signed,
			"access_expire": exp.UTC().Format(time.RFC3339),
			"refresh_token": refreshJTI,
			"user": map[string]interface{}{
				"name":             newUser.Name,
				"number":           newUser.Number,
				"reff_code":        newUser.ReffCode,
				"balance":          int64(newUser.Balance),
				"income":           int64(newUser.Income),
				"level":            newUser.Level,
				"total_invest":     int64(newUser.TotalInvest),
				"total_invest_vip": int64(newUser.TotalInvestVIP),
				"total_withdraw":   int64(TotalWithdraw),
				"spin_ticket":      newUser.SpinTicket,
				"active":           strings.ToLower(newUser.InvestmentStatus) == "active",
			},
			"application": map[string]interface{}{
				"name":            setting.Name,
				"company":         setting.Company,
				"popup":           setting.Popup,
				"popup_title":     setting.PopupTitle,
				"min_withdraw":    int64(setting.MinWithdraw),
				"max_withdraw":    int64(setting.MaxWithdraw),
				"withdraw_charge": int64(setting.WithdrawCharge),
				"link_cs":         setting.LinkCS,
				"link_group":      setting.LinkGroup,
				"link_app":        setting.LinkApp,
				"healthy":         healthy,
			},
		},
	})
}

func generateUniqueReffCode(db *gorm.DB, length int) (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttempts := 100

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := randomString(alphabet, length)
		if err != nil {
			return "", err
		}
		var count int64
		if err := db.Model(&models.User{}).Where("reff_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique referral code after %d attempts", maxAttempts)
}

func randomString(alphabet string, length int) (string, error) {
	buf := make([]byte, length)
	out := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := 0; i < length; i++ {
		out[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(out), nil
}

// ptrString returns a pointer to the given string.
func ptrString(s string) *string {
	return &s
}
