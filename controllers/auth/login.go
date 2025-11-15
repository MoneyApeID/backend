package auth

import (
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

type LoginRequest struct {
	Number   string `json:"number" validate:"required,phone8"`
	Password string `json:"password" validate:"required,pwdmin"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := middleware.ValidateJSON(w, r, &req); err != nil {
		return
	}

	// Check maintenance mode
	var appSetting models.Setting
	if err := database.DB.Model(&models.Setting{}).Select("maintenance, name").Take(&appSetting).Error; err == nil && appSetting.Maintenance {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Aplikasi sedang dalam pemeliharaan. Silakan coba lagi nanti.",
			Data:    map[string]interface{}{"maintenance": true, "application": appSetting.Name},
		})
		return
	}

	db := database.DB

	var user models.User
	if err := db.Where("number = ?", req.Number).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Nomor telpon atau password salah"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	// check account lockout
	if locked, retry := middleware.IsAccountLocked(user.ID); locked {
		utils.WriteJSON(w, http.StatusTooManyRequests, utils.APIResponse{Success: false, Message: "Terlalu banyak percobaan login. Coba lagi nanti.", Data: map[string]interface{}{"retry_after_seconds": int(retry.Seconds())}})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// record failed login attempt for lockout tracking
		middleware.RecordFailedLogin(user.ID)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Nomor telpon atau password salah"})
		return
	}

	// on successful login reset failed login counter
	middleware.ResetFailedLogin(user.ID)

	// generate access token (short-lived) and refresh token (stored in DB)
	accessToken, err := utils.GenerateAccessToken(user.ID, "user")
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal login"})
		return
	}
	refreshJTI, _, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menyimpan refresh token"})
		return
	}
	signed := accessToken
	exp := time.Now().Add(15 * time.Minute)

	var TotalWithdraw float64
	db.Model(&models.Withdrawal{}).
		Where("user_id = ? AND status = ?", user.ID, "Success").
		Select("COALESCE(SUM(amount),0)").Scan(&TotalWithdraw)

	// Ambil data settings
	var setting models.Setting
	err = db.Model(&models.Setting{}).
		Select("name, company, popup, popup_title, min_withdraw, max_withdraw, withdraw_charge, link_cs, link_group, link_app").
		Take(&setting).Error
	healthy := true
	if err != nil {
		healthy = false
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Login berhasil! Mengalihkan ke dashboard...",
		Data: map[string]interface{}{
			"access_token":  signed,
			"access_expire": exp.UTC().Format(time.RFC3339),
			"refresh_token": refreshJTI,
			"user": map[string]interface{}{
				"name":             user.Name,
				"number":           user.Number,
				"reff_code":        user.ReffCode,
				"balance":          int64(user.Balance),
				"income":           int64(user.Income),
				"level":            user.Level,
				"total_invest":     int64(user.TotalInvest),
				"total_invest_vip": int64(user.TotalInvestVIP),
				"total_withdraw":   int64(TotalWithdraw),
				"spin_ticket":      user.SpinTicket,
				"active":           strings.ToLower(user.InvestmentStatus) == "active",
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
