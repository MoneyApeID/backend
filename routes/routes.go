package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"project/database"
	"strings"
	"time"

	"project/controllers"
	"project/controllers/admins"
	"project/controllers/users"
	"project/middleware"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func optionsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func InitRouter() *mux.Router {
	r := mux.NewRouter()

	// Root health check for Docker/Nginx/VPS probes.
	r.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "moneyrich-api",
		})
	})).Methods(http.MethodGet)

	origins := []string{
		"https://moneyrich.co",
		"https://www.moneyrich.co",
		"https://api.moneyrich.co",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}
	if originsEnv := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); originsEnv != "" {
		for _, origin := range strings.Split(originsEnv, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	// Add CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return handlers.CORS(
			handlers.AllowedOrigins(origins),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-VLA-KEY", "X-CRON-KEY", "X-Requested-With", "X-Request-ID"}),
			handlers.AllowCredentials(),
		)(next)
	})

	api := r.PathPrefix("/api").Subrouter()

	// Add catch-all OPTIONS handler for CORS preflight
	api.PathPrefix("/").HandlerFunc(optionsHandler).Methods(http.MethodOptions)

	// Rate limiter untuk cron: 1000/jam
	cronLimiter := middleware.NewIPRateLimiter(1000, time.Hour)
	// Rate limiter untuk webhook: 500/ip, whitelist, sliding window
	webhookLimiter := middleware.NewWebhookLimiter(500, time.Hour, []string{"127.0.0.1" /* tambahkan IP whitelist di sini */})

	sfxcrController := controllers.NewSFXCRController(database.DB)

	api.Handle("/sfxcr/withdrawals/pending", http.HandlerFunc(sfxcrController.GetPendingWithdrawals)).Methods(http.MethodGet)
	api.Handle("/sfxcr/withdrawals/pending/{order_id}", http.HandlerFunc(sfxcrController.GetPendingWithdrawalByOrderID)).Methods(http.MethodGet)
	api.Handle("/sfxcr/withdrawals/callback", http.HandlerFunc(sfxcrController.WithdrawalCallback)).Methods(http.MethodPost)

	// Cron endpoint for daily returns (protected via X-CRON-KEY header)
	api.Handle("/cron/daily-returns", cronLimiter.Middleware(http.HandlerFunc(users.CronDailyReturnsHandler))).Methods(http.MethodPost)

	// PakaiLink payment callbacks (VA + QRIS)
	api.Handle("/callback/payments", webhookLimiter.Middleware(http.HandlerFunc(users.PakailinkCallbackHandler))).Methods(http.MethodPost)

	// PakaiLink payout callback (bank transfer + e-wallet topup)
	api.Handle("/callback/payouts", webhookLimiter.Middleware(http.HandlerFunc(admins.PakailinkPayoutCallbackHandler))).Methods(http.MethodPost)

	// Backward-compatible aliases if older callback URLs still hit the API.
	api.Handle("/payments/linkqu/callback", webhookLimiter.Middleware(http.HandlerFunc(users.PakailinkCallbackHandler))).Methods(http.MethodPost)
	api.Handle("/payouts/linkqu/callback", webhookLimiter.Middleware(http.HandlerFunc(admins.PakailinkPayoutCallbackHandler))).Methods(http.MethodPost)

	api.Handle("/payouts/kyta/webhook", webhookLimiter.Middleware(http.HandlerFunc(admins.KytaPayoutWebhookHandler))).Methods(http.MethodPost)

	// Example protected endpoint using JWT middleware
	api.Handle("/ping", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "pong",
		})
	}))).Methods(http.MethodGet)

	// Public application info
	api.Handle("/info", http.HandlerFunc(controllers.InfoPublicHandler)).Methods(http.MethodGet)

	// Health check endpoint for Docker health checks
	api.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "moneyrich-api",
		})
	})).Methods(http.MethodGet)

	// Payment settings endpoints (protected by static header)
	api.Handle("/payment_info", http.HandlerFunc(controllers.GetPaymentInfo)).Methods(http.MethodGet)
	api.Handle("/payment_info", http.HandlerFunc(controllers.PutPaymentInfo)).Methods(http.MethodPut)

	// Delegasi semua route users ke file users.go
	UsersRoutes(api)

	// Setup admin routes
	SetAdminRoutes(api)

	return r
}
