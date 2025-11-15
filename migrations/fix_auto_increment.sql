-- Script untuk menambahkan AUTO_INCREMENT pada semua tabel yang belum punya
-- Run dengan: docker exec -i vla-mysql mysql -u root -pvlaroot vla-db < migrations/fix_auto_increment.sql

-- Users
ALTER TABLE `users` MODIFY `id` int NOT NULL AUTO_INCREMENT;

-- Banks
ALTER TABLE `banks` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- Forums
ALTER TABLE `forums` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- Payments
ALTER TABLE `payments` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- Payment Settings
ALTER TABLE `payment_settings` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- Products
ALTER TABLE `products` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- Settings
ALTER TABLE `settings` MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

-- Spin Prizes
ALTER TABLE `spin_prizes` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- Tasks
ALTER TABLE `tasks` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- User Tasks
ALTER TABLE `user_tasks` MODIFY `id` int NOT NULL AUTO_INCREMENT;

-- Withdrawals
ALTER TABLE `withdrawals` MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

