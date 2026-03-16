package utils

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPakailinkBaseURL   = "https://api.pakailink.id"
	defaultPakailinkChannelID = "95222"
	defaultPakailinkTerminal  = "1701103543259012"
)

type pakailinkConfig struct {
	BaseURL            string
	ClientKey          string
	ClientSecret       string
	PartnerID          string
	PrivateKeyPath     string
	PaymentCallbackURL string
	PayoutCallbackURL  string
	MerchantID         string
	StoreID            string
	TerminalID         string
	ChannelID          string
}

func getPakailinkConfig() (*pakailinkConfig, error) {
	cfg := &pakailinkConfig{
		BaseURL:            strings.TrimRight(getenv("PAKAILINK_BASE_URL", defaultPakailinkBaseURL), "/"),
		ClientKey:          strings.TrimSpace(os.Getenv("PAKAILINK_CLIENT_KEY")),
		ClientSecret:       strings.TrimSpace(os.Getenv("PAKAILINK_CLIENT_SECRET")),
		PartnerID:          strings.TrimSpace(os.Getenv("PAKAILINK_PARTNER_ID")),
		PrivateKeyPath:     strings.TrimSpace(os.Getenv("PAKAILINK_PRIVATE_KEY_PATH")),
		PaymentCallbackURL: strings.TrimSpace(firstNonEmpty("PAKAILINK_PAYMENT_CALLBACK_URL", "PAKAILINK_CALLBACK_URL")),
		PayoutCallbackURL:  strings.TrimSpace(firstNonEmpty("PAKAILINK_PAYOUT_CALLBACK_URL", "PAKAILINK_CALLBACK_URL")),
		MerchantID:         strings.TrimSpace(os.Getenv("PAKAILINK_MERCHANT_ID")),
		StoreID:            strings.TrimSpace(os.Getenv("PAKAILINK_STORE_ID")),
		TerminalID:         strings.TrimSpace(getenv("PAKAILINK_TERMINAL_ID", defaultPakailinkTerminal)),
		ChannelID:          strings.TrimSpace(getenv("PAKAILINK_CHANNEL_ID", defaultPakailinkChannelID)),
	}

	if cfg.ClientKey == "" || cfg.ClientSecret == "" || cfg.PartnerID == "" || cfg.PrivateKeyPath == "" {
		return nil, fmt.Errorf("konfigurasi PakaiLink belum lengkap")
	}

	return cfg, nil
}

func firstNonEmpty(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

var VABankCodes = map[string]string{
	"BCA": "014", "BNI": "009", "BRI": "002", "BSI": "451", "CIMB": "022",
	"DANAMON": "011", "MANDIRI": "008", "BMI": "147", "BNC": "490",
	"OCBC": "028", "PERMATA": "013", "SINARMAS": "153", "PANIN": "019", "MAYBANK": "016",
}

func GetVABankCode(channel string) string {
	normalized := strings.ToUpper(strings.TrimSpace(channel))
	if code, ok := VABankCodes[normalized]; ok {
		return code
	}
	return normalized
}

func IsSupportedVABankChannel(channel string) bool {
	_, ok := VABankCodes[strings.ToUpper(strings.TrimSpace(channel))]
	return ok
}

func IsPakailinkEwalletCode(code string) bool {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "DANA", "GOPAY", "OVO", "SHOPEEPAY", "LINKAJA":
		return true
	default:
		return false
	}
}

func PakailinkTimestamp() string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc).Format("2006-01-02T15:04:05-07:00")
}

