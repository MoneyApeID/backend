package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// LinkQuInquiryResponse untuk response inquiry LinkQu
type LinkQuInquiryResponse struct {
	BankCode      string  `json:"bankcode"`
	BankName      string  `json:"bankname"`
	AccountNumber string  `json:"accountnumber"`
	AccountName   string  `json:"accountname"`
	Amount        float64 `json:"amount"`
	AdditionalFee float64 `json:"additionalfee"`
	Status        string  `json:"status"`
	ResponseCode  string  `json:"response_code"`
	ResponseDesc  string  `json:"response_desc"`
	PartnerReff   string  `json:"partner_reff"`
	InquiryReff   int64   `json:"inquiry_reff"`
	Signature     string  `json:"signature"`
}

// LinkQuPaymentResponse untuk response payment LinkQu
type LinkQuPaymentResponse struct {
	BankCode      string  `json:"bankcode"`
	AccountNumber string  `json:"accountnumber"`
	AccountName   string  `json:"accountname"`
	Amount        float64 `json:"amount"`
	AdditionalFee float64 `json:"additionalfee"`
	Status        string  `json:"status"`
	ResponseCode  string  `json:"response_code"`
	ResponseDesc  string  `json:"response_desc"`
	PartnerReff   string  `json:"partner_reff"`
	InquiryReff   int64   `json:"inquiry_reff"`
	PaymentReff   int64   `json:"payment_reff"`
	TotalCost     float64 `json:"totalcost"`
	BankName      string  `json:"bankname"`
	Signature     string  `json:"signature"`
}

// IsEwallet mengecek apakah bankCode adalah e-wallet
func IsEwallet(bankCode string) bool {
	ewallets := []string{"DANA", "GOPAY", "OVO", "LINKAJA", "SHOPEEPAY", "KASPRO"}
	for _, ew := range ewallets {
		if strings.ToUpper(bankCode) == strings.ToUpper(ew) {
			return true
		}
	}
	return false
}

// LinkQuInquiryBank melakukan inquiry untuk bank
func LinkQuInquiryBank(bankCode, accountNumber string, amount float64, orderID string) (*LinkQuInquiryResponse, error) {
	baseURL := os.Getenv("LINKQU_BASE_URL")
	username := os.Getenv("LINKQU_USERNAME")
	pin := os.Getenv("LINKQU_PIN")
	clientID := os.Getenv("LINKQU_CLIENT_ID")
	clientSecret := os.Getenv("LINKQU_CLIENT_SECRET")

	if baseURL == "" || username == "" || pin == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("konfigurasi LinkQu tidak lengkap")
	}

	baseURL = strings.TrimRight(baseURL, "/")
	url := baseURL + "/linkqu-partner/transaction/withdraw/inquiry"

	body := map[string]interface{}{
		"username":      username,
		"pin":           pin,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        int64(amount),
		"partner_reff":  orderID,
	}

	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-id", clientID)
	req.Header.Set("client-secret", clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("koneksi gagal: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response: %v", err)
	}

	var inquiryResp LinkQuInquiryResponse
	if err := json.Unmarshal(respBody, &inquiryResp); err != nil {
		return nil, fmt.Errorf("gagal parsing response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, inquiryResp.ResponseDesc)
	}

	// Check response code
	if inquiryResp.ResponseCode != "00" {
		return nil, fmt.Errorf("inquiry gagal: %s", inquiryResp.ResponseDesc)
	}

	return &inquiryResp, nil
}

