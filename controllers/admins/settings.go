package admins

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"project/database"
	"project/models"
	"project/utils"
)

// GET /api/admin/settings
func GetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	var setting models.Setting
	if err := db.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Transform to response format
	response := map[string]interface{}{
		"name":            setting.Name,
		"company":         setting.Company,
		"popup":           setting.Popup,
		"popup_title":     setting.PopupTitle,
		"min_withdraw":    setting.MinWithdraw,
		"max_withdraw":    setting.MaxWithdraw,
		"withdraw_charge": setting.WithdrawCharge,
		"auto_withdraw":   setting.AutoWithdraw,
		"maintenance":     setting.Maintenance,
		"closed_register": setting.ClosedRegister,
		"link_cs":         setting.LinkCS,
		"link_group":      setting.LinkGroup,
		"link_app":        setting.LinkApp,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

// PUT /api/admin/settings
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid form data",
		})
		return
	}

	db := database.DB

	// Get current settings
	var setting models.Setting
	if err := db.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Update text fields
	if name := strings.TrimSpace(r.FormValue("name")); name != "" {
		setting.Name = name
	}
	if company := strings.TrimSpace(r.FormValue("company")); company != "" {
		setting.Company = company
	}
	if popupTitle := strings.TrimSpace(r.FormValue("popup_title")); popupTitle != "" {
		setting.PopupTitle = popupTitle
	}
	if minWithdrawStr := strings.TrimSpace(r.FormValue("min_withdraw")); minWithdrawStr != "" {
		if minWithdraw, err := strconv.ParseFloat(minWithdrawStr, 64); err == nil {
			setting.MinWithdraw = minWithdraw
		}
	}
	if maxWithdrawStr := strings.TrimSpace(r.FormValue("max_withdraw")); maxWithdrawStr != "" {
		if maxWithdraw, err := strconv.ParseFloat(maxWithdrawStr, 64); err == nil {
			setting.MaxWithdraw = maxWithdraw
		}
	}
	if withdrawChargeStr := strings.TrimSpace(r.FormValue("withdraw_charge")); withdrawChargeStr != "" {
		if withdrawCharge, err := strconv.ParseFloat(withdrawChargeStr, 64); err == nil {
			setting.WithdrawCharge = withdrawCharge
		}
	}
	if autoWithdrawStr := strings.TrimSpace(r.FormValue("auto_withdraw")); autoWithdrawStr != "" {
		setting.AutoWithdraw = autoWithdrawStr == "true" || autoWithdrawStr == "1"
	}
	if maintenanceStr := strings.TrimSpace(r.FormValue("maintenance")); maintenanceStr != "" {
		setting.Maintenance = maintenanceStr == "true" || maintenanceStr == "1"
	}
	if closedRegisterStr := strings.TrimSpace(r.FormValue("closed_register")); closedRegisterStr != "" {
		setting.ClosedRegister = closedRegisterStr == "true" || closedRegisterStr == "1"
	}
	if linkCS := strings.TrimSpace(r.FormValue("link_cs")); linkCS != "" {
		setting.LinkCS = linkCS
	}
	if linkGroup := strings.TrimSpace(r.FormValue("link_group")); linkGroup != "" {
		setting.LinkGroup = linkGroup
	}
	if linkApp := strings.TrimSpace(r.FormValue("link_app")); linkApp != "" {
		setting.LinkApp = linkApp
	}

	// Handle popup image upload (optional)
	file, handler, err := r.FormFile("popup")
	if err == nil && file != nil && handler != nil && handler.Filename != "" {
		// Image is provided, process it
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gambar harus JPG/PNG"})
			return
		}

		if handler.Size > 10<<20 {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gambar maksimal 10MB"})
			return
		}

		// Read and validate image
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gagal membaca gambar"})
			return
		}
		detected := http.DetectContentType(buf[:n])
		if detected != "image/jpeg" && detected != "image/png" {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gambar harus JPG/PNG"})
			return
		}

		rest, err := io.ReadAll(file)
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gagal membaca gambar"})
			return
		}
		imageBytes := append(buf[:n], rest...)

		imgReader := bytes.NewReader(imageBytes)
		img, format, err := image.Decode(imgReader)
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid image format"})
			return
		}

		var outBuf bytes.Buffer
		switch format {
		case "jpeg":
			if err := jpeg.Encode(&outBuf, img, &jpeg.Options{Quality: 85}); err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memproses gambar"})
				return
			}
			ext = ".jpg"
		case "png":
			if err := png.Encode(&outBuf, img); err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal memproses gambar"})
				return
			}
			ext = ".png"
		default:
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gambar harus JPG/PNG"})
			return
		}

		// Delete old popup image from S3 if exists
		oldPopupFilename := setting.Popup
		if oldPopupFilename != "" {
			_ = utils.DeleteFromS3Server(oldPopupFilename) // Ignore error if file doesn't exist
		}

		// Upload new popup image with fixed name "popup.{ext}"
		objectName := "popup" + ext
		reader := bytes.NewReader(outBuf.Bytes())
		_, err = utils.UploadToS3Server(objectName, reader, int64(outBuf.Len()))
		if err != nil {
			log.Printf("Error uploading popup image to S3: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal upload gambar. Pastikan S3_BUCKET_SERVER, S3_ACCESS_KEY, dan S3_SECRET_KEY sudah di-set dengan benar. Error: " + err.Error(),
			})
			return
		}

		// Save filename only
		setting.Popup = objectName
	}

	if err := db.Save(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Transform to response format
	response := map[string]interface{}{
		"name":            setting.Name,
		"company":         setting.Company,
		"popup":           setting.Popup,
		"popup_title":     setting.PopupTitle,
		"min_withdraw":    setting.MinWithdraw,
		"max_withdraw":    setting.MaxWithdraw,
		"withdraw_charge": setting.WithdrawCharge,
		"auto_withdraw":   setting.AutoWithdraw,
		"maintenance":     setting.Maintenance,
		"closed_register": setting.ClosedRegister,
		"link_cs":         setting.LinkCS,
		"link_group":      setting.LinkGroup,
		"link_app":        setting.LinkApp,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Pengaturan berhasil diperbarui",
		Data:    response,
	})
}
