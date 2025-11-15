package utils

import (
	"project/database"
	"project/models"
	"sort"

	"gorm.io/gorm"
)

// getAllDownlineMembers mengumpulkan semua downline dari binary tree sampai level tertentu
func getAllDownlineMembers(userID uint, maxLevel int) []models.User {
	db := database.DB
	var allMembers []models.User

	// Helper rekursif untuk mengumpulkan semua downline
	var collectDownline func(nodeID *uint, currentLevel int)
	collectDownline = func(nodeID *uint, currentLevel int) {
		if nodeID == nil || currentLevel > maxLevel {
			return
		}

		// Get user
		var user models.User
		if err := db.Select("id, name, number").Where("id = ?", *nodeID).First(&user).Error; err != nil {
			return
		}
		allMembers = append(allMembers, user)

		// Get binary node
		var binaryNode models.BinaryNode
		if err := db.Where("user_id = ?", *nodeID).First(&binaryNode).Error; err != nil {
			return
		}

		// Recurse ke level berikutnya
		if currentLevel < maxLevel {
			collectDownline(binaryNode.LeftID, currentLevel+1)
			collectDownline(binaryNode.RightID, currentLevel+1)
		}
	}

	// Start dari root's binary node
	var rootBinaryNode models.BinaryNode
	if err := db.Where("user_id = ?", userID).First(&rootBinaryNode).Error; err != nil {
		return allMembers
	}

	// Collect all downline
	collectDownline(rootBinaryNode.LeftID, 1)
	collectDownline(rootBinaryNode.RightID, 1)

	return allMembers
}

// GetTopMembersByOmset mengembalikan top N members berdasarkan omset
// Digunakan untuk binary structure yang menampilkan member dengan omset terbesar
// targetLevel: level yang ingin diambil (1, 2, atau 3)
// limit: jumlah member yang ingin diambil (2 untuk level 1, 4 untuk level 2, 8 untuk level 3)
func GetTopMembersByOmset(userID uint, targetLevel int, limit int) []BinaryMemberWithOmset {
	db := database.DB

	// Get all downline di level target
	var allMembers []models.User

	// Helper untuk mengumpulkan semua member di level tertentu
	var collectAtLevel func(leftID, rightID *uint, currentLevel int)
	collectAtLevel = func(leftID, rightID *uint, currentLevel int) {
		if currentLevel > targetLevel {
			return
		}

		if currentLevel == targetLevel {
			// Ambil member di level ini
			if leftID != nil {
				var user models.User
				if err := db.Select("id, name, number").Where("id = ?", *leftID).First(&user).Error; err == nil {
					allMembers = append(allMembers, user)
				}
			}
			if rightID != nil {
				var user models.User
				if err := db.Select("id, name, number").Where("id = ?", *rightID).First(&user).Error; err == nil {
					allMembers = append(allMembers, user)
				}
			}
			return
		}

		// Recurse ke level berikutnya
		if leftID != nil {
			var leftNode models.BinaryNode
			if err := db.Where("user_id = ?", *leftID).First(&leftNode).Error; err == nil {
				collectAtLevel(leftNode.LeftID, leftNode.RightID, currentLevel+1)
			}
		}
		if rightID != nil {
			var rightNode models.BinaryNode
			if err := db.Where("user_id = ?", *rightID).First(&rightNode).Error; err == nil {
				collectAtLevel(rightNode.LeftID, rightNode.RightID, currentLevel+1)
			}
		}
	}

	// Start dari root's binary node
	var rootBinaryNode models.BinaryNode
	if err := db.Where("user_id = ?", userID).First(&rootBinaryNode).Error; err != nil {
		return []BinaryMemberWithOmset{}
	}

	// Collect all members at target level
	collectAtLevel(rootBinaryNode.LeftID, rootBinaryNode.RightID, 1)

	// Calculate omset untuk setiap member dan sort
	type MemberWithOmset struct {
		User  models.User
		Omset float64
	}
	membersWithOmset := make([]MemberWithOmset, 0, len(allMembers))
	for _, member := range allMembers {
		omset, _ := CalculateUserOmsetOnly(member.ID)
		membersWithOmset = append(membersWithOmset, MemberWithOmset{
			User:  member,
			Omset: omset,
		})
	}

	// Sort by omset descending, jika omset sama, sort by user_id (pilih salah satu)
	sort.Slice(membersWithOmset, func(i, j int) bool {
		if membersWithOmset[i].Omset == membersWithOmset[j].Omset {
			// Jika omset sama, pilih berdasarkan user_id (lebih kecil)
			return membersWithOmset[i].User.ID < membersWithOmset[j].User.ID
		}
		return membersWithOmset[i].Omset > membersWithOmset[j].Omset
	})

	// Take top N
	result := make([]BinaryMemberWithOmset, 0, limit)
	for i := 0; i < len(membersWithOmset) && i < limit; i++ {
		m := membersWithOmset[i]
		// Determine position based on binary tree structure (left or right)
		position := determinePosition(userID, m.User.ID, db)
		result = append(result, BinaryMemberWithOmset{
			UserID:   m.User.ID,
			Name:     m.User.Name,
			Number:   m.User.Number,
			Omset:    m.Omset,
			Position: position,
		})
	}

	return result
}

// BinaryMemberWithOmset untuk response member dengan omset
type BinaryMemberWithOmset struct {
	UserID   uint    `json:"user_id"`
	Name     string  `json:"name"`
	Number   string  `json:"number"`
	Omset    float64 `json:"omset"`
	Position string  `json:"position"` // "left" atau "right"
}

// determinePosition menentukan posisi member (left/right) dalam binary tree
func determinePosition(rootID, memberID uint, db *gorm.DB) string {
	// Cek apakah member ada di sisi kiri atau kanan dari root
	var rootBinaryNode models.BinaryNode
	if err := db.Where("user_id = ?", rootID).First(&rootBinaryNode).Error; err != nil {
		return "left" // default
	}

	// Check left side
	if rootBinaryNode.LeftID != nil && *rootBinaryNode.LeftID == memberID {
		return "left"
	}

	// Check recursively in left subtree
	if rootBinaryNode.LeftID != nil {
		if isInSubtree(*rootBinaryNode.LeftID, memberID, db) {
			return "left"
		}
	}

	// Check right side
	if rootBinaryNode.RightID != nil && *rootBinaryNode.RightID == memberID {
		return "right"
	}

	// Check recursively in right subtree
	if rootBinaryNode.RightID != nil {
		if isInSubtree(*rootBinaryNode.RightID, memberID, db) {
			return "right"
		}
	}

	return "left" // default
}

// isInSubtree mengecek apakah memberID ada di subtree dari nodeID
func isInSubtree(nodeID, memberID uint, db *gorm.DB) bool {
	if nodeID == memberID {
		return true
	}

	var binaryNode models.BinaryNode
	if err := db.Where("user_id = ?", nodeID).First(&binaryNode).Error; err != nil {
		return false
	}

	if binaryNode.LeftID != nil {
		if *binaryNode.LeftID == memberID || isInSubtree(*binaryNode.LeftID, memberID, db) {
			return true
		}
	}

	if binaryNode.RightID != nil {
		if *binaryNode.RightID == memberID || isInSubtree(*binaryNode.RightID, memberID, db) {
			return true
		}
	}

	return false
}

