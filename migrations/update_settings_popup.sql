-- Migration: Update settings table - change logo to popup and add popup_title, created_at, updated_at
-- Date: 2025-01-XX

-- Rename logo column to popup
ALTER TABLE `settings` 
  CHANGE COLUMN `logo` `popup` text DEFAULT NULL;

-- Add popup_title column (will fail if already exists, but that's OK)
ALTER TABLE `settings` 
  ADD COLUMN `popup_title` varchar(255) DEFAULT NULL AFTER `popup`;

-- Add created_at column (will fail if already exists, but that's OK)
ALTER TABLE `settings` 
  ADD COLUMN `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `link_app`;

-- Add updated_at column (will fail if already exists, but that's OK)
ALTER TABLE `settings` 
  ADD COLUMN `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER `created_at`;

