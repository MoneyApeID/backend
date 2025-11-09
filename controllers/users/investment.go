package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateInvestmentRequest struct {
	ProductID uint `json:"product_id"`
}

var errInsufficientBalance = errors.New("saldo tidak mencukupi")

// GET /api/users/investment/active
func GetActiveInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	db := database.DB

	// Get active categories (prioritize category ID 1)
	var categories []models.Category
	if err := db.Where("status = ?", "Active").Order("CASE WHEN id = 1 THEN 0 ELSE id END ASC").Find(&categories).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil kategori"})
		return
	}

	var investments []models.Investment
	if err := db.Preload("Category").Where("user_id = ? AND status IN ?", uid, []string{"Running", "Completed", "Suspended"}).Order("CASE WHEN category_id = 1 THEN 0 ELSE category_id END ASC, product_id ASC, id DESC").Find(&investments).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil investasi"})
		return
	}

	// Group investments by category name
	categoryMap := make(map[string][]map[string]interface{})
	for _, inv := range investments {
		var product models.Product
		if err := db.Preload("Category").Where("id = ?", inv.ProductID).First(&product).Error; err != nil {
			continue
		}

		catName := ""
		if inv.Category != nil {
			catName = inv.Category.Name
		}

		// Prepare product category info
		var productCategory map[string]interface{}
		if product.Category != nil {
			productCategory = map[string]interface{}{
				"id":          product.Category.ID,
				"name":        product.Category.Name,
				"status":      product.Category.Status,
				"profit_type": product.Category.ProfitType,
			}
		}

		m := map[string]interface{}{
			"id":               inv.ID,
			"user_id":          inv.UserID,
			"product_id":       inv.ProductID,
			"product_name":     product.Name,
			"product_category": productCategory,
			"category_id":      inv.CategoryID,
			"category_name":    catName,
			"amount":           int64(inv.Amount),
			"duration":         inv.Duration,
			"daily_profit":     int64(inv.DailyProfit),
			"total_paid":       inv.TotalPaid,
			"total_returned":   int64(inv.TotalReturned),
			"last_return_at":   inv.LastReturnAt,
			"next_return_at":   inv.NextReturnAt,
			"order_id":         inv.OrderID,
			"status":           inv.Status,
		}
		categoryMap[catName] = append(categoryMap[catName], m)
	}

	// Ensure all categories exist in response
	resp := make(map[string]interface{})
	for _, cat := range categories {
		if invs, ok := categoryMap[cat.Name]; ok {
			resp[cat.Name] = invs
		} else {
			resp[cat.Name] = []map[string]interface{}{}
		}
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: resp})
}

