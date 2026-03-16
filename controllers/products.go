package controllers

import (
	"net/http"

	"project/database"
	"project/models"
	"project/utils"
)

func ProductListHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Get active categories ordered by ID
	var categories []models.Category
	if err := db.Where("status = ?", "Active").Order("id ASC").Find(&categories).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}

	// Get active products with category info
	var products []models.Product
	if err := db.Preload("Category").Where("status = ?", "Active").Order("category_id ASC, id ASC").Find(&products).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
		return
	}

	// Group products by category name
	categoryMap := make(map[string][]models.Product)
	for _, p := range products {
		if p.Category != nil {
			categoryMap[p.Category.Name] = append(categoryMap[p.Category.Name], p)
		}
	}

	// Prepare response with all categories (even empty ones)
	resp := make(map[string]interface{})
	for _, cat := range categories {
		if prods, ok := categoryMap[cat.Name]; ok {
			resp[cat.Name] = prods
		} else {
			resp[cat.Name] = []models.Product{}
		}
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    resp,
	})
}