func createAsymmetricSignature(stringToSign, privateKeyPath string) (string, error) {
	keyData, err := os.ReadFile(filepath.Clean(privateKeyPath))
	if err != nil {
		return "", fmt.Errorf("gagal membaca private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return "", fmt.Errorf("format private key tidak valid")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	if err != nil {
		return "", err
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key bukan RSA")
	}

	sum := sha256.Sum256([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func createSymmetricSignature(method, path, accessToken string, body []byte, timestamp, clientSecret string) string {
	bodyHash := sha256.Sum256(body)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))
	stringToSign := method + ":" + path + ":" + accessToken + ":" + bodyHashHex + ":" + timestamp

	mac := hmac.New(sha512.New, []byte(clientSecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}

func minifyJSON(body []byte) []byte {
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return body
	}

	minified, err := json.Marshal(payload)
	if err != nil {
		return body
	}

	return minified
}

func generateExternalID() string {
	return strconv.FormatInt(time.Now().UnixNano()%10000000000, 10)
}

type PakailinkAccessTokenResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	AccessToken     string `json:"accessToken"`
	TokenType       string `json:"tokenType"`
	ExpiresIn       string `json:"expiresIn"`
}

func GetPakailinkAccessToken(ctx context.Context, client *http.Client) (string, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return "", err
	}

	path := "/snap/v1.0/access-token/b2b"
	url := cfg.BaseURL + path
	timestamp := PakailinkTimestamp()
	stringToSign := cfg.ClientKey + "|" + timestamp
	signature, err := createAsymmetricSignature(stringToSign, cfg.PrivateKeyPath)
	if err != nil {
		return "", err
	}

	body := []byte(`{"grantType":"client_credentials"}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-CLIENT-KEY", cfg.ClientKey)
	req.Header.Set("X-SIGNATURE", signature)

	log.Printf("[Pakailink] AccessToken request URL=%s X-TIMESTAMP=%s X-CLIENT-KEY=%s", url, timestamp, cfg.ClientKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request token PakaiLink gagal: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] AccessToken response status=%d body=%s", resp.StatusCode, string(respBody))
	for name, values := range resp.Header {
		for _, v := range values {
			log.Printf("[Pakailink] AccessToken response header: %s=%s", name, v)
		}
	}
	var result PakailinkAccessTokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("gagal parse response token PakaiLink: %w", err)
	}

	if result.ResponseCode != "2007300" || strings.TrimSpace(result.AccessToken) == "" {
		return "", fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return result.AccessToken, nil
}

type PakailinkCreateVAResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	VirtualAccountData struct {
		VirtualAccountNo   string `json:"virtualAccountNo"`
		PartnerReferenceNo string `json:"partnerReferenceNo"`
		ExpiredDate        string `json:"expiredDate"`
		TotalAmount        struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalAmount"`
		AdditionalInfo struct {
			CallbackURL string `json:"callbackUrl"`
			BankCode    string `json:"bankCode"`
			ReferenceNo string `json:"referenceNo"`
		} `json:"additionalInfo"`
	} `json:"virtualAccountData"`
}

func CreatePakailinkVA(ctx context.Context, client *http.Client, accessToken, partnerRefNo, customerNo, virtualAccountName string, amount float64, bankCode string) (*PakailinkCreateVAResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}
	if cfg.PaymentCallbackURL == "" {
		return nil, fmt.Errorf("PAKAILINK callback pembayaran belum diatur")
	}

	path := "/snap/v1.0/transfer-va/create-va"
	url := cfg.BaseURL + path
	loc, _ := time.LoadLocation("Asia/Jakarta")
	expired := time.Now().In(loc).Add(24 * time.Hour).Format(time.RFC3339)

	bodyObject := map[string]interface{}{
		"partnerReferenceNo": partnerRefNo,
		"customerNo":         customerNo,
		"virtualAccountName": virtualAccountName,
		"expiredDate":        expired,
		"totalAmount":        map[string]string{"value": fmt.Sprintf("%.2f", amount), "currency": "IDR"},
		"additionalInfo": map[string]string{
			"callbackUrl": cfg.PaymentCallbackURL,
			"bankCode":    bankCode,
		},
	}

	body, _ := json.Marshal(bodyObject)
	body = minifyJSON(body)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] CreateVA request URL=%s body=%s", url, string(body))
	log.Printf("[Pakailink] CreateVA request headers: X-TIMESTAMP=%s, X-PARTNER-ID=%s, CHANNEL-ID=%s", timestamp, cfg.PartnerID, cfg.ChannelID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] CreateVA response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkCreateVAResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2002700" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

type PakailinkCreateQRResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	QRContent          string `json:"qrContent"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ValidityPeriod     string `json:"validityPeriod"`
	Amount             struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

func CreatePakailinkQRIS(ctx context.Context, client *http.Client, accessToken, partnerRefNo string, amount float64) (*PakailinkCreateQRResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}
	if cfg.PaymentCallbackURL == "" {
		return nil, fmt.Errorf("PAKAILINK callback pembayaran belum diatur")
	}
	if cfg.MerchantID == "" || cfg.StoreID == "" || cfg.TerminalID == "" {
		return nil, fmt.Errorf("konfigurasi merchant QRIS PakaiLink belum lengkap")
	}

	path := "/snap/v1.0/qr/qr-mpm-generate"
	url := cfg.BaseURL + path
	loc, _ := time.LoadLocation("Asia/Jakarta")
	validity := time.Now().In(loc).Add(24 * time.Hour).Format(time.RFC3339)

	bodyObject := map[string]interface{}{
		"merchantId":         cfg.MerchantID,
		"storeId":            cfg.StoreID,
		"terminalId":         cfg.TerminalID,
		"partnerReferenceNo": partnerRefNo,
		"amount":             map[string]string{"value": fmt.Sprintf("%.2f", amount), "currency": "IDR"},
		"validityPeriod":     validity,
		"additionalInfo": map[string]string{
			"callbackUrl": cfg.PaymentCallbackURL,
			"type":        "dinamis",
		},
	}

	body, _ := json.Marshal(bodyObject)
	body = minifyJSON(body)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] CreateQRIS request URL=%s body=%s", url, string(body))
	log.Printf("[Pakailink] CreateQRIS request headers: X-TIMESTAMP=%s, X-PARTNER-ID=%s, CHANNEL-ID=%s", timestamp, cfg.PartnerID, cfg.ChannelID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] CreateQRIS response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkCreateQRResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2004700" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

type PakailinkVAStatusResponse struct {
	ResponseCode               string `json:"responseCode"`
	ResponseMessage            string `json:"responseMessage"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string `json:"latestTransactionStatus"`
	TransactionStatusDesc      string `json:"transactionStatusDesc"`
}

func InquiryPakailinkVAStatus(ctx context.Context, client *http.Client, accessToken, partnerRefNo string) (*PakailinkVAStatusResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}

	path := "/snap/v1.0/transfer-va/create-va-status"
	url := cfg.BaseURL + path
	bodyObject := map[string]string{"originalPartnerReferenceNo": partnerRefNo}
	body, _ := json.Marshal(bodyObject)
	body = minifyJSON(body)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] InquiryVA request URL=%s body=%s", url, string(body))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] InquiryVA response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkVAStatusResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2003300" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

