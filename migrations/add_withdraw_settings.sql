-- Migration: Add withdraw time and days to settings table
-- Date: 2026-03-17

ALTER TABLE settings ADD COLUMN withdraw_start_time varchar(10) NOT NULL DEFAULT '12:00';
ALTER TABLE settings ADD COLUMN withdraw_end_time varchar(10) NOT NULL DEFAULT '17:00';
ALTER TABLE settings ADD COLUMN withdraw_days varchar(255) NOT NULL DEFAULT '1,2,3,4,5,6';
