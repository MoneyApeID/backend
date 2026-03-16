package middleware

import (
	"context"
	"fmt"
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strings"
)

// AdminAuthMiddleware verifies that the request is from an authenticated admin
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
				Success: false,
				Message: "Unauthorized: No token provided",
			})
			return
		}

		// Extract token string
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		// Use centralized validation which checks aud/iss/exp/nbf and revocation
		_, claims, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
				Success: false,
				Message: "Unauthorized: Invalid token",
			})
			return
		}

		// Verify role is admin
		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			utils.WriteJSON(w, http.StatusForbidden, utils.APIResponse{
				Success: false,
				Message: "Forbidden: Admin access required",
			})
			return
		}

		// Get admin ID (support float64 from JSON numbers)
		var adminID int64
		if rawID, ok := claims["id"]; ok {
			switch v := rawID.(type) {
			case float64:
				adminID = int64(v)
			case int:
				adminID = int64(v)
			case int64:
				adminID = v
			case string:
				// best-effort parse
				var n int64
				_, _ = fmt.Sscanf(v, "%d", &n)
				adminID = n
			}
		}

		// Verify admin exists and is active
		var admin models.Admin
		if err := database.DB.First(&admin, adminID).Error; err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
				Success: false,
				Message: "Unauthorized: Admin not found",
			})
			return
		}

		if !admin.IsActive {
			utils.WriteJSON(w, http.StatusForbidden, utils.APIResponse{
				Success: false,
				Message: "Forbidden",
			})
			return
		}

		// Admin is authenticated, pass admin_id in context
		ctx := context.WithValue(r.Context(), "admin_id", uint(admin.ID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
