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

		newBalance := round3(userForUpdate.Balance - product.Amount)
		finalBalance = newBalance
		newTotalInvest := round3(userForUpdate.TotalInvest + product.Amount)
		updateFields := map[string]interface{}{
			"balance":            newBalance,
			"total_invest":       newTotalInvest,
			"total_invest_vip":   newTotalInvest,
			"investment_status":  "Active",
			"updated_at":         time.Now(),
		}

		if err := tx.Model(&models.User{}).Where("id = ?", uid).Updates(updateFields).Error; err != nil {
			return err
		}

		// Update VIP level if needed for locked category
		newLevel := calculateVIPLevel(newTotalInvest)
		if userForUpdate.Level == nil || *userForUpdate.Level != newLevel {
			if err := tx.Model(&models.User{}).Where("id = ?", uid).Update("level", newLevel).Error; err != nil {
				return err
			}
		}

		// Initialize reward progress jika user baru memiliki investasi aktif
		if userForUpdate.InvestmentStatus == "Inactive" {
			// User baru aktif, initialize reward progress
			// Note: Ini dilakukan setelah transaction commit
		}
		
		msg := fmt.Sprintf("Berhasil melakukan investasi pada produk %s", product.Name)
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

		// Bonus 15% untuk user yang membeli investasi (dalam bentuk Income)
		purchaseBonus := round3(product.Amount * 0.15)
		if err := tx.Model(&models.User{}).Where("id = ?", uid).UpdateColumn("income", gorm.Expr("income + ?", purchaseBonus)).Error; err != nil {
			return err
		}
		msgBonus := fmt.Sprintf("Bonus pembelian investasi produk %s", product.Name)
		bonusTrx := models.Transaction{
			UserID:          uid,
			Amount:          purchaseBonus,
			Charge:          0,
			OrderID:         utils.GenerateOrderID(uid),
			TransactionFlow: "debit",
			TransactionType: "investment",
			Message:         &msgBonus,
			Status:          "Success",
		}
		if err := tx.Create(&bonusTrx).Error; err != nil {
			return err
		}

		// Handle referral bonus (level 1)
		if userForUpdate.ReffBy != nil {
			var refUser1 models.User
			if err := tx.Select("id, spin_ticket, income").Where("id = ?", *userForUpdate.ReffBy).First(&refUser1).Error; err == nil {
				if product.Amount >= 100000 {
					if refUser1.SpinTicket == nil {
						one := uint(1)
						if err := tx.Model(&models.User{}).Where("id = ?", refUser1.ID).Update("spin_ticket", one).Error; err != nil {
							return err
						}
					} else {
						if err := tx.Model(&models.User{}).Where("id = ?", refUser1.ID).UpdateColumn("spin_ticket", gorm.Expr("spin_ticket + ?", 1)).Error; err != nil {
							return err
						}
					}
				}

				bonus := round3(product.Amount * 0.15)
				if err := tx.Model(&models.User{}).Where("id = ?", refUser1.ID).UpdateColumn("income", gorm.Expr("income + ?", bonus)).Error; err != nil {
					return err
				}
				msgBonus := "Bonus Rujukan (Sponsor Bonus) Level 1"
				refTrx := models.Transaction{
					UserID:          refUser1.ID,
					Amount:          bonus,
					Charge:          0,
					OrderID:         utils.GenerateOrderID(refUser1.ID),
					TransactionFlow: "debit",
					TransactionType: "team",
					Message:         &msgBonus,
					Status:          "Success",
				}
				if err := tx.Create(&refTrx).Error; err != nil {
					return err
				}

				// Level 2: inviter of level 1
				if refUser1.ReffBy != nil {
					var refUser2 models.User
					if err := tx.Select("id, reff_by").Where("id = ?", *refUser1.ReffBy).First(&refUser2).Error; err == nil {
						bonus2 := round3(product.Amount * 0.02)
						if err := tx.Model(&models.User{}).Where("id = ?", refUser2.ID).UpdateColumn("income", gorm.Expr("income + ?", bonus2)).Error; err != nil {
							return err
						}
						msg2 := "Bonus Rujukan (Sponsor Bonus) Level 2"
						refTrx2 := models.Transaction{
							UserID:          refUser2.ID,
							Amount:          bonus2,
							Charge:          0,
							OrderID:         utils.GenerateOrderID(refUser2.ID),
							TransactionFlow: "debit",
							TransactionType: "team",
							Message:         &msg2,
							Status:          "Success",
						}
						if err := tx.Create(&refTrx2).Error; err != nil {
							return err
						}

						// Level 3: inviter of level 2
						if refUser2.ReffBy != nil {
							var refUser3 models.User
							if err := tx.Select("id").Where("id = ?", *refUser2.ReffBy).First(&refUser3).Error; err == nil {
								bonus3 := round3(product.Amount * 0.01)
								if err := tx.Model(&models.User{}).Where("id = ?", refUser3.ID).UpdateColumn("income", gorm.Expr("income + ?", bonus3)).Error; err != nil {
									return err
								}
								msg3 := "Bonus Rujukan (Sponsor Bonus) Level 3"
								refTrx3 := models.Transaction{
									UserID:          refUser3.ID,
									Amount:          bonus3,
									Charge:          0,
									OrderID:         utils.GenerateOrderID(refUser3.ID),
									TransactionFlow: "debit",
									TransactionType: "team",
									Message:         &msg3,
									Status:          "Success",
								}
								if err := tx.Create(&refTrx3).Error; err != nil {
									return err
								}
							}
						}
					}
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

	// Initialize reward progress jika user baru memiliki investasi aktif (setelah transaction commit)
	var updatedUser models.User
	if err := db.Select("investment_status").Where("id = ?", uid).First(&updatedUser).Error; err == nil {
		if updatedUser.InvestmentStatus == "Active" {
			// Initialize reward progress untuk user ini
			_ = utils.InitializeRewardProgress(uid)
			// Update reward progress untuk user ini
			_ = utils.UpdateRewardProgress(uid)

			// Update reward progress untuk semua upline yang terpengaruh (level 1-3)
			if err := updateUplineRewardProgress(uid, db); err != nil {
				// Log error but don't fail the request
			}
		}
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
		err := db.Transaction(func(tx *gorm.DB) error {
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
				newBalance := round3(user.Income + amount)
				if err := tx.Model(&user).Update("income", newBalance).Error; err != nil {
					return err
				}

				orderID := utils.GenerateOrderID(inv.UserID)
				msg := fmt.Sprintf("Pengembalian profit investasi produk %s", product.Name)
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
				newBalance := round3(user.Income + totalProfit)
				if err := tx.Model(&user).Update("income", newBalance).Error; err != nil {
					return err
				}

				orderID := utils.GenerateOrderID(inv.UserID)
				msg := fmt.Sprintf("Pengembalian profit investasi produk %s selesai", product.Name)
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

			// Bonus manajemen tim level 1, 2, 3
			levelPercents := []float64{0.03, 0.03, 0.03}
			levelMsgs := []string{
				"Bonus Rebat Kedalaman Jaringan Level 1",
				"Bonus Rebat Kedalaman Jaringan Level 2",
				"Bonus Rebat Kedalaman Jaringan Level 3",
			}
			currReffBy := user.ReffBy
			for level := 0; level < 3 && currReffBy != nil; level++ {
				var reffUser models.User
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, income, reff_by, level").Where("id = ?", *currReffBy).First(&reffUser).Error; err != nil {
					break // stop if not found
				}
				bonus := round3(amount * levelPercents[level])
				if bonus > 0 {
					newReffBalance := round3(reffUser.Income + bonus)
					if err := tx.Model(&reffUser).Update("income", newReffBalance).Error; err != nil {
						return err
					}
					orderID := utils.GenerateOrderID(reffUser.ID)
					msg := levelMsgs[level]
					trxBonus := models.Transaction{
						UserID:          reffUser.ID,
						Amount:          bonus,
						Charge:          0,
						OrderID:         orderID,
						TransactionFlow: "debit",
						TransactionType: "team",
						Message:         &msg,
						Status:          "Success",
					}
					if err := tx.Create(&trxBonus).Error; err != nil {
						return err
					}
				}
				currReffBy = reffUser.ReffBy
			}

			nowTime := time.Now()
			nextTime := nowTime.Add(24 * time.Hour)
			updates := map[string]interface{}{"total_paid": paid, "total_returned": returned, "last_return_at": nowTime, "next_return_at": nextTime}
			if paid >= inv.Duration {
				updates["status"] = "Completed"
			}
			if err := tx.Model(&inv).Updates(updates).Error; err != nil {
				return err
			}
			processed++
			return nil
		})

		// Update reward progress setelah transaction commit (jika berhasil)
		if err == nil {
			// Update reward progress untuk user yang investasinya di-update
			_ = utils.UpdateRewardProgress(inv.UserID)
			// Update reward progress untuk semua upline yang terpengaruh
			_ = updateUplineRewardProgress(inv.UserID, db)
		}
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Cron executed", Data: map[string]interface{}{"processed": processed}})
}

func round3(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// updateUplineRewardProgress mengupdate reward progress untuk semua upline yang terpengaruh
// Upline yang terpengaruh adalah yang memiliki user ini di binary tree mereka (level 1-3)
func updateUplineRewardProgress(userID uint, db *gorm.DB) error {
	// Cari semua upline yang memiliki user ini di binary tree mereka
	// Kita perlu traverse ke atas untuk menemukan semua upline yang terpengaruh

	// Get user's referral chain (up to 3 levels)
	var user models.User
	if err := db.Select("reff_by").Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	// Collect all uplines (up to 3 levels)
	uplineIDs := make([]uint, 0)
	currentReffBy := user.ReffBy
	level := 0
	for currentReffBy != nil && level < 3 {
		uplineIDs = append(uplineIDs, *currentReffBy)
		var upline models.User
		if err := db.Select("reff_by").Where("id = ?", *currentReffBy).First(&upline).Error; err != nil {
			break
		}
		currentReffBy = upline.ReffBy
		level++
	}

	// Update reward progress for each upline
	for _, uplineID := range uplineIDs {
		// Check if upline has active investment
		var upline models.User
		if err := db.Select("investment_status").Where("id = ?", uplineID).First(&upline).Error; err != nil {
			continue
		}
		if upline.InvestmentStatus != "Active" {
			continue
		}

		// Update reward progress
		_ = utils.UpdateRewardProgress(uplineID)
	}

	return nil
}

func calculateVIPLevel(totalInvestVIP float64) uint {
	if totalInvestVIP >= 10000 {
		return 2
	}
	return 1
}