// LinkQuInquiryEwallet melakukan inquiry untuk e-wallet
func LinkQuInquiryEwallet(bankCode, accountNumber string, amount float64, orderID string) (*LinkQuInquiryResponse, error) {
	baseURL := os.Getenv("LINKQU_BASE_URL")
	username := os.Getenv("LINKQU_USERNAME")
	pin := os.Getenv("LINKQU_PIN")
	clientID := os.Getenv("LINKQU_CLIENT_ID")
	clientSecret := os.Getenv("LINKQU_CLIENT_SECRET")

	if baseURL == "" || username == "" || pin == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("konfigurasi LinkQu tidak lengkap")
	}

	baseURL = strings.TrimRight(baseURL, "/")
	url := baseURL + "/linkqu-partner/transaction/reload/inquiry"

	body := map[string]interface{}{
		"username":      username,
		"pin":           pin,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        int64(amount),
		"partner_reff":  orderID,
	}

	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-id", clientID)
	req.Header.Set("client-secret", clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("koneksi gagal: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response: %v", err)
	}

	var inquiryResp LinkQuInquiryResponse
	if err := json.Unmarshal(respBody, &inquiryResp); err != nil {
		return nil, fmt.Errorf("gagal parsing response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, inquiryResp.ResponseDesc)
	}

	// Check response code
	if inquiryResp.ResponseCode != "00" {
		return nil, fmt.Errorf("inquiry gagal: %s", inquiryResp.ResponseDesc)
	}

	return &inquiryResp, nil
}

// LinkQuPaymentBank melakukan payment untuk bank
func LinkQuPaymentBank(bankCode, accountNumber string, amount float64, orderID string, inquiryReff int64) (*LinkQuPaymentResponse, error) {
	baseURL := os.Getenv("LINKQU_BASE_URL")
	username := os.Getenv("LINKQU_USERNAME")
	pin := os.Getenv("LINKQU_PIN")
	clientID := os.Getenv("LINKQU_CLIENT_ID")
	clientSecret := os.Getenv("LINKQU_CLIENT_SECRET")
	callbackURL := os.Getenv("LINKQU_CALLBACK_PAYOUT")

	if baseURL == "" || username == "" || pin == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("konfigurasi LinkQu tidak lengkap")
	}

	baseURL = strings.TrimRight(baseURL, "/")
	url := baseURL + "/linkqu-partner/transaction/withdraw/payment"

	body := map[string]interface{}{
		"username":      username,
		"pin":           pin,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        int64(amount),
		"partner_reff":  orderID,
		"inquiry_reff":  inquiryReff,
		"url_callback":  callbackURL,
	}

	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-id", clientID)
	req.Header.Set("client-secret", clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("koneksi gagal: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response: %v", err)
	}

	var paymentResp LinkQuPaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, fmt.Errorf("gagal parsing response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, paymentResp.ResponseDesc)
	}

	return &paymentResp, nil
}

// LinkQuPaymentEwallet melakukan payment untuk e-wallet
func LinkQuPaymentEwallet(bankCode, accountNumber string, amount float64, orderID string, inquiryReff int64) (*LinkQuPaymentResponse, error) {
	baseURL := os.Getenv("LINKQU_BASE_URL")
	username := os.Getenv("LINKQU_USERNAME")
	pin := os.Getenv("LINKQU_PIN")
	clientID := os.Getenv("LINKQU_CLIENT_ID")
	clientSecret := os.Getenv("LINKQU_CLIENT_SECRET")
	callbackURL := os.Getenv("LINKQU_CALLBACK_PAYOUT")

	if baseURL == "" || username == "" || pin == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("konfigurasi LinkQu tidak lengkap")
	}

	baseURL = strings.TrimRight(baseURL, "/")
	url := baseURL + "/linkqu-partner/transaction/reload/payment"

	body := map[string]interface{}{
		"username":      username,
		"pin":           pin,
		"bankcode":      bankCode,
		"accountnumber": accountNumber,
		"amount":        int64(amount),
		"partner_reff":  orderID,
		"inquiry_reff":  inquiryReff,
		"url_callback":  callbackURL,
	}

	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client-id", clientID)
	req.Header.Set("client-secret", clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("koneksi gagal: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response: %v", err)
	}

	var paymentResp LinkQuPaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, fmt.Errorf("gagal parsing response: %v", err)
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, paymentResp.ResponseDesc)
	}

	return &paymentResp, nil
}

