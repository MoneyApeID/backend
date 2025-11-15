package admins

import (
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strings"
	"time"
)

type DailyGrowth struct {
	Day   string `json:"day"`
	Count *int64 `json:"count"`
}

type DailyInvestment struct {
	Day    string   `json:"day"`
	Amount *float64 `json:"amount"`
}

type TransactionDetail struct {
	UserName  string    `json:"user_name"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	Message   *string   `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type TypeTransactions struct {
	Investment *int64 `json:"investment"`
	Withdrawal *int64 `json:"withdrawal"`
	Return     *int64 `json:"return"`
	Team       *int64 `json:"team"`
	Bonus      *int64 `json:"bonus"`
}

type DashboardStats struct {
	TotalUsers          int64               `json:"total_users"`
	ActiveUsers         int64               `json:"active_users"`
	GrowthUsers         []DailyGrowth       `json:"growth_users"`
	TotalInvestments    int64               `json:"total_investments"`
	ActiveInvestments   int64               `json:"active_investments"`
	OverviewInvestments []DailyInvestment   `json:"overview_investments"`
	TotalWithdrawals    int64               `json:"total_withdrawals"`
	PendingWithdrawals  int64               `json:"pending_withdrawals"`
	TotalBalance        float64             `json:"total_balance"`
	TotalForums         int64               `json:"total_forums"`
	PendingForums       int64               `json:"pending_forums"`
	TypeTransactions    TypeTransactions    `json:"type_transactions"`
	LastTransactions    []TransactionDetail `json:"last_transactions"`
}

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	var stats DashboardStats
	db := database.DB

	// initialize slices to ensure empty arrays are returned (not null)
	stats.GrowthUsers = make([]DailyGrowth, 0)
	stats.OverviewInvestments = make([]DailyInvestment, 0)
	stats.LastTransactions = make([]TransactionDetail, 0)

	// Get total users count
	db.Model(&models.User{}).Count(&stats.TotalUsers)

	// Get active users count (users with active investments status)
	db.Model(&models.User{}).
		Where("investment_status = ?", "Active").
		Count(&stats.ActiveUsers)

	// Get growth users count by day (users created in the last 7 days)
	// Fetch counts grouped by day name
	growthMap := map[string]int64{}
	rows, err := db.Model(&models.User{}).
		Select("DATE_FORMAT(created_at, '%W') as day, COUNT(*) as count").
		Where("created_at >= NOW() - INTERVAL 7 DAY").
		Group("DATE_FORMAT(created_at, '%W')").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day string
			var count int64
			if scanErr := rows.Scan(&day, &count); scanErr == nil {
				growthMap[strings.TrimSpace(day)] = count
			}
		}
	}
	// Build last 7 days list (from 6 days ago to today)
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i)
		dayName := d.Format("Monday")
		if val, ok := growthMap[dayName]; ok {
			v := val
			stats.GrowthUsers = append(stats.GrowthUsers, DailyGrowth{Day: dayName, Count: &v})
		} else {
			stats.GrowthUsers = append(stats.GrowthUsers, DailyGrowth{Day: dayName, Count: nil})
		}
	}

	// Get total investments count (excluding pending)
	db.Model(&models.Investment{}).
		Where("status != ?", "Pending").
		Count(&stats.TotalInvestments)

	// Get overview investments amount by day with payment status "Success"
	investMap := map[string]float64{}
	rows, err = db.Model(&models.Investment{}).
		Select("DATE_FORMAT(investments.created_at, '%Y-%m-%d') as day, COALESCE(SUM(investments.amount), 0) as amount").
		Where("status IN (?) AND investments.created_at >= CURDATE() - INTERVAL 6 DAY", []string{"Running", "Completed", "Suspended"}).
		Group("DATE_FORMAT(investments.created_at, '%Y-%m-%d')").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day string
			var amount float64
			if scanErr := rows.Scan(&day, &amount); scanErr == nil {
				investMap[strings.TrimSpace(day)] = amount
			}
		}
	}
	// Build last 7 days list for investments using date keys (YYYY-MM-DD)
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i)
		dateKey := d.Format("2006-01-02") // matches SQL grouping
		dayName := d.Format("Monday")
		if val, ok := investMap[dateKey]; ok {
			v := val
			stats.OverviewInvestments = append(stats.OverviewInvestments, DailyInvestment{Day: dayName, Amount: &v})
		} else {
			stats.OverviewInvestments = append(stats.OverviewInvestments, DailyInvestment{Day: dayName, Amount: nil})
		}
	}

	// Get pending withdrawals count
	db.Model(&models.Withdrawal{}).
		Where("status = ?", "Pending").
		Count(&stats.PendingWithdrawals)

	// Get total balance of all users (balance + income)
	type Result struct {
		TotalBalance float64
	}
	var result Result
	db.Model(&models.User{}).
		Select("COALESCE(SUM(balance + income), 0) as total_balance").
		Scan(&result)
	stats.TotalBalance = result.TotalBalance

	// Get active investments count
	db.Model(&models.Investment{}).
		Where("status = ?", "Running").
		Count(&stats.ActiveInvestments)

	// Get total withdrawals count
	db.Model(&models.Withdrawal{}).
		Count(&stats.TotalWithdrawals)

	// Get total forums count
	db.Model(&models.Forum{}).
		Count(&stats.TotalForums)

	// Get pending forum count
	db.Model(&models.Forum{}).
		Where("status = ?", "Pending").
		Count(&stats.PendingForums)

	// Type transactions counts (set to null when zero)
	var cnt int64

	// investment
	cnt = 0
	db.Model(&models.Transaction{}).Where("transaction_type = ?", "investment").Count(&cnt)
	if cnt > 0 {
		val := cnt
		stats.TypeTransactions.Investment = &val
	}

	// withdrawal
	cnt = 0
	db.Model(&models.Transaction{}).Where("transaction_type = ?", "withdrawal").Count(&cnt)
	if cnt > 0 {
		val := cnt
		stats.TypeTransactions.Withdrawal = &val
	}

	// return
	cnt = 0
	db.Model(&models.Transaction{}).Where("transaction_type = ?", "return").Count(&cnt)
	if cnt > 0 {
		val := cnt
		stats.TypeTransactions.Return = &val
	}

	// team
	cnt = 0
	db.Model(&models.Transaction{}).Where("transaction_type = ?", "team").Count(&cnt)
	if cnt > 0 {
		val := cnt
		stats.TypeTransactions.Team = &val
	}

	// bonus
	cnt = 0
	db.Model(&models.Transaction{}).Where("transaction_type = ?", "bonus").Count(&cnt)
	if cnt > 0 {
		val := cnt
		stats.TypeTransactions.Bonus = &val
	}

	// Get last 10 transactions (join with users table to get user name)
	rows, err = db.Model(&models.Transaction{}).
		Select("users.name as user_name, transactions.amount, transactions.transaction_type, transactions.message, transactions.created_at").
		Joins("JOIN users ON transactions.user_id = users.id").
		Order("transactions.created_at DESC").
		Limit(10).
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var td TransactionDetail
			if scanErr := rows.Scan(&td.UserName, &td.Amount, &td.Type, &td.Message, &td.CreatedAt); scanErr == nil {
				stats.LastTransactions = append(stats.LastTransactions, td)
			}
		}
	}

	// Send response

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    stats,
	})
}
