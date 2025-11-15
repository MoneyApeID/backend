package admins

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"project/database"
	"project/models"
	"project/utils"
)

// POST /api/admin/tutorials
func CreateTutorialHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid form data"})
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	link := strings.TrimSpace(r.FormValue("link"))
	status := strings.TrimSpace(r.FormValue("status"))

	if title == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Title wajib diisi"})
		return
	}

	if link == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Link wajib diisi"})
		return
	}

	if status != "Active" && status != "Inactive" {
		status = "Active"
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gambar diperlukan"})
		return
	}
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

	// Read first 512 bytes to detect MIME type
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

	// Read rest of file
	rest, err := io.ReadAll(file)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Gagal membaca gambar"})
		return
	}
	imageBytes := append(buf[:n], rest...)

	// Decode and re-encode to sanitize
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

	// Generate UUID for filename
	uuidStr := uuid.New().String()
	objectName := uuidStr + ext

	// Upload to S3
	reader := bytes.NewReader(outBuf.Bytes())
	_, err = utils.UploadToS3Server(objectName, reader, int64(outBuf.Len()))
	if err != nil {
		log.Printf("Error uploading image to S3 (objectName: %s): %v", objectName, err)
		// Return error message untuk membantu debugging
		// Di production, bisa disamarkan jika diperlukan
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal upload gambar. Pastikan S3_BUCKET_SERVER, S3_ACCESS_KEY, dan S3_SECRET_KEY sudah di-set dengan benar. Error: " + err.Error(),
		})
		return
	}

	// Create tutorial - simpan hanya filename, bukan URL
	tutorial := models.Tutorial{
		Title:  title,
		Image:  objectName, // Simpan hanya filename: uuid.ext
		Link:   link,
		Status: status,
	}

	db := database.DB
	if err := db.Create(&tutorial).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal membuat tutorial"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Tutorial berhasil dibuat",
		Data:    tutorial,
	})
}

// GET /api/admin/tutorials
func ListTutorialsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	var tutorials []models.Tutorial
	if err := db.Order("id DESC").Find(&tutorials).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengambil data tutorial"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    map[string]interface{}{"tutorials": tutorials},
	})
}

// PUT /api/admin/tutorials
func UpdateTutorialHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid form data"})
		return
	}

	idStr := strings.TrimSpace(r.FormValue("id"))
	if idStr == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID wajib diisi"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID tidak valid"})
		return
	}

	db := database.DB
	var tutorial models.Tutorial
	if err := db.First(&tutorial, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Tutorial tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Update fields
	title := strings.TrimSpace(r.FormValue("title"))
	link := strings.TrimSpace(r.FormValue("link"))
	status := strings.TrimSpace(r.FormValue("status"))

	if title != "" {
		tutorial.Title = title
	}
	if link != "" {
		tutorial.Link = link
	}
	if status == "Active" || status == "Inactive" {
		tutorial.Status = status
	}

	// Handle image update if provided (optional)
	// Image is optional: if not provided, keep the existing image (tutorial.Image remains unchanged)
	file, handler, err := r.FormFile("image")
	if err != nil {
		// File not provided or error getting file - this is OK, image is optional
		// Skip image update, keep existing image in database
		// Note: http.ErrMissingFile indicates file was not sent, which is expected for optional fields
		if err != http.ErrMissingFile {
			// Log non-standard errors for debugging, but don't fail the request
			log.Printf("Warning: Error getting image file (non-missing file error): %v", err)
		}
		// Continue without updating image - tutorial.Image will remain as-is
	} else if file != nil && handler != nil && handler.Filename != "" {
		// Image file is provided - process and update image
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

		// Delete old image from S3 (Image adalah filename langsung)
		oldImageFilename := tutorial.Image
		if oldImageFilename != "" {
			_ = utils.DeleteFromS3Server(oldImageFilename) // Ignore error if file doesn't exist
		}

		// Upload new image
		uuidStr := uuid.New().String()
		objectName := uuidStr + ext
		reader := bytes.NewReader(outBuf.Bytes())
		_, err = utils.UploadToS3Server(objectName, reader, int64(outBuf.Len()))
		if err != nil {
			log.Printf("Error uploading image to S3 (objectName: %s): %v", objectName, err)
			// Return error message untuk membantu debugging
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Gagal upload gambar. Pastikan S3_BUCKET_SERVER, S3_ACCESS_KEY, dan S3_SECRET_KEY sudah di-set dengan benar. Error: " + err.Error(),
			})
			return
		}

		// Simpan hanya filename, bukan URL
		tutorial.Image = objectName
	}

	// Save updates
	if err := db.Save(&tutorial).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal mengupdate tutorial"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Tutorial berhasil diupdate",
		Data:    tutorial,
	})
}

// DELETE /api/admin/tutorials/{id}
func DeleteTutorialHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var idStr string
	if len(parts) >= 4 {
		idStr = parts[3]
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "ID tidak valid"})
		return
	}

	db := database.DB
	var tutorial models.Tutorial
	if err := db.First(&tutorial, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "Tutorial tidak ditemukan"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Terjadi kesalahan"})
		return
	}

	// Delete image from S3 (Image adalah filename langsung)
	if tutorial.Image != "" {
		_ = utils.DeleteFromS3Server(tutorial.Image) // Ignore error if file doesn't exist
	}

	// Delete from database
	if err := db.Delete(&tutorial).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Gagal menghapus tutorial"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Tutorial berhasil dihapus",
	})
}