// POST /api/users/investments - FIXED VERSION
func CreateInvestmentHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateInvestmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Not valid JSON"})
		return
	}

	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB
	var product models.Product
	if err := db.Preload("Category").Where("id = ? AND status = 'Active'", req.ProductID).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Produk tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
		return
	}

	if product.Category == nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kategori produk tidak valid"})
		return
	}

	var user models.User
	if err := db.Where("id = ?", uid).First(&user).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
		return
	}

	userLevel := uint(0)
	if user.Level != nil {
		userLevel = *user.Level
	}

	if userLevel < uint(product.RequiredVIP) {
		msg := fmt.Sprintf("Produk %s memerlukan VIP level %d. Level VIP Anda saat ini: %d", product.Name, product.RequiredVIP, userLevel)
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: msg})
		return
	}

	if product.PurchaseLimit > 0 {
		var purchaseCount int64
		if err := db.Model(&models.Investment{}).
			Where("user_id = ? AND product_id = ? AND status IN ?", uid, product.ID, []string{"Running", "Completed", "Suspended"}).
			Count(&purchaseCount).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan, coba lagi"})
			return
		}
		if purchaseCount >= int64(product.PurchaseLimit) {
			msg := fmt.Sprintf("Anda telah mencapai batas pembelian untuk produk %s (maksimal %dx)", product.Name, product.PurchaseLimit)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: msg})
			return
		}
	}

	orderID := utils.GenerateOrderID(uid)
	now := time.Now()
	nextReturn := now.Add(24 * time.Hour)

	inv := models.Investment{
		UserID:        uid,
		ProductID:     product.ID,
		CategoryID:    product.CategoryID,
		Amount:        product.Amount,
		DailyProfit:   product.DailyProfit,
		Duration:      product.Duration,
		TotalPaid:     0,
		TotalReturned: 0,
		OrderID:       orderID,
		Status:        "Running",
		NextReturnAt:  &nextReturn,
	}

	var finalBalance float64

	if err := db.Transaction(func(tx *gorm.DB) error {
		var userForUpdate models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", uid).First(&userForUpdate).Error; err != nil {
			return err
		}

		if userForUpdate.Balance < product.Amount {
			return errInsufficientBalance
		}

		if err := tx.Create(&inv).Error; err != nil {
			return err
		}

		isLockedCategory := false
		if product.Category != nil && strings.EqualFold(product.Category.ProfitType, "locked") {
			isLockedCategory = true
		}

        newBalance := round3(userForUpdate.Balance - product.Amount)
		finalBalance = newBalance
		newTotalInvest := round3(userForUpdate.TotalInvest + product.Amount)
		newTotalInvestVIP := userForUpdate.TotalInvestVIP
		updateFields := map[string]interface{}{
			"balance":            newBalance,
			"total_invest":       newTotalInvest,
			"investment_status":  "Active",
			"updated_at":         time.Now(),
		}

		if isLockedCategory {
			newTotalInvestVIP = round3(userForUpdate.TotalInvestVIP + product.Amount)
			updateFields["total_invest_vip"] = newTotalInvestVIP
		}

		if err := tx.Model(&models.User{}).Where("id = ?", uid).Updates(updateFields).Error; err != nil {
			return err
		}

		// Update VIP level if needed for locked category
		if isLockedCategory {
			newLevel := calculateVIPLevel(newTotalInvestVIP)
			if userForUpdate.Level == nil || *userForUpdate.Level != newLevel {
				if err := tx.Model(&models.User{}).Where("id = ?", uid).Update("level", newLevel).Error; err != nil {
					return err
				}
			}
		}

		msg := fmt.Sprintf("Investasi %s", product.Name)
		trx := models.Transaction{
			UserID:          uid,
			Amount:          product.Amount,
			Charge:          0,
			OrderID:         orderID,
			TransactionFlow: "credit",
			TransactionType: "investment",
			Message:         &msg,
			Status:          "Success",
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}

		// Handle referral bonus (level 1)
		if userForUpdate.ReffBy != nil {
			var refUser models.User
			if err := tx.Select("id, spin_ticket, balance").Where("id = ?", *userForUpdate.ReffBy).First(&refUser).Error; err == nil {
				if product.Amount >= 100000 {
					if refUser.SpinTicket == nil {
						one := uint(1)
						if err := tx.Model(&models.User{}).Where("id = ?", refUser.ID).Update("spin_ticket", one).Error; err != nil {
							return err
						}
					} else {
						if err := tx.Model(&models.User{}).Where("id = ?", refUser.ID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket + ?", 1)).Error; err != nil {
							return err
						}
					}
				}

				bonus := round3(product.Amount * 0.30)
				if err := tx.Model(&models.User{}).Where("id = ?", refUser.ID).UpdateColumn("balance", gorm.Expr("balance + ?", bonus)).Error; err != nil {
					return err
				}
				msgBonus := "Bonus rekomendasi investor"
				refTrx := models.Transaction{
					UserID:          refUser.ID,
					Amount:          bonus,
					Charge:          0,
					OrderID:         utils.GenerateOrderID(refUser.ID),
					TransactionFlow: "debit",
					TransactionType: "team",
					Message:         &msgBonus,
					Status:          "Success",
				}
				if err := tx.Create(&refTrx).Error; err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		if errors.Is(err, errInsufficientBalance) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Saldo tidak mencukupi"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat investasi"})
		return
	}

	resp := map[string]interface{}{
		"order_id":     inv.OrderID,
		"amount":       inv.Amount,
		"product":      product.Name,
		"category":     product.Category.Name,
		"category_id":  product.CategoryID,
		"duration":     product.Duration,
		"daily_profit": product.DailyProfit,
		"status":       inv.Status,
		"next_return": func() interface{} {
			if inv.NextReturnAt == nil {
				return nil
			}
			return inv.NextReturnAt.Format(time.RFC3339)
		}(),
		"balance": round3(finalBalance),
	}
	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{Success: true, Message: "Investasi berhasil menggunakan saldo", Data: resp})
}

// GET /api/users/investments
func ListInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
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
	offset := (page - 1) * limit

	db := database.DB
	var rows []models.Investment
	if err := db.Where("user_id = ?", uid).Order("id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: map[string]interface{}{"investments": rows}})
}