type PakailinkQRStatusResponse struct {
	ResponseCode               string `json:"responseCode"`
	ResponseMessage            string `json:"responseMessage"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string `json:"latestTransactionStatus"`
	TransactionStatusDesc      string `json:"transactionStatusDesc"`
}

func InquiryPakailinkQRStatus(ctx context.Context, client *http.Client, accessToken, partnerRefNo string) (*PakailinkQRStatusResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}

	path := "/snap/v1.0/qr/qr-mpm-status"
	url := cfg.BaseURL + path
	bodyObject := map[string]string{"originalPartnerReferenceNo": partnerRefNo}
	body, _ := json.Marshal(bodyObject)
	body = minifyJSON(body)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] InquiryQR request URL=%s body=%s", url, string(body))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] InquiryQR response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkQRStatusResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2005300" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

func IsPakailinkSuccessStatus(status string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(status))
	return normalized == "00" || normalized == "SUCCESS" || normalized == "SUCCESSFUL" || normalized == "SETTLEMENT"
}

func IsPakailinkPendingStatus(status string) bool {
	switch strings.TrimSpace(strings.ToUpper(status)) {
	case "01", "02", "03", "PENDING", "INIT", "INITIATED", "INISIASI", "PAYING", "PROCESSING", "PROCESS", "CREATED", "DIMULAI", "MEMBAYAR":
		return true
	default:
		return false
	}
}

func IsPakailinkFailedStatus(status string) bool {
	switch strings.TrimSpace(strings.ToUpper(status)) {
	case "04", "05", "06", "07", "08", "FAILED", "FAIL", "CANCEL", "CANCELED", "CANCELLED", "EXPIRED", "CLOSED", "DIBATALKAN", "GAGAL":
		return true
	default:
		return false
	}
}

func IsPakailinkSuccessState(status, description string) bool {
	return IsPakailinkSuccessStatus(status) || IsPakailinkSuccessStatus(description)
}

func IsPakailinkPendingState(status, description string) bool {
	return IsPakailinkPendingStatus(status) || IsPakailinkPendingStatus(description)
}

func IsPakailinkFailedState(status, description string) bool {
	return IsPakailinkFailedStatus(status) || IsPakailinkFailedStatus(description)
}

type PakailinkBankInquiryResponse struct {
	ResponseCode           string `json:"responseCode"`
	ResponseMessage        string `json:"responseMessage"`
	SessionID              string `json:"sessionId"`
	PartnerReferenceNo     string `json:"partnerReferenceNo"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNumber"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName"`
	BeneficiaryBankName    string `json:"beneficiaryBankName"`
}

func PakailinkBankInquiry(ctx context.Context, client *http.Client, accessToken, partnerRefNo, accountNumber, bankCode string) (*PakailinkBankInquiryResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}

	path := "/snap/v1.0/emoney/bank-account-inquiry"
	url := cfg.BaseURL + path

	type inquiryAdditionalInfo struct {
		BeneficiaryBankCode string `json:"beneficiaryBankCode"`
	}
	type inquiryPayload struct {
		PartnerReferenceNo   string                `json:"partnerReferenceNo"`
		BeneficiaryAccountNo string                `json:"beneficiaryAccountNumber"`
		AdditionalInfo       inquiryAdditionalInfo `json:"additionalInfo"`
	}

	bodyObject := inquiryPayload{
		PartnerReferenceNo:   partnerRefNo,
		BeneficiaryAccountNo: accountNumber,
		AdditionalInfo:       inquiryAdditionalInfo{BeneficiaryBankCode: bankCode},
	}
	body, _ := json.Marshal(bodyObject)
	// Do not use minifyJSON as it sorts map keys alphabetically which breaks strict gateway parsers
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] BankInquiry request URL=%s", url)
	for k, v := range req.Header {
		log.Printf("[Pakailink] BankInquiry request header: %s=%s", k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] BankInquiry request body=%s", string(body))
	log.Printf("[Pakailink] BankInquiry response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkBankInquiryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2004200" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

type PakailinkBankTransferResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	ReferenceNo        string `json:"referenceNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	AdditionalInfo     struct {
		TransactionStatus string `json:"transactionStatus"`
	} `json:"additionalInfo"`
}

