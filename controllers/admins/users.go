package admins

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"project/database"
	"project/models"
	"project/utils"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserResponse struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Number           string  `json:"number"`
	ReffCode         string  `json:"reff_code"`
	ReffBy           uint    `json:"reff_by"`
	Balance          float64 `json:"balance"`
	Income           float64 `json:"income"`
	Level            int     `json:"level,omitempty"`
	TotalInvest      float64 `json:"total_invest"`
	SpinTicket       int     `json:"spin_ticket"`
	Status           string  `json:"status"`
	InvestmentStatus string  `json:"investment_status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Start the query
	db := database.DB
	query := db.Model(&models.User{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		search = "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR number LIKE ? OR reff_code LIKE ?", search, search, search)
	}

	// Get users with pagination
	var users []models.User
	query.Offset(offset).Limit(limit).Find(&users)

	// Transform to response format
	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Number:   user.Number,
			ReffCode: user.ReffCode,
			ReffBy: func() uint {
				if user.ReffBy != nil {
					return *user.ReffBy
				}
				return 0
			}(),
			Balance:     user.Balance,
			Income:      user.Income,
			TotalInvest: user.TotalInvest,
			SpinTicket: func() int {
				if user.SpinTicket != nil {
					return int(*user.SpinTicket)
				} else {
					return 0
				}
			}(),
			Status:           user.Status,
			InvestmentStatus: user.InvestmentStatus,
			CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "User tidak valid",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "User tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	response := UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Number:   user.Number,
		ReffCode: user.ReffCode,
		ReffBy: func() uint {
			if user.ReffBy != nil {
				return *user.ReffBy
			} else {
				return 0
			}
		}(),
		Balance: user.Balance,
		Income: user.Income,
		Level: func() int {
			if user.Level != nil {
				return int(*user.Level)
			} else {
				return 0
			}
		}(),
		TotalInvest: user.TotalInvest,
		SpinTicket: func() int {
			if user.SpinTicket != nil {
				return int(*user.SpinTicket)
			} else {
				return 0
			}
		}(),
		Status:           user.Status,
		InvestmentStatus: user.InvestmentStatus,
		CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

type UpdateUserRequest struct {
	Name             string `json:"name"`
	Number           string `json:"number"`
	Status           string `json:"status"`
	InvestmentStatus string `json:"investment_status"`
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID pengguna tidak valid",
		})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Format data tidak valid",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Pengguna tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data pengguna",
		})
		return
	}

	// Check if phone number is already used by another user
	if user.Number != req.Number { // Only check if number is being changed
		var existingUser models.User
		if err := database.DB.Where("number = ? AND id != ?", req.Number, id).First(&existingUser).Error; err == nil {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Nomor telepon sudah digunakan pengguna lain",
			})
			return
		} else if err != gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memeriksa nomor telepon",
			})
			return
		}
	}

	// Update fields
	user.Name = req.Name
	user.Number = req.Number
	user.Status = req.Status
	user.InvestmentStatus = req.InvestmentStatus

	if err := database.DB.Save(&user).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui data pengguna",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Berhasil memperbarui data pengguna",
		Data: UserResponse{
			ID:               user.ID,
			Name:             user.Name,
			Number:           user.Number,
			Status:           user.Status,
			InvestmentStatus: user.InvestmentStatus,
			Level: func() int {
				if user.Level != nil {
					return int(*user.Level)
				} else {
					return 0
				}
			}(),
			SpinTicket: func() int {
				if user.SpinTicket != nil {
					return int(*user.SpinTicket)
				} else {
					return 0
				}
			}(),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

type UpdateBalanceRequest struct {
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`         // "add" or "less"
	BalanceType string  `json:"balance_type"` // "balance" or "income"
}

func UpdateUserBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID pengguna tidak valid",
		})
		return
	}

	var req UpdateBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Format data tidak valid",
		})
		return
	}

	if req.Amount <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Jumlah harus lebih besar dari 0",
		})
		return
	}

	// Validate balance_type, default to "balance" if not provided
	if req.BalanceType == "" {
		req.BalanceType = "balance"
	}
	if req.BalanceType != "balance" && req.BalanceType != "income" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "balance_type harus 'balance' atau 'income'",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Pengguna tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data pengguna",
		})
		return
	}

	db := database.DB

	switch req.Type {
	case "add":
		// Update balance or income based on balance_type
		if req.BalanceType == "balance" {
			user.Balance += req.Amount
		} else {
			user.Income += req.Amount
		}

		// Jalankan dalam transaksi: update saldo/income + buat log transaksi
		err = db.Transaction(func(tx *gorm.DB) error {
			// Simpan perubahan saldo/income
			if err := tx.Save(&user).Error; err != nil {
				return err
			}

			// Buat record transaksi
			msg := fmt.Sprintf("Bonus %s dari admin", req.BalanceType)
			trx := models.Transaction{
				UserID:          user.ID,
				Amount:          req.Amount,
				Charge:          0,
				OrderID:         utils.GenerateOrderID(user.ID),
				TransactionFlow: "debit",
				TransactionType: "bonus",
				Message:         &msg,
				Status:          "Success",
			}

			if err := tx.Create(&trx).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui saldo/income dan mencatat transaksi",
			})
			return
		}

	case "less":
		// Check if balance/income is sufficient
		if req.BalanceType == "balance" {
			if user.Balance < req.Amount {
				utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
					Success: false,
					Message: "Saldo tidak mencukupi",
				})
				return
			}
			user.Balance -= req.Amount
		} else {
			if user.Income < req.Amount {
				utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
					Success: false,
					Message: "Income tidak mencukupi",
				})
				return
			}
			user.Income -= req.Amount
		}

		// Jalankan dalam transaksi: hanya update saldo/income
		err = db.Transaction(func(tx *gorm.DB) error {
			return tx.Save(&user).Error
		})

		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui saldo/income pengguna",
			})
			return
		}

	default:
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Tipe transaksi tidak valid",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Berhasil memperbarui saldo pengguna",
	})
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

func UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "ID pengguna tidak valid",
		})
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Format data tidak valid",
		})
		return
	}

	if len(req.Password) < 6 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Password minimal 6 karakter",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
				Success: false,
				Message: "Pengguna tidak ditemukan",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengambil data pengguna",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui password",
		})
		return
	}

	user.Password = string(hashedPassword)

	if err := database.DB.Save(&user).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal memperbarui password",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Berhasil memperbarui password pengguna",
	})
}
