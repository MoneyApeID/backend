package users

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

func userIDFromAuthHeader(r *http.Request) (uint, error) {
	if uid, ok := utils.GetUserID(r); ok && uid != 0 {
		return uid, nil
	}
	return 0, fmt.Errorf("unauthorized")
}

// CalculateChancePercentage computes ChancePercent for response
// calculateChancePercentage is no longer needed

// GET /api/spin-prize-list
func SpinPrizeListHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Use a struct to control the response format
	type PrizeResponse struct {
		ID     uint    `json:"id"`
		Amount float64 `json:"amount"`
		Code   string  `json:"code"`
		Chance float64 `json:"chance"`
		Status string  `json:"status"`
	}

	var prizes []models.SpinPrize
	if err := db.Select("id, amount, code, chance_weight, status").Where("status = ?", "Active").Order("amount ASC").Find(&prizes).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Calculate total weight
	totalWeight := 0
	for _, p := range prizes {
		totalWeight += p.ChanceWeight
	}

	// Transform to response format with calculated chances
	var response []PrizeResponse
	for _, p := range prizes {
		chance := float64(0)
		if totalWeight > 0 {
			chance = float64(p.ChanceWeight) / float64(totalWeight) * 100
		}
		response = append(response, PrizeResponse{
			ID:     p.ID,
			Amount: p.Amount,
			Code:   p.Code,
			Chance: utils.RoundFloat(chance, 2), // Calculate percentage and round to 2 decimal places
			Status: p.Status,
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Berhasil mengambil daftar hadiah spin",
		Data:    response,
	})
}

//type spinClaimRequest struct {
//	Code string `json:"code"`
//}

// POST /api/users/spin
// func UserSpinClaimHandler(w http.ResponseWriter, r *http.Request) {
// 	userID, err := userIDFromAuthHeader(r)
// 	if err != nil {
// 		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
// 		return
// 	}

// 	var req spinClaimRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
// 		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Hadiah tidak valid atau sudah tidak tersedia"})
// 		return
// 	}

// 	db := database.DB

// 	// Get user and check spin_ticket
// 	var user models.User
// 	if err := db.Select("id, balance, spin_ticket").Where("id = ?", userID).First(&user).Error; err != nil {
// 		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
// 		return
// 	}
// 	if user.SpinTicket == nil || *user.SpinTicket == 0 {
// 		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Tiket spin Anda habis, silakan dapatkan tiket terlebih dahulu"})
// 		return
// 	}

// 	// Validate code existence and Active
// 	var claimedPrize models.SpinPrize
// 	if err := db.Where("code = ? AND status = 'Active'", req.Code).First(&claimedPrize).Error; err != nil {
// 		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Hadiah tidak valid atau sudah tidak tersedia"})
// 		return
// 	}

// 	finalPrize := claimedPrize
// 	previousBalance := user.Balance
// 	var currentBalance float64

// 	err = db.Transaction(func(tx *gorm.DB) error {
// 		// Decrement spin_ticket
// 		if err := tx.Model(&models.User{}).Where("id = ? AND spin_ticket > 0", userID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket - 1")).Error; err != nil {
// 			return err
// 		}

// 		// Create transaction
// 		msg := "Hadiah Spin Wheel"
// 		orderID := fmt.Sprintf("INV-%d%d", time.Now().Unix(), userID)

// 		trx := models.Transaction{
// 			UserID:          userID,
// 			Amount:          finalPrize.Amount,
// 			Charge:          0,
// 			OrderID:         orderID,
// 			TransactionFlow: "debit",
// 			TransactionType: "bonus",
// 			Message:         &msg,
// 			Status:          "Success",
// 		}
// 		if err := tx.Create(&trx).Error; err != nil {
// 			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
// 			return err
// 		}

// 		// Increase user's balance
// 		if err := tx.Model(&models.User{}).
// 			Where("id = ?", userID).
// 			UpdateColumn("balance", gorm.Expr("balance + ?", finalPrize.Amount)).Error; err != nil {
// 			return err
// 		}

// 		// Read updated balance
// 		if err := tx.Select("balance").Where("id = ?", userID).First(&user).Error; err != nil {
// 			return err
// 		}
// 		currentBalance = user.Balance
// 		return nil
// 	})

