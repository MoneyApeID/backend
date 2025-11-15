package routes

import (
	"net/http"
	"time"

	"project/controllers/admins"
	"project/middleware"

	"github.com/gorilla/mux"
)

func SetAdminRoutes(api *mux.Router) {
	// Rate limiter for admin login: 5 attempts per IP per minute
	adminLoginLimiter := middleware.NewIPRateLimiter(5, time.Minute)

	// Public admin routes
	api.Handle("/admin/login", adminLoginLimiter.Middleware(http.HandlerFunc(admins.Login))).Methods(http.MethodPost)

	// Protected admin routes
	adminRouter := api.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminAuthMiddleware)

	// Dashboard stats
	adminRouter.Handle("/dashboard", http.HandlerFunc(admins.GetDashboardStats)).Methods(http.MethodGet)

	// Admin info
	adminRouter.Handle("/info", http.HandlerFunc(admins.GetAdminInfo)).Methods(http.MethodGet)

	// Admin profile
	adminRouter.Handle("/profile", http.HandlerFunc(admins.GetAdminProfile)).Methods(http.MethodGet)
	adminRouter.Handle("/profile", http.HandlerFunc(admins.UpdateAdminProfile)).Methods(http.MethodPut)
	adminRouter.Handle("/password", http.HandlerFunc(admins.UpdateAdminPassword)).Methods(http.MethodPut)

	// User management
	adminRouter.Handle("/users", http.HandlerFunc(admins.GetUsers)).Methods(http.MethodGet)
	adminRouter.Handle("/users/{id:[0-9]+}", http.HandlerFunc(admins.GetUserDetail)).Methods(http.MethodGet)
	adminRouter.Handle("/users/{id:[0-9]+}", http.HandlerFunc(admins.UpdateUser)).Methods(http.MethodPut)
	adminRouter.Handle("/users/balance/{id:[0-9]+}", http.HandlerFunc(admins.UpdateUserBalance)).Methods(http.MethodPut)
	adminRouter.Handle("/users/password/{id:[0-9]+}", http.HandlerFunc(admins.UpdateUserPassword)).Methods(http.MethodPut)

	// Investment management
	adminRouter.Handle("/investments", http.HandlerFunc(admins.GetInvestments)).Methods(http.MethodGet)
	adminRouter.Handle("/investments/{id:[0-9]+}", http.HandlerFunc(admins.GetInvestmentDetail)).Methods(http.MethodGet)
	adminRouter.Handle("/investments/{id:[0-9]+}/status", http.HandlerFunc(admins.UpdateInvestmentStatus)).Methods(http.MethodPut)

	// Category management
	adminRouter.Handle("/categories", http.HandlerFunc(admins.ListCategoriesHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/categories", http.HandlerFunc(admins.CreateCategoryHandler)).Methods(http.MethodPost)
	adminRouter.Handle("/categories/{id:[0-9]+}", http.HandlerFunc(admins.GetCategoryHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/categories/{id:[0-9]+}", http.HandlerFunc(admins.UpdateCategoryHandler)).Methods(http.MethodPut)
	adminRouter.Handle("/categories/{id:[0-9]+}", http.HandlerFunc(admins.DeleteCategoryHandler)).Methods(http.MethodDelete)

	// Product management
	adminRouter.Handle("/products", http.HandlerFunc(admins.ListProductsHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/products", http.HandlerFunc(admins.CreateProductHandler)).Methods(http.MethodPost)
	adminRouter.Handle("/products/{id:[0-9]+}", http.HandlerFunc(admins.GetProductHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/products/{id:[0-9]+}", http.HandlerFunc(admins.UpdateProductHandler)).Methods(http.MethodPut)
	adminRouter.Handle("/products/{id:[0-9]+}", http.HandlerFunc(admins.DeleteProductHandler)).Methods(http.MethodDelete)

	//Withdrawal management
	adminRouter.Handle("/withdrawals", http.HandlerFunc(admins.GetWithdrawals)).Methods(http.MethodGet)
	adminRouter.Handle("/withdrawals/{id:[0-9]+}/approve", http.HandlerFunc(admins.ApproveWithdrawal)).Methods(http.MethodPut)
	adminRouter.Handle("/withdrawals/{id:[0-9]+}/reject", http.HandlerFunc(admins.RejectWithdrawal)).Methods(http.MethodPut)

	// Bank management
	adminRouter.Handle("/banks", http.HandlerFunc(admins.GetBanks)).Methods(http.MethodGet)
	adminRouter.Handle("/banks", http.HandlerFunc(admins.CreateBank)).Methods(http.MethodPost)
	adminRouter.Handle("/banks/{id:[0-9]+}", http.HandlerFunc(admins.UpdateBank)).Methods(http.MethodPut)

	// Bank accounts management
	adminRouter.Handle("/bank-accounts", http.HandlerFunc(admins.GetBankAccounts)).Methods(http.MethodGet)

	// Transaction management
	adminRouter.Handle("/transactions", http.HandlerFunc(admins.GetTransactions)).Methods(http.MethodGet)

	// Payment management
	adminRouter.Handle("/payments", http.HandlerFunc(admins.GetPayments)).Methods(http.MethodGet)

	// Spin prize management
	adminRouter.Handle("/spin-prizes", http.HandlerFunc(admins.GetSpinPrizes)).Methods(http.MethodGet)
	adminRouter.Handle("/spin-prizes/{id:[0-9]+}", http.HandlerFunc(admins.UpdateSpinPrize)).Methods(http.MethodPut)

	// Task management
	adminRouter.Handle("/tasks", http.HandlerFunc(admins.TaskListHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/tasks", http.HandlerFunc(admins.CreateTaskHandler)).Methods(http.MethodPost)
	adminRouter.Handle("/tasks/{id:[0-9]+}", http.HandlerFunc(admins.UpdateTaskHandler)).Methods(http.MethodPut)

	adminRouter.Handle("/user-tasks", http.HandlerFunc(admins.UserTasksHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/user-spins", http.HandlerFunc(admins.UserSpinsHandler)).Methods(http.MethodGet)

	// Forum management
	adminRouter.Handle("/forums", http.HandlerFunc(admins.GetForumsHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/forums/{id:[0-9]+}/approve", http.HandlerFunc(admins.ApproveForumHandler)).Methods(http.MethodPut)
	adminRouter.Handle("/forums/{id:[0-9]+}/reject", http.HandlerFunc(admins.RejectForumHandler)).Methods(http.MethodPut)

	// Settings management
	adminRouter.Handle("/settings", http.HandlerFunc(admins.GetSettingsHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/settings", http.HandlerFunc(admins.UpdateSettingsHandler)).Methods(http.MethodPut)

	// Tutorial management
	adminRouter.Handle("/tutorials", http.HandlerFunc(admins.CreateTutorialHandler)).Methods(http.MethodPost)
	adminRouter.Handle("/tutorials", http.HandlerFunc(admins.ListTutorialsHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/tutorials", http.HandlerFunc(admins.UpdateTutorialHandler)).Methods(http.MethodPut)
	adminRouter.Handle("/tutorials/{id:[0-9]+}", http.HandlerFunc(admins.DeleteTutorialHandler)).Methods(http.MethodDelete)

	// Binary system management
	adminRouter.Handle("/binary", http.HandlerFunc(admins.GetBinaryStructureAdminHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/binary/details/{id:[0-9]+}", http.HandlerFunc(admins.GetBinaryDetailsAdminHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/binary/claim", http.HandlerFunc(admins.ClaimRewardAdminHandler)).Methods(http.MethodPost)
	adminRouter.Handle("/binary/rewards", http.HandlerFunc(admins.GetBinaryRewardsAdminHandler)).Methods(http.MethodGet)

	// Reward management
	adminRouter.Handle("/rewards", http.HandlerFunc(admins.ListRewardsHandler)).Methods(http.MethodGet)
	adminRouter.Handle("/rewards", http.HandlerFunc(admins.CreateRewardHandler)).Methods(http.MethodPost)
	adminRouter.Handle("/rewards", http.HandlerFunc(admins.UpdateRewardHandler)).Methods(http.MethodPut)
	adminRouter.Handle("/rewards/{id:[0-9]+}", http.HandlerFunc(admins.DeleteRewardHandler)).Methods(http.MethodDelete)

	// Deposit management
	adminRouter.Handle("/deposits", http.HandlerFunc(admins.GetDeposits)).Methods(http.MethodGet)
	adminRouter.Handle("/deposits/{id:[0-9]+}/approve", http.HandlerFunc(admins.ApproveDeposit)).Methods(http.MethodPut)
	adminRouter.Handle("/deposits/{id:[0-9]+}/reject", http.HandlerFunc(admins.RejectDeposit)).Methods(http.MethodPut)
}
