package controllers

import (
	"net/http"

	"project/database"
	"project/models"
	"project/utils"
)

func InfoPublicHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	var setting models.Setting
	if err := db.Model(&models.Setting{}).
		Select("name, company, maintenance, closed_register").
		Take(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil informasi aplikasi",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: map[string]interface{}{
			"name":            setting.Name,
			"company":         setting.Company,
			"maintenance":     setting.Maintenance,
			"closed_register": setting.ClosedRegister,
			"min_deposit":     setting.MinDeposit,
		},
	})
}