// GET /api/users/investments/{id}
func GetInvestmentHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var idStr string
	if len(parts) >= 4 {
		idStr = parts[3]
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID tidak valid"})
		return
	}
	db := database.DB
	var row models.Investment
	if err := db.Where("id = ? AND user_id = ?", uint(id64), uid).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Data tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Successfully", Data: row})
}

// POST /api/cron/daily-returns
func CronDailyReturnsHandler(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-CRON-KEY")
	if key == "" || key != os.Getenv("CRON_KEY") {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	db := database.DB
	now := time.Now()
	var due []models.Investment
	if err := db.Where("status = 'Running' AND next_return_at IS NOT NULL AND next_return_at <= ? AND total_paid < duration", now).Find(&due).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}
	processed := 0
	for i := range due {
		inv := due[i]
		_ = db.Transaction(func(tx *gorm.DB) error {
			var user models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, inv.UserID).Error; err != nil {
				return err
			}

			// Get category to check profit type
			var category models.Category
			if err := tx.Where("id = ?", inv.CategoryID).First(&category).Error; err != nil {
				return err
			}

			amount := inv.DailyProfit
			paid := inv.TotalPaid + 1
			returned := round3(inv.TotalReturned + amount)

			var product models.Product
			if err := tx.Where("id = ?", inv.ProductID).First(&product).Error; err != nil {
				return err
			}

			// For locked (Monitor) category: Don't pay to balance until completion, just accumulate
			// For unlocked (Insight/AutoPilot): Pay to balance immediately
			if category.ProfitType == "unlocked" {
				newBalance := round3(user.Balance + amount)
				if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
					return err
				}

				orderID := utils.GenerateOrderID(inv.UserID)
				msg := fmt.Sprintf("Profit investasi produk %s", product.Name)
				trx := models.Transaction{
					UserID:          inv.UserID,
					Amount:          amount,
					Charge:          0,
					OrderID:         orderID,
					TransactionFlow: "debit",
					TransactionType: "return",
					Message:         &msg,
					Status:          "Success",
				}
				if err := tx.Create(&trx).Error; err != nil {
					return err
				}
			}

			// For locked (Monitor): If completing, pay total accumulated profit
			if category.ProfitType == "locked" && paid >= inv.Duration {
				totalProfit := round3(inv.DailyProfit * float64(inv.Duration))
				newBalance := round3(user.Balance + totalProfit)
				if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
					return err
				}

				orderID := utils.GenerateOrderID(inv.UserID)
				msg := fmt.Sprintf("Total profit investasi produk %s selesai", product.Name)
				trx := models.Transaction{
					UserID:          inv.UserID,
					Amount:          totalProfit,
					Charge:          0,
					OrderID:         orderID,
					TransactionFlow: "debit",
					TransactionType: "return",
					Message:         &msg,
					Status:          "Success",
				}
				if err := tx.Create(&trx).Error; err != nil {
					return err
				}
			}

			// NO TEAM BONUSES - removed completely

			nowTime := time.Now()
			nextTime := nowTime.Add(24 * time.Hour)
			updates := map[string]interface{}{"total_paid": paid, "total_returned": returned, "last_return_at": nowTime, "next_return_at": nextTime}
			if paid >= inv.Duration {
				updates["status"] = "Completed"

				newBalance := round3(user.Balance + inv.Amount)
				if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
					return err
				}

				orderID := utils.GenerateOrderID(inv.UserID)
				msg := fmt.Sprintf("Pengembalian modal investasi produk %s", product.Name)
				trx := models.Transaction{
					UserID:          inv.UserID,
					Amount:          inv.Amount,
					Charge:          0,
					OrderID:         orderID,
					TransactionFlow: "debit",
					TransactionType: "return",
					Message:         &msg,
					Status:          "Success",
				}
				if err := tx.Create(&trx).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&inv).Updates(updates).Error; err != nil {
				return err
			}
			processed++
			return nil
		})
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Cron executed", Data: map[string]interface{}{"processed": processed}})
}

func round3(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// calculateVIPLevel determines VIP level based on total locked category investments
// VIP1: 50k, VIP2: 1.2M, VIP3: 7M, VIP4: 30M, VIP5: 150M
func calculateVIPLevel(totalInvestVIP float64) uint {
	if totalInvestVIP >= 150000000 {
		return 5
	} else if totalInvestVIP >= 30000000 {
		return 4
	} else if totalInvestVIP >= 7000000 {
		return 3
	} else if totalInvestVIP >= 1200000 {
		return 2
	} else if totalInvestVIP >= 50000 {
		return 1
	}
	return 0
}
