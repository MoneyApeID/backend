package main

import (
	"bytes"
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
	mathrand "math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var baseURL = "https://api.pakailink.id"
var clientKey = "ee7f8fc2564211f0a993fa163e117483"
var clientSecret = "921988da7032bc8683c795dba81e4e84"
var partnerID = "PTR00000TI"
var channelID = "95222"

func main() {
	privKeyPath := "./keys/rsa_private_key.pem"

	fmt.Println("=== Getting Access Token ===")
	timestamp := timeNow()
	stringToSign := clientKey + "|" + timestamp

	privKeyData, err := os.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal("Failed to read private key:", err)
	}

	signature := createAsymmetricSignature(stringToSign, privKeyData)
	body := []byte(`{"grantType":"client_credentials"}`)
	req, _ := http.NewRequest("POST", baseURL+"/snap/v1.0/access-token/b2b", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-CLIENT-KEY", clientKey)
	req.Header.Set("X-SIGNATURE", signature)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Token request error:", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Token Response: %s\n\n", string(respBody))

	var tokenResp struct {
		AccessToken     string `json:"accessToken"`
		ResponseCode    string `json:"responseCode"`
		ResponseMessage string `json:"responseMessage"`
	}
	json.Unmarshal(respBody, &tokenResp)

	if tokenResp.AccessToken == "" {
		log.Fatal("Failed to get access token:", tokenResp.ResponseCode, tokenResp.ResponseMessage)
	}

	accessToken := tokenResp.AccessToken
	fmt.Printf("Access Token: %s...\n\n", accessToken[:50])

	// Test Bank Inquiry with various scenarios
	bankTests := []struct {
		name       string
		bodyFunc   func(externalID string) ([]byte, string)
		externalID string
	}{
		{"Test1: UUID v4 for both", func(extID string) ([]byte, string) {
			refNo := uuid.New().String()
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": "53602400026652",
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "116"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, uuid.New().String()},
		{"Test2: UUID without dashes", func(extID string) ([]byte, string) {
			refNo := strings.ReplaceAll(uuid.New().String(), "-", "")
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": "53602400026652",
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "116"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, strings.ReplaceAll(uuid.New().String(), "-", "")},
		{"Test3: Numeric strings", func(extID string) ([]byte, string) {
			refNo := fmt.Sprintf("%d", time.Now().UnixMilli())
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": "53602400026652",
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "116"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, fmt.Sprintf("%d", time.Now().UnixMilli())},
		{"Test4: Alpha only", func(extID string) ([]byte, string) {
			refNo := randomAlpha(32)
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": "53602400026652",
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "116"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, randomAlpha(10)},
		{"Test5: With space in JSON key (like docs)", func(extID string) ([]byte, string) {
			refNo := randomAlpha(32)
			body := `{"partnerReferenceNo": "` + refNo + `", "beneficiaryAccountNumber" : "53602400026652", "additionalInfo": {"beneficiaryBankCode": "116"}}`
			return []byte(body), refNo
		}, fmt.Sprintf("%d", time.Now().UnixMilli())},
		{"Test6: Different bank code 014", func(extID string) ([]byte, string) {
			refNo := randomAlpha(32)
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": "53602400026652",
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "014"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, fmt.Sprintf("%d", time.Now().UnixMilli())},
		{"Test7: Account as int number", func(extID string) ([]byte, string) {
			refNo := randomAlpha(32)
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": 53602400026652,
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "116"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, fmt.Sprintf("%d", time.Now().UnixMilli())},
		{"Test8: Minified body exact order", func(extID string) ([]byte, string) {
			refNo := "ilFpX51e0CAttU2DW7dDWV7TCWqk1cE1wyJj"
			body := `{"partnerReferenceNo":"` + refNo + `","beneficiaryAccountNumber":"53602400026652","additionalInfo":{"beneficiaryBankCode":"116"}}`
			return []byte(body), refNo
		}, "1155348175"},
		{"Test9: Different field order", func(extID string) ([]byte, string) {
			refNo := randomAlpha(32)
			body := `{"additionalInfo":{"beneficiaryBankCode":"116"},"beneficiaryAccountNumber":"53602400026652","partnerReferenceNo":"` + refNo + `"}`
			return []byte(body), refNo
		}, fmt.Sprintf("%d", time.Now().UnixMilli())},
		{"Test10: Short refNo", func(extID string) ([]byte, string) {
			refNo := "test123"
			body := map[string]interface{}{
				"partnerReferenceNo":       refNo,
				"beneficiaryAccountNumber": "53602400026652",
				"additionalInfo":           map[string]string{"beneficiaryBankCode": "116"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, "12345"},
	}

	fmt.Println("\n========== BANK INQUIRY TESTS ==========")
	for _, tc := range bankTests {
		bodyBytes, refNo := tc.bodyFunc(tc.externalID)
		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("PartnerRefNo: %s\n", refNo)
		fmt.Printf("ExternalID: %s\n", tc.externalID)
		fmt.Printf("Body: %s\n", string(bodyBytes))
		result := testRequest(accessToken, "/snap/v1.0/emoney/bank-account-inquiry", bodyBytes, tc.externalID)
		fmt.Printf("Result: %s\n", result)
		time.Sleep(300 * time.Millisecond)
	}

	// Test Ewallet Inquiry
	fmt.Println("\n\n========== EWALLET INQUIRY TESTS ==========")
	ewalletTests := []struct {
		name       string
		bodyFunc   func(externalID string) ([]byte, string)
		externalID string
	}{
		{"Ewallet1: UUID v4", func(extID string) ([]byte, string) {
			refNo := uuid.New().String()
			body := map[string]interface{}{
				"partnerReferenceNo": refNo,
				"customerNumber":     "085165667472",
				"additionalInfo":     map[string]string{"productCode": "DANA"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, uuid.New().String()},
		{"Ewallet2: Numeric", func(extID string) ([]byte, string) {
			refNo := fmt.Sprintf("%d", time.Now().UnixMilli())
			body := map[string]interface{}{
				"partnerReferenceNo": refNo,
				"customerNumber":     "085165667472",
				"additionalInfo":     map[string]string{"productCode": "DANA"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, fmt.Sprintf("%d", time.Now().UnixMilli())},
		{"Ewallet3: Alpha only", func(extID string) ([]byte, string) {
			refNo := randomAlpha(32)
			body := map[string]interface{}{
				"partnerReferenceNo": refNo,
				"customerNumber":     "085165667472",
				"additionalInfo":     map[string]string{"productCode": "DANA"},
			}
			b, _ := json.Marshal(body)
			return b, refNo
		}, randomAlpha(10)},
	}

	for _, tc := range ewalletTests {
		bodyBytes, refNo := tc.bodyFunc(tc.externalID)
		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("PartnerRefNo: %s\n", refNo)
		fmt.Printf("ExternalID: %s\n", tc.externalID)
		fmt.Printf("Body: %s\n", string(bodyBytes))
		result := testRequest(accessToken, "/snap/v1.0/emoney/account-inquiry", bodyBytes, tc.externalID)
		fmt.Printf("Result: %s\n", result)
		time.Sleep(300 * time.Millisecond)
	}
}

func testRequest(accessToken, path string, body []byte, externalID string) string {
	timestamp := timeNow()
	sig := createSymmetricSignature("POST", path, accessToken, body, timestamp)

	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-PARTNER-ID", partnerID)
	req.Header.Set("X-EXTERNAL-ID", externalID)
	req.Header.Set("CHANNEL-ID", channelID)
	req.Header.Set("X-SIGNATURE", sig)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return string(respBody)
}

func timeNow() string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc).Format("2006-01-02T15:04:05-07:00")
}

func randomAlpha(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func createSymmetricSignature(method, path, accessToken string, body []byte, timestamp string) string {
	bodyHash := sha256.Sum256(body)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))
	stringToSign := method + ":" + path + ":" + accessToken + ":" + bodyHashHex + ":" + timestamp
	mac := hmac.New(sha512.New, []byte(clientSecret))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func createAsymmetricSignature(stringToSign string, privateKeyData []byte) string {
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		log.Fatal("Failed to parse PEM block")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal("Failed to parse private key:", err)
		}
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		log.Fatal("Private key is not RSA")
	}
	hash := sha256.Sum256([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hash[:])
	if err != nil {
		log.Fatal("Failed to sign:", err)
	}
	return base64.StdEncoding.EncodeToString(signature)
}
