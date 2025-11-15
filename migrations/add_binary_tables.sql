-- Migration script untuk menambahkan tabel binary system
-- Hanya menambahkan tabel yang belum ada, tidak menghapus data yang sudah ada

-- Tabel binary_nodes
CREATE TABLE IF NOT EXISTS `binary_nodes` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` int UNSIGNED NOT NULL,
  `left_id` int UNSIGNED DEFAULT NULL COMMENT 'User ID di sisi kiri',
  `right_id` int UNSIGNED DEFAULT NULL COMMENT 'User ID di sisi kanan',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `left_id` (`left_id`),
  KEY `right_id` (`right_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Binary tree structure kiri-kanan';

-- Tabel rewards
CREATE TABLE IF NOT EXISTS `rewards` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `omset_target` decimal(15,2) NOT NULL COMMENT 'Target omset untuk mendapatkan reward',
  `reward_desc` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Deskripsi reward (untuk manual distribution)',
  `duration` int NOT NULL COMMENT 'Durasi dalam hari',
  `is_accumulative` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1 = akumulasi, 0 = reset',
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Definisi reward yang tersedia';

-- Tabel reward_progress
CREATE TABLE IF NOT EXISTS `reward_progress` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` int UNSIGNED NOT NULL,
  `reward_id` int UNSIGNED NOT NULL,
  `omset_left` decimal(15,2) NOT NULL DEFAULT '0.00' COMMENT 'Omset dari sisi kiri',
  `omset_right` decimal(15,2) NOT NULL DEFAULT '0.00' COMMENT 'Omset dari sisi kanan',
  `total_omset` decimal(15,2) NOT NULL DEFAULT '0.00' COMMENT 'Total omset (kiri + kanan)',
  `is_completed` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Apakah sudah mencapai target',
  `is_claimed` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Apakah sudah di-claim (manual)',
  `started_at` datetime NOT NULL COMMENT 'Kapan periode dimulai',
  `expires_at` datetime DEFAULT NULL COMMENT 'Kapan periode berakhir (untuk reset)',
  `last_reset_at` datetime DEFAULT NULL COMMENT 'Kapan terakhir di-reset',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_reward_id` (`reward_id`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `idx_user_reward` (`user_id`,`reward_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Progress reward untuk setiap user';

-- Seed data untuk rewards (hanya insert jika belum ada)
INSERT IGNORE INTO `rewards` (`name`, `omset_target`, `reward_desc`, `duration`, `is_accumulative`, `status`) VALUES
('Money Reward 1', 5000000.00, 'Rp500.000 (tunai)', 30, 0, 'Active'),
('Money Reward 2', 15000000.00, 'Ponsel/ Gadget (Rp1,5–2 juta)', 30, 0, 'Active'),
('Money Reward 3', 30000000.00, 'Kulkas kecil / TV (Rp2,5–3 juta)', 60, 0, 'Active'),
('Money Reward 4', 75000000.00, 'Uang tunai 10.000.000', 90, 1, 'Active'),
('Money Reward 5', 200000000.00, 'Rp20.000.000 (tunai)', 150, 1, 'Active');

