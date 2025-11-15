-- Migration: Create tutorials table
-- Date: 2025-01-XX

-- Create tutorials table
CREATE TABLE IF NOT EXISTS `tutorials` (
  `id` int UNSIGNED NOT NULL,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `image` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `link` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add indexes
ALTER TABLE `tutorials`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_status` (`status`);

-- Add AUTO_INCREMENT
ALTER TABLE `tutorials`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

