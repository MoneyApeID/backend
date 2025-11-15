package users

import (
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strconv"
	"strings"
)

// GET /api/users/team-invited/{level}
// TeamInvitedHandler supports both /api/users/team-invited and /api/users/team-invited/{level}
func TeamInvitedHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var levelStr string
	if len(parts) >= 5 {
		levelStr = parts[4]
	}
	level, levelErr := strconv.Atoi(levelStr)
	hasLevel := (levelErr == nil && level >= 1 && level <= 3)

	// Helper to get users by reff_by
	getUsers := func(parentIDs []uint) ([]models.User, error) {
		var users []models.User
		if len(parentIDs) == 0 {
			return users, nil
		}
		if err := db.Where("reff_by IN ?", parentIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		return users, nil
	}

	// Level 1
	var level1 []models.User
	if err := db.Where("reff_by = ?", uid).Find(&level1).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "DB error"})
		return
	}
	// Level 2
	level1IDs := make([]uint, 0, len(level1))
	for _, u := range level1 {
		level1IDs = append(level1IDs, u.ID)
	}
	level2, _ := getUsers(level1IDs)
	// Level 3
	level2IDs := make([]uint, 0, len(level2))
	for _, u := range level2 {
		level2IDs = append(level2IDs, u.ID)
	}
	level3, _ := getUsers(level2IDs)

	// Helper to count active
	countActive := func(users []models.User) int {
		n := 0
		for _, u := range users {
			if strings.ToLower(u.InvestmentStatus) == "active" {
				n++
			}
		}
		return n
	}

	// Helper to sum total_invest
	sumTotalInvest := func(users []models.User) float64 {
		total := 0.0
		for _, u := range users {
			total += u.TotalInvest
		}
		return total
	}

	// If /api/users/team-invited/{level}
	if hasLevel {
		var users []models.User
		switch level {
		case 1:
			users = level1
		case 2:
			users = level2
		case 3:
			users = level3
		}
		resp := map[string]interface{}{
			strconv.Itoa(level): map[string]interface{}{
				"count":        len(users),
				"active":       countActive(users),
				"total_invest": sumTotalInvest(users),
			},
		}
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
			Success: true,
			Message: "Successfully",
			Data:    resp,
		})
		return
	}

	// If /api/users/team-invited (all levels)
	resp := map[string]interface{}{
		"level": map[string]interface{}{
			"1": map[string]interface{}{"count": len(level1), "active": countActive(level1), "total_invest": sumTotalInvest(level1)},
			"2": map[string]interface{}{"count": len(level2), "active": countActive(level2), "total_invest": sumTotalInvest(level2)},
			"3": map[string]interface{}{"count": len(level3), "active": countActive(level3), "total_invest": sumTotalInvest(level3)},
		},
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    resp["level"],
	})
}

// TeamDataHandler for /api/users/team-data/{level}
func TeamDataHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var levelStr string
	if len(parts) >= 5 {
		levelStr = parts[4]
	}
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 3 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Level must be 1, 2, or 3"})
		return
	}

	// Helper to get users by reff_by
	getUsers := func(parentIDs []uint) ([]models.User, error) {
		var users []models.User
		if len(parentIDs) == 0 {
			return users, nil
		}
		if err := db.Where("reff_by IN ?", parentIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		return users, nil
	}

	// Level 1
	var level1 []models.User
	if err := db.Where("reff_by = ?", uid).Find(&level1).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "DB error"})
		return
	}
	// Level 2
	level1IDs := make([]uint, 0, len(level1))
	for _, u := range level1 {
		level1IDs = append(level1IDs, u.ID)
	}
	level2, _ := getUsers(level1IDs)
	// Level 3
	level2IDs := make([]uint, 0, len(level2))
	for _, u := range level2 {
		level2IDs = append(level2IDs, u.ID)
	}
	level3, _ := getUsers(level2IDs)

	var users []models.User
	switch level {
	case 1:
		users = level1
	case 2:
		users = level2
	case 3:
		users = level3
	}

	// Pagination: page + limit from query
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
	start := (page - 1) * limit
	end := start + limit
	if start > len(users) {
		start = len(users)
	}
	if end > len(users) {
		end = len(users)
	}

	var data []map[string]interface{}
	for _, u := range users[start:end] {
		data = append(data, map[string]interface{}{
			"name":         u.Name,
			"number":       u.Number,
			"active":       strings.ToLower(u.InvestmentStatus) == "active",
			"total_invest": u.TotalInvest,
		})
	}

	resp := map[string]interface{}{
		"level":   level,
		"members": data,
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    resp,
	})
}