func PakailinkBankTransfer(ctx context.Context, client *http.Client, accessToken, partnerRefNo, accountNumber, bankCode, sessionID string, amount float64, callbackURL string) (*PakailinkBankTransferResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}

	path := "/snap/v1.0/emoney/transfer-bank"
	url := cfg.BaseURL + path
	type transferAmount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	}
	type transferAdditionalInfo struct {
		CallbackUrl string `json:"callbackUrl,omitempty"`
		Remark      string `json:"remark"`
	}
	type transferPayload struct {
		PartnerReferenceNo       string                 `json:"partnerReferenceNo"`
		BeneficiaryAccountNumber string                 `json:"beneficiaryAccountNumber"`
		BeneficiaryBankCode      string                 `json:"beneficiaryBankCode"`
		SessionID                string                 `json:"sessionId"`
		Amount                   transferAmount         `json:"amount"`
		AdditionalInfo           transferAdditionalInfo `json:"additionalInfo"`
	}

	bodyObject := transferPayload{
		PartnerReferenceNo:       partnerRefNo,
		BeneficiaryAccountNumber: accountNumber,
		BeneficiaryBankCode:      bankCode,
		SessionID:                sessionID,
		Amount: transferAmount{
			Value:    fmt.Sprintf("%.2f", amount),
			Currency: "IDR",
		},
		AdditionalInfo: transferAdditionalInfo{
			CallbackUrl: strings.TrimSpace(callbackURL),
			Remark:      "",
		},
	}

	body, _ := json.Marshal(bodyObject)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] BankTransfer request URL=%s body=%s", url, string(body))
	log.Printf("[Pakailink] BankTransfer request headers: X-TIMESTAMP=%s, X-PARTNER-ID=%s", timestamp, cfg.PartnerID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] BankTransfer response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkBankTransferResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2004300" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

type PakailinkEwalletInquiryResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	SessionID          string `json:"sessionId"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	CustomerNumber     string `json:"customerNumber"`
	CustomerName       string `json:"customerName"`
}

func PakailinkEwalletInquiry(ctx context.Context, client *http.Client, accessToken, partnerRefNo, customerNumber, productCode string) (*PakailinkEwalletInquiryResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}

	path := "/snap/v1.0/emoney/account-inquiry"
	url := cfg.BaseURL + path
	bodyObject := map[string]interface{}{
		"partnerReferenceNo": partnerRefNo,
		"customerNumber":     customerNumber,
		"additionalInfo":     map[string]string{"productCode": productCode},
	}
	body, _ := json.Marshal(bodyObject)
	body = minifyJSON(body)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] EwalletInquiry request body=%s", string(body))
	log.Printf("[Pakailink] EwalletInquiry response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkEwalletInquiryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2003700" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

type PakailinkEwalletTopupResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	ReferenceNo        string `json:"referenceNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	AdditionalInfo     struct {
		TransactionStatus string `json:"transactionStatus"`
	} `json:"additionalInfo"`
}

func PakailinkEwalletTopup(ctx context.Context, client *http.Client, accessToken, partnerRefNo, customerNumber, productCode, sessionID string, amount float64, callbackURL string) (*PakailinkEwalletTopupResponse, error) {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return nil, err
	}

	path := "/snap/v1.0/emoney/topup"
	url := cfg.BaseURL + path
	additionalInfo := map[string]interface{}{}
	if strings.TrimSpace(callbackURL) != "" {
		additionalInfo["callbackUrl"] = callbackURL
	}
	bodyObject := map[string]interface{}{
		"partnerReferenceNo": partnerRefNo,
		"customerNumber":     customerNumber,
		"productCode":        productCode,
		"amount":             map[string]string{"value": fmt.Sprintf("%.2f", amount), "currency": "IDR"},
		"additionalInfo":     additionalInfo,
	}
	if strings.TrimSpace(sessionID) != "" {
		bodyObject["sessionId"] = sessionID
	}

	body, _ := json.Marshal(bodyObject)
	body = minifyJSON(body)
	timestamp := PakailinkTimestamp()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", generateExternalID())
	req.Header.Set("CHANNEL-ID", cfg.ChannelID)
	req.Header.Set("X-SIGNATURE", createSymmetricSignature(http.MethodPost, path, accessToken, body, timestamp, cfg.ClientSecret))

	log.Printf("[Pakailink] EwalletTopup request URL=%s body=%s", url, string(body))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Pakailink] EwalletTopup response status=%d body=%s", resp.StatusCode, string(respBody))
	var result PakailinkEwalletTopupResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ResponseCode != "2003800" {
		return nil, fmt.Errorf("%s: %s", result.ResponseCode, result.ResponseMessage)
	}

	return &result, nil
}

func GetPakailinkPaymentCallbackURL() string {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return ""
	}
	return cfg.PaymentCallbackURL
}

