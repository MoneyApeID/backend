package users

import (
	"errors"
	"net/http"
	"strings"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	// Auth middleware sets user ID in context; use that
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB
	var user models.User
	if err := db.First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "User not found"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Database error"})
		return
	}

	var setting models.Setting
	err := db.Model(&models.Setting{}).
		Select("name, company, logo, min_withdraw, max_withdraw, withdraw_charge, link_cs, link_group, link_app").
		Take(&setting).Error
	healthy := true
	if err != nil {
		healthy = false
	}

	var TotalWithdraw float64
	db.Model(&models.Withdrawal{}).
		Where("user_id = ? AND status = ?", user.ID, "Success").
		Select("COALESCE(SUM(amount),0)").Scan(&TotalWithdraw)

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Succesfully",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"name":           user.Name,
				"number":         user.Number,
				"reff_code":      user.ReffCode,
				"balance":        int64(user.Balance),
				"income":		  int64(user.Income),
				"level":          user.Level,
				"total_invest":   int64(user.TotalInvest),
				"total_invest_vip": int64(user.TotalInvestVIP),
				"total_withdraw": int64(TotalWithdraw),
				"spin_ticket":    user.SpinTicket,
				"active":         strings.ToLower(user.InvestmentStatus) == "active",
			},
			"application": map[string]interface{}{
				"name":            setting.Name,
				"company":         setting.Company,
				"logo":            setting.Logo,
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
