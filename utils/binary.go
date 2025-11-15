package utils

import (
	"project/database"
	"project/models"

	"gorm.io/gorm"
)

// AssignBinaryNode mengassign user baru ke binary tree (kiri atau kanan)
// Logic: Cari posisi kosong terdekat dari upline, mulai dari kiri dulu
func AssignBinaryNode(uplineID uint, newUserID uint) error {
	db := database.DB

	// Cek apakah upline sudah punya binary node
	var binaryNode models.BinaryNode
	err := db.Where("user_id = ?", uplineID).First(&binaryNode).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		// Upline belum punya binary node, buat baru dengan newUser di kiri
		binaryNode = models.BinaryNode{
			UserID:  uplineID,
			LeftID:  &newUserID,
			RightID: nil,
		}
		return db.Create(&binaryNode).Error
	} else if err != nil {
		return err
	}

	// Upline sudah punya binary node
	// Cek apakah kiri kosong, jika ya assign ke kiri
	if binaryNode.LeftID == nil {
		binaryNode.LeftID = &newUserID
		return db.Save(&binaryNode).Error
	}

	// Kiri sudah terisi, cek kanan
	if binaryNode.RightID == nil {
		binaryNode.RightID = &newUserID
		return db.Save(&binaryNode).Error
	}

	// Kedua sisi sudah terisi, cari posisi kosong di level berikutnya
	// Gunakan algoritma: cari posisi kosong terdekat dengan prioritas kiri dulu
	return findAndAssignPosition(binaryNode.LeftID, binaryNode.RightID, newUserID, db)
}

// findAndAssignPosition mencari posisi kosong di level berikutnya
// Prioritas: kiri dulu, baru kanan
func findAndAssignPosition(leftID, rightID *uint, newUserID uint, db *gorm.DB) error {
	// Coba kiri dulu
	if leftID != nil {
		var leftNode models.BinaryNode
		if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
			if leftNode.LeftID == nil {
				leftNode.LeftID = &newUserID
				return db.Save(&leftNode).Error
			}
			if leftNode.RightID == nil {
				leftNode.RightID = &newUserID
				return db.Save(&leftNode).Error
			}
			// Kedua sisi terisi, rekursif ke level berikutnya
			if err := findAndAssignPosition(leftNode.LeftID, leftNode.RightID, newUserID, db); err == nil {
				return nil
			}
		}
	}

	// Jika kiri tidak bisa, coba kanan
	if rightID != nil {
		var rightNode models.BinaryNode
		if err := db.Where("user_id = ?", *rightID).First(&rightNode).Error; err == nil {
			if rightNode.LeftID == nil {
				rightNode.LeftID = &newUserID
				return db.Save(&rightNode).Error
			}
			if rightNode.RightID == nil {
				rightNode.RightID = &newUserID
				return db.Save(&rightNode).Error
			}
			// Kedua sisi terisi, rekursif ke level berikutnya
			return findAndAssignPosition(rightNode.LeftID, rightNode.RightID, newUserID, db)
		}
	}

	// Fallback: jika semua penuh, assign ke kiri dari leftID (spillover)
	if leftID != nil {
		var leftNode models.BinaryNode
		if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
			return findAndAssignPosition(leftNode.LeftID, leftNode.RightID, newUserID, db)
		}
	}

	// Jika masih tidak bisa, buat binary node untuk newUser (sebagai root baru jika diperlukan)
	// Tapi seharusnya tidak sampai sini jika struktur binary sudah benar
	return nil
}

// CalculateOmset menghitung omset dari level 1-3 untuk user tertentu
// Omset = total investasi dari semua downline di level 1-3 (kiri + kanan)
func CalculateOmset(userID uint) (omsetLeft, omsetRight, totalOmset float64, err error) {
	db := database.DB

	var binaryNode models.BinaryNode
	if err := db.Where("user_id = ?", userID).First(&binaryNode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// User belum punya binary node, omset = 0
			return 0, 0, 0, nil
		}
		return 0, 0, 0, err
	}

	// Hitung omset dari sisi kiri (level 1-3)
	omsetLeft = calculateOmsetRecursive(binaryNode.LeftID, db, 1, 3)

	// Hitung omset dari sisi kanan (level 1-3)
	omsetRight = calculateOmsetRecursive(binaryNode.RightID, db, 1, 3)

	totalOmset = omsetLeft + omsetRight
	return omsetLeft, omsetRight, totalOmset, nil
}

// calculateOmsetRecursive menghitung omset secara rekursif sampai level tertentu
// Omset = total_returned (penghasilan dari return harian) dari investasi aktif (status Running) user ini + semua downline di level 1-3
func calculateOmsetRecursive(userID *uint, db *gorm.DB, currentLevel, maxLevel int) float64 {
	if userID == nil || currentLevel > maxLevel {
		return 0
	}

	// Hitung omset dari total_returned (penghasilan yang sudah dibayar) dari investasi aktif user ini (status Running)
	var totalOmset float64
	var investments []models.Investment
	if err := db.Select("total_returned").Where("user_id = ? AND status = ?", *userID, "Running").Find(&investments).Error; err == nil {
		for _, inv := range investments {
			totalOmset += inv.TotalReturned
		}
	}

	// Ambil binary node user ini
	var binaryNode models.BinaryNode
	if err := db.Where("user_id = ?", *userID).First(&binaryNode).Error; err != nil {
		// User belum punya binary node, hanya return omset dari investasinya sendiri
		return totalOmset
	}

	// Rekursif ke level berikutnya (kiri dan kanan)
	if currentLevel < maxLevel {
		totalOmset += calculateOmsetRecursive(binaryNode.LeftID, db, currentLevel+1, maxLevel)
		totalOmset += calculateOmsetRecursive(binaryNode.RightID, db, currentLevel+1, maxLevel)
	}

	return totalOmset
}

// CalculateUserOmsetOnly menghitung omset dari user tersebut saja (tanpa downline)
// Omset = total total_returned dari semua investasi aktif (status Running) user tersebut
func CalculateUserOmsetOnly(userID uint) (float64, error) {
	db := database.DB

	var totalOmset float64
	var investments []models.Investment
	if err := db.Select("total_returned").Where("user_id = ? AND status = ?", userID, "Running").Find(&investments).Error; err != nil {
		return 0, err
	}

	for _, inv := range investments {
		totalOmset += inv.TotalReturned
	}

	return totalOmset, nil
}

