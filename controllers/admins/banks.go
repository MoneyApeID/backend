package admins

import (
	"encoding/json"
	"net/http"
	"strconv"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type BankResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name,omitempty"`
	Type      string `json:"type,omitempty"`
	Code      string `json:"code"`
	Status    string `json:"status"`
}

type CreateBankRequest struct {
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Type      string `json:"type"`
	Code      string `json:"code"`
	Status    string `json:"status"`
}

func GetBanks(w http.ResponseWriter, r *http.Request) {
	var banks []models.Bank
	if err := database.DB.Find(&banks).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data bank",
		})
		return
	}

	var response []BankResponse
	for _, bank := range banks {
		response = append(response, BankResponse{
			ID:        bank.ID,
			Name:      bank.Name,
			ShortName: bank.ShortName,
			Type:      bank.Type,
			Code:      bank.Code,
			Status:    bank.Status,
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

func CreateBank(w http.ResponseWriter, r *http.Request) {
	var req CreateBankRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}
	if req.Type == "" {
		req.Type = "bank"
	}
	if req.Status == "" {
		req.Status = "Active"
	}

	// Check for duplicate bank code
	var existingBank models.Bank
	if err := database.DB.Where("code = ?", req.Code).First(&existingBank).Error; err == nil {
		utils.WriteJSON(w, http.StatusConflict, utils.APIResponse{
			Success: false,
			Message: "Bank dengan kode ini sudah digunakan",
		})
		return
	}

	bank := models.Bank{
		Name:      req.Name,
		ShortName: req.ShortName,
		Type:      req.Type,
		Code:      req.Code,
		Status:    req.Status,
	}

	if err := database.DB.Create(&bank).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menambahkan bank",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Bank berhasil ditambahkan",
		Data: BankResponse{
			ID:        bank.ID,
			Name:      bank.Name,
			ShortName: bank.ShortName,
			Type:      bank.Type,
			Code:      bank.Code,
			Status:    bank.Status,
		},
	})
}

func UpdateBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid bank ID",
		})
		return
	}

	var req CreateBankRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}
	if req.Type == "" {
		req.Type = "bank"
	}
	if req.Status == "" {
		req.Status = "Active"
	}

	var bank models.Bank
	if err := database.DB.First(&bank, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Bank tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data bank",
		})
		return
	}

	//check for duplicate bank code
	var existingBank models.Bank
	if err := database.DB.Where("code = ? AND id <> ?", req.Code, bank.ID).First(&existingBank).Error; err == nil {
		utils.WriteJSON(w, http.StatusConflict, utils.APIResponse{
			Success: false,
			Message: "Bank dengan kode ini sudah digunakan",
		})
		return
	}

	bank.Name = req.Name
	bank.ShortName = req.ShortName
	bank.Type = req.Type
	bank.Code = req.Code
	bank.Status = req.Status

	if err := database.DB.Save(&bank).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui bank",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Bank berhasil diperbarui",
		Data: map[string]interface{}{
			"id":         bank.ID,
			"name":       bank.Name,
			"short_name": bank.ShortName,
			"type":       bank.Type,
			"code":       bank.Code,
			"status":     bank.Status,
		},
	})
}