func GetPakailinkPayoutCallbackURL() string {
	cfg, err := getPakailinkConfig()
	if err != nil {
		return ""
	}
	if strings.TrimSpace(cfg.PayoutCallbackURL) != "" {
		return cfg.PayoutCallbackURL
	}
	return cfg.PaymentCallbackURL
}

func VerifyPakailinkCallbackSignature(r *http.Request, body []byte) error {
	signature := strings.TrimSpace(r.Header.Get("X-SIGNATURE"))
	timestamp := strings.TrimSpace(r.Header.Get("X-TIMESTAMP"))

	log.Printf("[Pakailink Callback] Headers received - X-SIGNATURE: %s, X-TIMESTAMP: %s",
		maskString(signature), timestamp)

	publicKeyPath := strings.TrimSpace(os.Getenv("PAKAILINK_CALLBACK_PUBLIC_KEY_PATH"))
	if publicKeyPath == "" {
		// No public key configured - accept callback without signature verification
		// This is useful for development or when PakaiLink public key is not available
		log.Printf("[Pakailink Callback] No PAKAILINK_CALLBACK_PUBLIC_KEY_PATH configured - skipping signature verification")
		return nil
	}

	if signature == "" {
		log.Printf("[Pakailink Callback] Missing X-SIGNATURE header")
		return fmt.Errorf("header X-SIGNATURE wajib ada")
	}
	if timestamp == "" {
		log.Printf("[Pakailink Callback] Missing X-TIMESTAMP header")
		return fmt.Errorf("header X-TIMESTAMP wajib ada")
	}

	keyData, err := os.ReadFile(filepath.Clean(publicKeyPath))
	if err != nil {
		log.Printf("[Pakailink Callback] Failed to read public key: %v", err)
		return fmt.Errorf("gagal membaca public key callback: %w", err)
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		log.Printf("[Pakailink Callback] Invalid public key format")
		return fmt.Errorf("format public key callback tidak valid")
	}

	publicKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		certificate, certErr := x509.ParseCertificate(block.Bytes)
		if certErr == nil {
			publicKeyAny = certificate.PublicKey
		} else {
			log.Printf("[Pakailink Callback] Failed to parse public key: %v", err)
			return fmt.Errorf("gagal parse public key callback: %w", err)
		}
	}

	publicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		log.Printf("[Pakailink Callback] Public key is not RSA")
		return fmt.Errorf("public key callback bukan RSA")
	}

	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Printf("[Pakailink Callback] Failed to decode signature: %v", err)
		return fmt.Errorf("signature callback tidak valid: %w", err)
	}

	bodyHash := sha256.Sum256(minifyJSON(body))
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))
	method := strings.ToUpper(strings.TrimSpace(r.Method))

	host := strings.TrimSpace(r.Host)
	forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if forwardedProto == "" {
		forwardedProto = "https"
	}

	candidates := []string{
		method + ":" + r.URL.Path + ":" + bodyHashHex + ":" + timestamp,
		method + ":" + r.URL.RequestURI() + ":" + bodyHashHex + ":" + timestamp,
		method + ":" + forwardedProto + "://" + host + r.URL.Path + ":" + bodyHashHex + ":" + timestamp,
		method + ":" + forwardedProto + "://" + host + r.URL.RequestURI() + ":" + bodyHashHex + ":" + timestamp,
	}

	signatureValid := false
	for i, candidate := range candidates {
		sum := sha256.Sum256([]byte(candidate))
		if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sum[:], decodedSignature); err == nil {
			log.Printf("[Pakailink Callback] Signature verified successfully with candidate #%d", i+1)
			signatureValid = true
			break
		}
	}

	if !signatureValid {
		log.Printf("[Pakailink Callback] Signature verification failed - all candidates rejected")
		return fmt.Errorf("signature callback PakaiLink tidak valid")
	}

	return nil
}

// maskString masks a string for safe logging (shows first and last 4 chars)
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
