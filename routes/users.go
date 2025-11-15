package routes

import (
	"net/http"
	"project/controllers"
	"project/controllers/auth"
	"project/controllers/users"
	"project/middleware"
	"time"

	"github.com/gorilla/mux"
)

// UsersRoutes mendaftarkan semua route terkait user ke subrouter yang diberikan
func UsersRoutes(api *mux.Router) {
	// Active investments by product
	// Rate limiter login/register: 10 per IP per menit
	loginLimiter := middleware.NewIPRateLimiter(10, time.Minute)
	// Rate limiter session: 120 per user per menit (GET), 60 per user per menit (POST/PUT/DELETE)
	userLimiter := middleware.NewUserRateLimiter(120, 60, 60) // 120 read, 60 write, window 60 detik

	// Register & Login
	api.Handle("/register", loginLimiter.Middleware(http.HandlerFunc(auth.RegisterHandler))).Methods(http.MethodPost)
	api.Handle("/login", loginLimiter.Middleware(http.HandlerFunc(auth.LoginHandler))).Methods(http.MethodPost)
	api.Handle("/refresh", loginLimiter.Middleware(http.HandlerFunc(auth.RefreshHandler))).Methods(http.MethodPost)
	api.Handle("/logout", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(auth.LogoutHandler)))).Methods(http.MethodPost)
	api.Handle("/logout-all", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(auth.LogoutAllHandler)))).Methods(http.MethodPost)

	// Change password (write)
	api.Handle("/users/change-password", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ChangePasswordHandler)))).Methods(http.MethodPost)

	// User info (read)
	api.Handle("/users/info", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.InfoHandler)))).Methods(http.MethodGet)

	// Get Bank List, Add, Edit, Delete
	api.Handle("/bank", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.BankListHandler)))).Methods(http.MethodGet)
	api.Handle("/users/bank", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.AddBankAccountHandler)))).Methods(http.MethodPost)
	api.Handle("/users/bank", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetBankAccountHandler)))).Methods(http.MethodGet)
	api.Handle("/users/bank/{id}", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetBankAccountHandler)))).Methods(http.MethodGet)
	api.Handle("/users/bank", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.EditBankAccountHandler)))).Methods(http.MethodPut)
	api.Handle("/users/bank", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.DeleteBankAccountHandler)))).Methods(http.MethodDelete)

	// Public: list products
	api.Handle("/products", userLimiter.Middleware(http.HandlerFunc(controllers.ProductListHandler))).Methods(http.MethodGet)

	// Investment endpoints (replace deposit flow)
	api.Handle("/users/investments", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.CreateInvestmentHandler)))).Methods(http.MethodPost)
	api.Handle("/users/investments", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ListInvestmentsHandler)))).Methods(http.MethodGet)
	api.Handle("/users/investments/active", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetActiveInvestmentsHandler)))).Methods(http.MethodGet)
	api.Handle("/users/investments/{id:[0-9]+}", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetInvestmentHandler)))).Methods(http.MethodGet)

	// Deposit endpoints
	api.Handle("/users/deposits", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.CreateDepositHandler)))).Methods(http.MethodPost)
	api.Handle("/users/deposits", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ListDepositsHandler)))).Methods(http.MethodGet)

	// Handle Payments get
	api.Handle("/users/payments/{order_id}", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetDepositDetailsHandler)))).Methods(http.MethodGet)

	// Protected endpoint: withdrawal request
	api.Handle("/users/withdrawal", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.WithdrawalHandler)))).Methods(http.MethodPost)
	api.Handle("/users/withdrawal", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ListWithdrawalHandler)))).Methods(http.MethodGet)

	// Spin endpoints
	api.Handle("/spin-prize-list", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.SpinPrizeListHandler)))).Methods(http.MethodGet)
	api.Handle("/users/spin", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.UserSpinHandler)))).Methods(http.MethodPost)
	//api.Handle("/users/spin-v2", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.UserSpinHandler)))).Methods(http.MethodGet)

	api.Handle("/users/transaction", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetTransactionHistory)))).Methods(http.MethodGet)
	api.Handle("/users/transaction/{type}", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetTransactionHistory)))).Methods(http.MethodGet)

	api.Handle("/users/team-invited", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.TeamInvitedHandler)))).Methods(http.MethodGet)
	api.Handle("/users/team-invited/{level}", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.TeamInvitedHandler)))).Methods(http.MethodGet)
	api.Handle("/users/team-data/{level}", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.TeamDataHandler)))).Methods(http.MethodGet)

	api.Handle("/users/forum", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ForumListHandler)))).Methods(http.MethodGet)
	api.Handle("/users/check-forum", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.CheckWithdrawalForumHandler)))).Methods(http.MethodGet)
	api.Handle("/users/forum/submit", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ForumSubmitHandler)))).Methods(http.MethodPost)

	api.Handle("/users/task", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.TaskListHandler)))).Methods(http.MethodGet)
	api.Handle("/users/task/submit", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.TaskSubmitHandler)))).Methods(http.MethodPost)

	// Tutorial endpoints
	api.Handle("/users/tutorials", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.ListTutorialsHandler)))).Methods(http.MethodGet)

	// Popup endpoint
	api.Handle("/users/popup", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetPopupHandler)))).Methods(http.MethodGet)

	// Binary system endpoints
	api.Handle("/users/binary/structure", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetBinaryStructureHandler)))).Methods(http.MethodGet)
	api.Handle("/users/binary/omset", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetBinaryOmsetHandler)))).Methods(http.MethodGet)
	api.Handle("/users/rewards", userLimiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(users.GetRewardsHandler)))).Methods(http.MethodGet)
}
