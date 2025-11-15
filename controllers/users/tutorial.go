package users

import (
	"net/http"

	"project/database"
	"project/models"
	"project/utils"
)

// GET /api/users/tutorials
func ListTutorialsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	var tutorials []models.Tutorial
	if err := db.Where("status = ?", "Active").Order("id DESC").Find(&tutorials).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil data tutorial"})
		return
	}

	// Map to response format without status field
	type tutorialDTO struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		Image     string `json:"image"`
		Link      string `json:"link"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	items := make([]tutorialDTO, 0, len(tutorials))
	for _, t := range tutorials {
		items = append(items, tutorialDTO{
			ID:        t.ID,
			Title:     t.Title,
			Image:     t.Image,
			Link:      t.Link,
			CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"tutorials": items},
	})
}

