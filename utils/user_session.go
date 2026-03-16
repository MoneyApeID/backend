package utils

import (
	"strings"
	"time"

	"project/database"
	"project/models"
)

// BuildUserSessionPayload returns the same session payload shape used by the user login flow.
func BuildUserSessionPayload(user models.User) (map[string]interface{}, error) {
	accessToken, err := GenerateAccessToken(user.ID, "user")
	if err != nil {
		return nil, err
	}

	refreshJTI, _, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	var totalWithdraw float64
	db := database.DB
	if db != nil {
		db.Model(&models.Withdrawal{}).
			Where("user_id = ? AND status = ?", user.ID, "Success").
			Select("COALESCE(SUM(amount),0)").
			Scan(&totalWithdraw)
	}

	var setting models.Setting
	healthy := true
	if db == nil || db.Model(&models.Setting{}).
		Select("name, company, popup, popup_title, min_withdraw, max_withdraw, withdraw_charge, link_cs, link_group, link_app").
		Take(&setting).Error != nil {
		healthy = false
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"access_expire": time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339),
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
			"total_withdraw":   int64(totalWithdraw),
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
	}, nil
}
