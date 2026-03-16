package admins

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"project/database"
	"project/models"
	"project/utils"
)

type adminWithdrawInquiryRequest struct {
	BankCode      string  `json:"bank_code"`
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
}

// POST /api/admin/withdraw/inquiry
func AdminWithdrawInquiryHandler(w http.ResponseWriter, r *http.Request) {
	var req adminWithdrawInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Data tidak valid"})
		return
	}

	req.BankCode = strings.TrimSpace(req.BankCode)
	req.AccountNumber = strings.TrimSpace(req.AccountNumber)

	if req.BankCode == "" || req.AccountNumber == "" || req.Amount <= 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank, nomor rekening/customer, dan nominal wajib diisi"})
		return
	}

	// Look up bank from DB to determine type
	db := database.DB
	var bank models.Bank
	if err := db.Where("code = ? AND status = ?", req.BankCode, "Active").First(&bank).Error; err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank/ewallet tidak ditemukan atau tidak aktif"})
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, err := utils.GetPakailinkAccessToken(r.Context(), client)
	if err != nil {
		log.Printf("[AdminWithdraw] GetAccessToken error: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menghubungi payment gateway"})
		return
	}

	partnerRefNo := fmt.Sprintf("%d%04d", time.Now().UnixMilli(), rand.Intn(10000))
	// Pad with random characters to ensure it's around 32 chars to match the provider sample
	for len(partnerRefNo) < 32 {
		partnerRefNo += fmt.Sprintf("%x", rand.Intn(16))
	}

	adminFee := 2000.0
	finalAmount := req.Amount + adminFee
	isEwallet := bank.Type == "ewallet"

	respData := map[string]interface{}{
		"bank_code":      req.BankCode,
		"bank_name":      bank.Name,
		"bank_type":      bank.Type,
		"account_number": req.AccountNumber,
		"amount":         req.Amount,
		"admin_fee":      adminFee,
		"final_amount":   finalAmount,
		"partner_ref_no": partnerRefNo,
	}

	if isEwallet {
		// Ewallet inquiry
		inquiryResp, err := utils.PakailinkEwalletInquiry(r.Context(), client, accessToken, partnerRefNo, req.AccountNumber, req.BankCode)
		if err != nil {
			log.Printf("[AdminWithdraw] EwalletInquiry error: %v", err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Inquiry ewallet gagal: " + err.Error()})
			return
		}
		respData["session_id"] = inquiryResp.SessionID
		respData["account_name"] = inquiryResp.CustomerName
	} else {
		// Bank inquiry
		inquiryResp, err := utils.PakailinkBankInquiry(r.Context(), client, accessToken, partnerRefNo, req.AccountNumber, req.BankCode)
		if err != nil {
			log.Printf("[AdminWithdraw] BankInquiry error: %v", err)
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Inquiry bank gagal: " + err.Error()})
			return
		}
		respData["session_id"] = inquiryResp.SessionID
		respData["account_name"] = inquiryResp.BeneficiaryAccountName
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Inquiry berhasil",
		Data:    respData,
	})
}

type adminWithdrawTransferRequest struct {
	SessionID     string  `json:"session_id"`
	BankCode      string  `json:"bank_code"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount"`
	PartnerRefNo  string  `json:"partner_ref_no"`
}

// POST /api/admin/withdraw/transfer
func AdminWithdrawTransferHandler(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("admin_id").(uint)
	if !ok {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	var req adminWithdrawTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Data tidak valid"})
		return
	}

	req.BankCode = strings.TrimSpace(req.BankCode)
	req.AccountNumber = strings.TrimSpace(req.AccountNumber)
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.PartnerRefNo = strings.TrimSpace(req.PartnerRefNo)

	if req.BankCode == "" || req.AccountNumber == "" || req.Amount <= 0 || req.PartnerRefNo == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Data tidak lengkap"})
		return
	}

	// Look up bank type from DB
	db := database.DB
	var bank models.Bank
	if err := db.Where("code = ? AND status = ?", req.BankCode, "Active").First(&bank).Error; err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Bank/ewallet tidak ditemukan"})
		return
	}

	adminFee := 2000.0
	finalAmount := req.Amount + adminFee
	isEwallet := bank.Type == "ewallet"

	// Create record
	record := models.AdminWithdrawal{
		AdminID:       adminID,
		BankCode:      req.BankCode,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		Amount:        req.Amount,
		AdminFee:      adminFee,
		FinalAmount:   finalAmount,
		OrderID:       req.PartnerRefNo,
		SessionID:     req.SessionID,
		Status:        "Pending",
	}
	if err := db.Create(&record).Error; err != nil {
		log.Printf("[AdminWithdraw] Create record error: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menyimpan data"})
		return
	}

	// Execute transfer
	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, err := utils.GetPakailinkAccessToken(r.Context(), client)
	if err != nil {
		log.Printf("[AdminWithdraw] GetAccessToken error: %v", err)
		db.Model(&record).Update("status", "Failed")
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menghubungi payment gateway"})
		return
	}

	var transferStatus string

	if isEwallet {
		topupResp, err := utils.PakailinkEwalletTopup(r.Context(), client, accessToken, req.PartnerRefNo, req.AccountNumber, req.BankCode, req.SessionID, req.Amount, "")
		if err != nil {
			log.Printf("[AdminWithdraw] EwalletTopup error: %v", err)
			db.Model(&record).Update("status", "Failed")
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Topup ewallet gagal: " + err.Error()})
			return
		}
		transferStatus = "Success"
		if topupResp.AdditionalInfo.TransactionStatus == "Pending" {
			transferStatus = "Pending"
		}
	} else {
		transferResp, err := utils.PakailinkBankTransfer(r.Context(), client, accessToken, req.PartnerRefNo, req.AccountNumber, req.BankCode, req.SessionID, req.Amount, "")
		if err != nil {
			log.Printf("[AdminWithdraw] BankTransfer error: %v", err)
			db.Model(&record).Update("status", "Failed")
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Transfer bank gagal: " + err.Error()})
			return
		}
		transferStatus = "Success"
		if transferResp.AdditionalInfo.TransactionStatus == "Pending" {
			transferStatus = "Pending"
		}
	}

	db.Model(&record).Update("status", transferStatus)

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Transfer berhasil diproses",
		Data: map[string]interface{}{
			"id":           record.ID,
			"order_id":     record.OrderID,
			"status":       transferStatus,
			"amount":       req.Amount,
			"admin_fee":    adminFee,
			"final_amount": finalAmount,
		},
	})
}

// GET /api/admin/withdraw/history
func AdminWithdrawHistoryHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}

	var total int64
	db.Model(&models.AdminWithdrawal{}).Count(&total)

	var records []models.AdminWithdrawal
	db.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&records)

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"items": records,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total_rows":  total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}