// 	if err != nil {
// 		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan sistem, silakan coba lagi"})
// 		return
// 	}

// 	// Success response
// 	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
// 		Success: true,
// 		Message: fmt.Sprintf("Selamat! Anda memenangkan Rp%.0f", finalPrize.Amount),
// 		Data: map[string]interface{}{
// 			"spin_result": map[string]interface{}{
// 				"amount": finalPrize.Amount,
// 				"code":   finalPrize.Code,
// 			},
// 			"balance_info": map[string]interface{}{
// 				"previous_balance": int64(previousBalance),
// 				"prize_amount":     int64(finalPrize.Amount),
// 				"current_balance":  int64(currentBalance),
// 			},
// 		},
// 	})
// }

// GET /api/users/spin-v2
func UserSpinHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromAuthHeader(r)
	if err != nil || userID == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB

	// Get user and check spin_ticket
	var user models.User
	if err := db.Select("id, income, spin_ticket").Where("id = ?", userID).First(&user).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan data, silakan coba lagi"})
		log.Println(err)
		return
	}
	if user.SpinTicket == nil || *user.SpinTicket == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Tiket spin Anda habis, silakan dapatkan tiket terlebih dahulu"})
		return
	}

	// Load active prizes (ensure chance field populated from chance_weight)

	var prizes []models.SpinPrize
	if err := db.Select("id, amount, code, chance_weight, status").Where("status = 'Active'").Order("amount ASC").Find(&prizes).Error; err != nil || len(prizes) == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Hadiah tidak valid atau sudah tidak tersedia"})
		return
	}

	// Sum weights
	totalWeight := 0
	for _, p := range prizes {
		if p.ChanceWeight > 0 {
			totalWeight += p.ChanceWeight
		}
	}
	if totalWeight <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Hadiah tidak valid atau sudah tidak tersedia"})
		return
	}

	// Pick prize by weighted chance
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	pick := rnd.Intn(totalWeight) + 1
	acc := 0
	var finalPrize models.SpinPrize
	for _, p := range prizes {
		acc += p.ChanceWeight
		if pick <= acc {
			finalPrize = p
			break
		}
	}

	previousBalance := user.Income
	var currentBalance float64

	err = db.Transaction(func(tx *gorm.DB) error {
		// Decrement spin_ticket
		if err := tx.Model(&models.User{}).Where("id = ? AND spin_ticket > 0", userID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket - 1")).Error; err != nil {
			return err
		}

		// Create transaction
		msg := "Hadiah Spin Wheel"
		orderID := utils.GenerateOrderID(userID)

		trx := models.Transaction{
			UserID:          userID,
			Amount:          finalPrize.Amount,
			Charge:          0,
			OrderID:         orderID,
			TransactionFlow: "debit",
			TransactionType: "bonus",
			Message:         &msg,
			Status:          "Success",
		}
		if err := tx.Create(&trx).Error; err != nil {
			log.Println("Error creating transaction:", err)
			return err
		}

		// Record user spin history
		userSpin := models.UserSpin{
			UserID:  userID,
			PrizeID: finalPrize.ID,
			Amount:  finalPrize.Amount,
			Code:    finalPrize.Code,
			WonAt:   time.Now(),
		}
		if err := tx.Create(&userSpin).Error; err != nil {
			return err
		}

		// Increase user's balance
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			UpdateColumn("income", gorm.Expr("income + ?", finalPrize.Amount)).Error; err != nil {
			return err
		}

		// Read updated balance
		if err := tx.Select("income").Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		currentBalance = user.Income
		return nil
	})

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan server, silakan coba lagi"})
		log.Println(err)
		return
	}

	// Success response (same shape as claim handler)
	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Selamat! Anda memenangkan Rp%.0f", finalPrize.Amount),
		Data: map[string]interface{}{
			"spin_result": map[string]interface{}{
				"amount": finalPrize.Amount,
				"code":   finalPrize.Code,
			},
			"balance_info": map[string]interface{}{
				"previous_balance": int64(previousBalance),
				"prize_amount":     int64(finalPrize.Amount),
				"current_balance":  int64(currentBalance),
			},
		},
	})
}
