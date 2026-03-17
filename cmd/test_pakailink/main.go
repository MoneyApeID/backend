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
	fmt.Printf("Timestamp: %s\n", timestamp)
	fmt.Printf("Signature: %s\n", signature)

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

	tests := []struct {
		name          string
		partnerRefNo  string
		externalID    string
		accountNumber string
		bankCode      string
	}{
		{"Test1: Alpha refNo + Numeric extID", randomAlpha(28) + "1234", fmt.Sprintf("%d", time.Now().UnixMilli()), "53602400026652", "116"},
		{"Test2: All numeric", fmt.Sprintf("%d", time.Now().UnixMilli()), fmt.Sprintf("%d", time.Now().UnixMilli()+1), "53602400026652", "116"},
		{"Test3: Doc example BCA", "ilFpX51e0CAttU2DW7dDWV7TCWqk1cE1wyJj", fmt.Sprintf("%d", time.Now().UnixMilli()), "6750620416", "014"},
		{"Test4: Alpha 32", randomAlpha(32), fmt.Sprintf("%d", time.Now().UnixMilli()), "53602400026652", "116"},
	}

	for _, tc := range tests {
		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("PartnerRefNo: %s\n", tc.partnerRefNo)
		fmt.Printf("ExternalID: %s\n", tc.externalID)
		fmt.Printf("AccountNumber: %s\n", tc.accountNumber)
		fmt.Printf("BankCode: %s\n", tc.bankCode)
		result := testBankInquiry(accessToken, tc.partnerRefNo, tc.externalID, tc.accountNumber, tc.bankCode)
		fmt.Printf("Result: %s\n", result)
		time.Sleep(500 * time.Millisecond)
	}
}

func testBankInquiry(accessToken, partnerRefNo, externalID, accountNumber, bankCode string) string {
	timestamp := timeNow()
	bodyMap := map[string]interface{}{
		"partnerReferenceNo":       partnerRefNo,
		"beneficiaryAccountNumber": accountNumber,
		"additionalInfo":           map[string]string{"beneficiaryBankCode": bankCode},
	}
	body, _ := json.Marshal(bodyMap)
	path := "/snap/v1.0/emoney/bank-account-inquiry"
	sig := createSymmetricSignature("POST", path, accessToken, body, timestamp)
	fmt.Printf("Body: %s\n", string(body))
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
