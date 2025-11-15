package users

import (
	"net/http"

	"project/database"
	"project/models"
	"project/utils"
)

// GET /api/users/popup
func GetPopupHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	var setting models.Setting
	if err := db.Select("popup, popup_title, created_at, updated_at").First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	response := map[string]interface{}{
		"image":     setting.Popup,
		"title":     setting.PopupTitle,
		"created":   setting.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated":   setting.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

