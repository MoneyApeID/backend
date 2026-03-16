-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Waktu pembuatan: 05 Okt 2025 pada 01.34
-- Versi server: 8.4.3
-- Versi PHP: 8.3.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `sf`
--

-- --------------------------------------------------------

--
-- Struktur dari tabel `admins`
--

CREATE TABLE `admins` (
  `id` bigint UNSIGNED NOT NULL,
  `username` varchar(191) NOT NULL,
  `password` longtext NOT NULL,
  `name` longtext NOT NULL,
  `email` varchar(191) DEFAULT NULL,
  `role` varchar(191) DEFAULT 'admin',
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Dumping data untuk tabel `admins`
--

INSERT INTO `admins` (`id`, `username`, `password`, `name`, `email`, `role`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'admin', '$2y$10$I4qWolurBpmNKJlQUqb6CeBASh/8Sv59gWu6Ys.m9UsXPLdRLm0du', 'Admin', 'admin@vladevs.com', 'admin', 1, '2000-01-01 00:00:00.000', '2000-01-01 00:00:00.000');

INSERT INTO `admins` (`id`, `username`, `password`, `name`, `email`, `role`, `is_active`, `created_at`, `updated_at`) VALUES
(2, 'admin2', '$2y$10$I4qWolurBpmNKJlQUqb6CeBASh/8Sv59gWu6Ys.m9UsXPLdRLm0du', 'Admin2', 'admin2@vladevs.com', 'admin', 1, '2000-01-01 00:00:00.000', '2000-01-01 00:00:00.000');

-- --------------------------------------------------------

--
-- Struktur dari tabel `banks`
--

CREATE TABLE `banks` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Bank Rakyat Indonesia, Bank Central Asia, Dana, GoPay',
  `short_name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'BCA, BRI - untuk search/display',
  `type` enum('bank','ewallet') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'bank' COMMENT 'bank atau ewallet',
  `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '014 untuk BCA, DANA untuk ewallet - kode gateway/Pakailink',
  `status` enum('Active','Maintenance','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Available banks and e-wallets for withdrawal';

--
-- Dumping data untuk tabel `banks`
--

INSERT INTO `banks` (`id`, `name`, `short_name`, `type`, `code`, `status`) VALUES
(1, 'Dana', 'DANA', 'ewallet', 'DANA', 'Active'),
(2, 'GoPay', 'GOPAY', 'ewallet', 'GOPAY', 'Active'),
(3, 'OVO', 'OVO', 'ewallet', 'OVO', 'Active'),
(4, 'ShopeePay', 'SHOPEEPAY', 'ewallet', 'SHOPEEPAY', 'Active'),
(5, 'LinkAja', 'LINKAJA', 'ewallet', 'LINKAJA', 'Active'),
(6, 'Bank BRI', 'BRI', 'bank', '002', 'Active'),
(7, 'Bank Mandiri', 'MANDIRI', 'bank', '008', 'Active'),
(8, 'Bank BNI', 'BNI', 'bank', '009', 'Active'),
(9, 'Bank Danamon Indonesia', 'DANAMON', 'bank', '011', 'Active'),
(10, 'Bank Permata', 'PERMATA', 'bank', '013', 'Active'),
(11, 'Bank BCA', 'BCA', 'bank', '014', 'Active'),
(12, 'Bank Maybank Indonesia', 'MAYBANK', 'bank', '016', 'Active'),
(13, 'Bank Panin', 'PANIN', 'bank', '019', 'Active'),
(14, 'Bank CIMB Niaga', 'CIMB', 'bank', '022', 'Active'),
(15, 'Bank UOB Indonesia', 'UOB', 'bank', '023', 'Active'),
(16, 'Bank OCBC Indonesia', 'OCBC', 'bank', '028', 'Active'),
(17, 'Citibank, N.A', 'CITIBANK', 'bank', '031', 'Active'),
(18, 'JP. Morgan Chase Bank, N.A', 'JPM', 'bank', '032', 'Active'),
(19, 'Bank of America, N.A', 'BOA', 'bank', '033', 'Active'),
(20, 'China Construction Bank Indonesia', 'CCB', 'bank', '036', 'Active'),
(21, 'Bank Artha Graha Internasional', 'AGI', 'bank', '037', 'Active'),
(22, 'Bangkok Bank', 'BANGKOK', 'bank', '040', 'Active'),
(23, 'MUFG Bank, Ltd.', 'MUFG', 'bank', '042', 'Active'),
(24, 'Bank DBS Indonesia', 'DBS', 'bank', '046', 'Active'),
(25, 'Bank Resona Perdania', 'BRP', 'bank', '047', 'Active'),
(26, 'Bank Mizuho Indonesia', 'MIZUHO', 'bank', '048', 'Active'),
(27, 'Standard Chartered Bank', 'CHARTERED', 'bank', '050', 'Active'),
(28, 'Bank Capital Indonesia', 'CAPITAL', 'bank', '054', 'Active'),
(29, 'Bank BNP Paribas Indonesia', 'BNP', 'bank', '057', 'Active'),
(30, 'Bank ANZ Indonesia', 'ANZ', 'bank', '061', 'Active'),
(31, 'Deutsche Bank AG', 'DEUTSCHE', 'bank', '067', 'Active'),
(32, 'Bank of China', 'BOC', 'bank', '069', 'Active'),
(33, 'Bank Bumi Arta', 'ARTA', 'bank', '076', 'Active'),
(34, 'Bank HSBC Indonesia', 'HSBC', 'bank', '087', 'Active'),
(35, 'Bank J Trust Indonesia', 'JTRUST', 'bank', '095', 'Active'),
(36, 'Bank Mayapada', 'MAYAPADA', 'bank', '097', 'Active'),
(37, 'Bank BJB', 'BJB', 'bank', '110', 'Active'),
(38, 'Bank DKI', 'DKI', 'bank', '111', 'Active'),
(39, 'Bank BPD DIY', 'DIY', 'bank', '112', 'Active'),
(40, 'Bank Jateng', 'JATENG', 'bank', '113', 'Active'),
(41, 'Bank Jatim', 'JATIM', 'bank', '114', 'Active'),
(42, 'Bank Jambi', 'JAMBI', 'bank', '115', 'Active'),
(43, 'Bank Aceh Syariah', 'ACEHSYARIAH', 'bank', '116', 'Active'),
(44, 'Bank Sumut', 'SUMUT', 'bank', '117', 'Active'),
(45, 'Bank Nagari', 'NAGARI', 'bank', '118', 'Active'),
(46, 'Bank Riau Kepri Syariah', 'RIAUSYARIAH', 'bank', '119', 'Active'),
(47, 'Bank Sumsel Babel', 'SUMSEL', 'bank', '120', 'Active'),
(48, 'Bank Lampung', 'LAMPUNG', 'bank', '121', 'Active'),
(49, 'Bank Kalsel', 'KALSEL', 'bank', '122', 'Active'),
(50, 'Bank Kalbar', 'KALBAR', 'bank', '123', 'Active'),
(51, 'Bank Kaltimtara', 'KALTIM', 'bank', '124', 'Active'),
(52, 'Bank Kalteng', 'KALTENG', 'bank', '125', 'Active'),
(53, 'Bank Sulselbar', 'SULSELBAR', 'bank', '126', 'Active'),
(54, 'Bank SulutGo', 'SULUTGO', 'bank', '127', 'Active'),
(55, 'Bank NTB Syariah', 'NTBSYARIAH', 'bank', '128', 'Active'),
(56, 'Bank BPD Bali', 'BALI', 'bank', '129', 'Active'),
(57, 'Bank NTT', 'NTT', 'bank', '130', 'Active'),
(58, 'Bank Maluku Malut', 'MALUKU', 'bank', '131', 'Active'),
(59, 'Bank Papua', 'PAPUA', 'bank', '132', 'Active'),
(60, 'Bank Bengkulu', 'BENGKULU', 'bank', '133', 'Active'),
(61, 'Bank Sulteng', 'SULTENG', 'bank', '134', 'Active'),
(62, 'Bank Sultra', 'SULTRA', 'bank', '135', 'Active'),
(63, 'Bank Banten', 'BANTEN', 'bank', '137', 'Active'),
(64, 'Bank of India Indonesia', 'BOI', 'bank', '146', 'Active'),
(65, 'Bank Muamalat Indonesia', 'MUAMALAT', 'bank', '147', 'Active'),
(66, 'Bank Mestika Dharma', 'MESTIKA', 'bank', '151', 'Active'),
(67, 'Bank Shinhan Indonesia', 'SHINHAN', 'bank', '152', 'Active'),
(68, 'Bank Sinarmas', 'SINARMAS', 'bank', '153', 'Active'),
(69, 'Bank Maspion Indonesia', 'MASPION', 'bank', '157', 'Active'),
(70, 'Bank Ganesha', 'GANESHA', 'bank', '161', 'Active'),
(71, 'Bank ICBC Indonesia', 'ICBC', 'bank', '164', 'Active'),
(72, 'Bank QNB Indonesia', 'QNB', 'bank', '167', 'Active'),
(73, 'Bank BTN', 'BTN', 'bank', '200', 'Active'),
(74, 'Bank Woori Saudara', 'WOORI', 'bank', '212', 'Active'),
(75, 'Bank SMBC Indonesia (Jenius)', 'JENIUS', 'bank', '213', 'Active'),
(76, 'Bank BJB Syariah', 'BJBSYARIAH', 'bank', '425', 'Active'),
(77, 'Bank Mega', 'MEGA', 'bank', '426', 'Active'),
(78, 'Bank KB Bukopin', 'BUKOPIN', 'bank', '441', 'Active'),
(79, 'Bank Syariah Indonesia (BSI)', 'BSI', 'bank', '451', 'Active'),
(80, 'Krom Bank Indonesia', 'KROM', 'bank', '459', 'Active'),
(81, 'Bank Jasa Jakarta', 'BJJ', 'bank', '472', 'Active'),
(82, 'Bank Hana (Line Bank)', 'LINE', 'bank', '484', 'Active'),
(83, 'Bank MNC Internasional', 'MNC', 'bank', '485', 'Active'),
(84, 'Bank Neo Commerce', 'NEO', 'bank', '490', 'Active'),
(85, 'Bank Raya Indonesia', 'RAYA', 'bank', '494', 'Active'),
(86, 'Bank SBI Indonesia', 'SBI', 'bank', '498', 'Active'),
(87, 'Bank Digital BCA (blu)', 'BLU', 'bank', '501', 'Active'),
(88, 'Bank National Nobu', 'NOBU', 'bank', '503', 'Active'),
(89, 'Bank Mega Syariah', 'MEGASYARIAH', 'bank', '506', 'Active'),
(90, 'Bank Ina Perdana', 'INA', 'bank', '513', 'Active'),
(91, 'Bank Panin Dubai Syariah', 'PANINSYARIAH', 'bank', '517', 'Active'),
(92, 'Prima Master Bank', 'PRIMA', 'bank', '520', 'Active'),
(93, 'Bank KB Bukopin Syariah', 'BUKOPINSYARIAH', 'bank', '521', 'Active'),
(94, 'Bank Sahabat Sampoerna', 'SAMPOERNA', 'bank', '523', 'Active'),
(95, 'Bank Oke Indonesia', 'OKE', 'bank', '526', 'Active'),
(96, 'Bank Amar Indonesia', 'AMAR', 'bank', '531', 'Active'),
(97, 'SeaBank Indonesia', 'SEABANK', 'bank', '535', 'Active'),
(98, 'Bank BCA Syariah', 'BCASYARIAH', 'bank', '536', 'Active'),
(99, 'Bank Jago', 'JAGO', 'bank', '542', 'Active'),
(100, 'Bank BTPN Syariah', 'BTPNSYARIAH', 'bank', '547', 'Active'),
(101, 'Bank Multi Arta Sentosa', 'MAS', 'bank', '548', 'Active'),
(102, 'Bank Hibank Indonesia', 'HIBANK', 'bank', '553', 'Active'),
(103, 'Bank Index Selindo', 'INDEX SELINDO', 'bank', '555', 'Active'),
(104, 'Super Bank Indonesia', 'SUPERBANK', 'bank', '562', 'Active'),
(105, 'Bank Mandiri Taspen', 'MANDIRITASPEN', 'bank', '564', 'Active'),
(106, 'Bank Victoria International', 'VICTORIA', 'bank', '566', 'Active'),
(107, 'Allo Bank Indonesia', 'ALLO', 'bank', '567', 'Active'),
(108, 'Bank IBK Indonesia', 'IBK', 'bank', '945', 'Active'),
(109, 'Bank Aladin Syariah', 'ALADIN', 'bank', '947', 'Active'),
(110, 'Bank CTBC Indonesia', 'CTBC', 'bank', '949', 'Active'),
(111, 'Bank Commonwealth', 'COMMONWEALTH', 'bank', '950', 'Active');

-- --------------------------------------------------------

--
-- Struktur dari tabel `bank_accounts`
--

CREATE TABLE `bank_accounts` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `bank_id` int UNSIGNED NOT NULL,
  `account_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Nama penerima/pemilik rekening',
  `account_number` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Nomor rekening atau nomor e-wallet'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User linked bank accounts and e-wallets';

-- --------------------------------------------------------

--
-- Struktur dari tabel `categories`
--

CREATE TABLE `categories` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `profit_type` enum('locked','unlocked') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'unlocked' COMMENT 'locked=paid at completion, unlocked=paid daily',
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Product categories';

--
-- Dumping data untuk tabel `categories`
--

INSERT INTO `categories` (`id`, `name`, `description`, `profit_type`, `status`, `created_at`, `updated_at`) VALUES
(1, 'Neura', 'Profit terkunci, dibayarkan saat investasi selesai', 'locked', 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(2, 'Finora', 'Profit langsung dibayarkan', 'unlocked', 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(3, 'Corex', 'Profit langsung dibayarkan', 'unlocked', 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00');

-- --------------------------------------------------------

--
-- Struktur dari tabel `forums`
--

CREATE TABLE `forums` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  `reward` decimal(15,2) DEFAULT '0.00',
  `description` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `image` varchar(255) NOT NULL,
  `status` enum('Accepted','Pending','Rejected') DEFAULT 'Pending',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `investments`
--

CREATE TABLE `investments` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `product_id` int UNSIGNED NOT NULL,
  `category_id` int UNSIGNED NOT NULL COMMENT 'Reference to categories table for profit handling',
  `amount` decimal(15,2) NOT NULL,
  `daily_profit` decimal(15,2) NOT NULL,
  `duration` int NOT NULL,
  `total_paid` int NOT NULL DEFAULT '0' COMMENT 'Number of days paid',
  `total_returned` decimal(15,2) NOT NULL DEFAULT '0.00' COMMENT 'Total profit accumulated (not paid for locked categories until completion)',
  `last_return_at` datetime DEFAULT NULL,
  `next_return_at` datetime DEFAULT NULL,
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('Pending','Running','Completed','Suspended','Cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `deposits`
--

CREATE TABLE `deposits` (
  `id` bigint UNSIGNED NOT NULL,
  `user_id` int UNSIGNED NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `payment_method` enum('QRIS','BANK') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `payment_channel` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `payment_code` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `status` enum('Success','Pending','Failed','Expired') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `expired_at` datetime NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `payments`
--

CREATE TABLE `payments` (
  `id` bigint UNSIGNED NOT NULL,
  `investment_id` int NOT NULL,
  `reference_id` varchar(191) DEFAULT NULL,
  `order_id` varchar(191) NOT NULL,
  `payment_method` varchar(16) DEFAULT NULL,
  `payment_channel` varchar(16) DEFAULT NULL,
  `payment_code` text,
  `payment_link` text,
  `status` varchar(16) NOT NULL DEFAULT 'Pending',
  `expired_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `payment_settings`
--

CREATE TABLE `payment_settings` (
  `id` bigint UNSIGNED NOT NULL,
  `pakasir_api_key` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `pakasir_project` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `deposit_amount` decimal(15,2) NOT NULL DEFAULT '0.00',
  `bank_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `bank_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `account_number` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `account_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `withdraw_amount` decimal(15,2) NOT NULL DEFAULT '0.00',
  `wishlist_id` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `payment_settings`
--

INSERT INTO `payment_settings` (`id`, `pakasir_api_key`, `pakasir_project`, `deposit_amount`, `bank_name`, `bank_code`, `account_number`, `account_name`, `withdraw_amount`, `wishlist_id`, `created_at`, `updated_at`) VALUES
(1, 'AWD1A2AWD132', 'AWD1SAD2A1W', 10000.00, 'Bank BCA', '014', '1234567890', 'Money Rich Admin', 50000.00, '1', '2025-09-26 12:13:38', '2025-09-26 12:13:38');

-- --------------------------------------------------------

--
-- Struktur dari tabel `products`
--

CREATE TABLE `products` (
  `id` int UNSIGNED NOT NULL,
  `category_id` int UNSIGNED NOT NULL COMMENT 'Reference to categories table',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `amount` decimal(15,2) NOT NULL COMMENT 'Fixed investment amount',
  `daily_profit` decimal(15,2) NOT NULL COMMENT 'Fixed daily profit amount',
  `duration` int NOT NULL COMMENT 'Duration in days',
  `required_vip` int DEFAULT '0' COMMENT 'Required VIP level (0 means no requirement)',
  `purchase_limit` int DEFAULT '0' COMMENT 'Maximum purchases per user (0 = unlimited)',
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `products`
--

INSERT INTO `products` (`id`, `category_id`, `name`, `amount`, `daily_profit`, `duration`, `required_vip`, `purchase_limit`, `status`, `created_at`, `updated_at`) VALUES
-- Neura Category (category_id=1, Locked Profit, No Purchase Limit)
(1, 1, 'Neura 1', 50000.00, 15000.00, 70, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(2, 1, 'Neura 2', 200000.00, 68000.00, 60, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(3, 1, 'Neura 3', 500000.00, 175000.00, 65, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(4, 1, 'Neura 4', 1250000.00, 432000.00, 65, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(5, 1, 'Neura 5', 2800000.00, 1050000.00, 65, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(6, 1, 'Neura 6', 7000000.00, 2660000.00, 50, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(7, 1, 'Neura 7', 20000000.00, 8000000.00, 50, 0, 0, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
-- Finora Category (category_id=2, Unlocked Profit, Limited to 1x per product)
(8, 2, 'Finora 1', 50000.00, 20000.00, 1, 1, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(9, 2, 'Finora 2', 250000.00, 275000.00, 1, 2, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(10, 2, 'Finora 3', 700000.00, 950000.00, 1, 3, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(11, 2, 'Finora 4', 2000000.00, 3600000.00, 1, 4, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(12, 2, 'Finora 5', 8000000.00, 16000000.00, 1, 5, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
-- Corex Category (category_id=3, All require VIP3, Limited purchases)
(13, 3, 'Corex 1', 80000.00, 70000.00, 1, 3, 2, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(14, 3, 'Corex 2', 165000.00, 150000.00, 1, 3, 2, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(15, 3, 'Corex 3', 750000.00, 1000000.00, 1, 3, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00'),
(16, 3, 'Corex 4', 2450000.00, 4000000.00, 1, 3, 1, 'Active', '2025-10-11 00:00:00', '2025-10-11 00:00:00');

-- --------------------------------------------------------

--
-- Struktur dari tabel `refresh_tokens`
--

CREATE TABLE `refresh_tokens` (
  `id` char(64) NOT NULL,
  `user_id` bigint NOT NULL,
  `expires_at` datetime(3) DEFAULT NULL,
  `revoked` tinyint(1) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `revoked_tokens`
--

CREATE TABLE `revoked_tokens` (
  `id` varchar(128) NOT NULL,
  `revoked_at` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `settings`
--

CREATE TABLE `settings` (
  `id` bigint UNSIGNED NOT NULL,
  `name` text NOT NULL,
  `company` text NOT NULL,
  `popup` text DEFAULT NULL,
  `popup_title` varchar(255) DEFAULT NULL,
  `min_withdraw` decimal(15,2) NOT NULL,
  `max_withdraw` decimal(15,2) NOT NULL,
  `withdraw_charge` decimal(15,2) NOT NULL,
  `maintenance` tinyint(1) NOT NULL DEFAULT '0',
  `closed_register` tinyint(1) NOT NULL DEFAULT '0',
  `auto_withdraw` tinyint(1) NOT NULL DEFAULT '0',
  `link_cs` text NOT NULL,
  `link_group` text NOT NULL,
  `link_app` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Dumping data untuk tabel `settings`
--

INSERT INTO `settings` (`id`, `name`, `company`, `popup`, `popup_title`, `min_withdraw`, `max_withdraw`, `withdraw_charge`, `maintenance`, `closed_register`, `auto_withdraw`, `link_cs`, `link_group`, `link_app`) VALUES
(1, 'Vla Devs', 'Vla Devs', NULL, NULL, 33000.00, 10000000.00, 10.00, 0, 0, 0, 'https://t.me/', 'https://t.me/', 'https://vladevs.com');

-- --------------------------------------------------------

--
-- Struktur dari tabel `spin_prizes`
--

CREATE TABLE `spin_prizes` (
  `id` int UNSIGNED NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Unique code untuk validasi claim prize',
  `chance_weight` int NOT NULL COMMENT 'Weight untuk random selection (semakin besar semakin sering muncul)',
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Available spin wheel prizes';

--
-- Dumping data untuk tabel `spin_prizes`
--

INSERT INTO `spin_prizes` (`id`, `amount`, `code`, `chance_weight`, `status`, `created_at`, `updated_at`) VALUES
(1, 1000.00, 'SPIN_1K', 5000, 'Active', '2025-08-31 02:48:48', '2025-09-18 12:18:21'),
(2, 5000.00, 'SPIN_5K', 500, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:11:12'),
(3, 10000.00, 'SPIN_10K', 300, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:11:16'),
(4, 50000.00, 'SPIN_50K', 30, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:17:32'),
(5, 100000.00, 'SPIN_100K', 10, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:17:28'),
(6, 200000.00, 'SPIN_200K', 5, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:04:43'),
(7, 500000.00, 'SPIN_500K', 2, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:04:46'),
(8, 1000000.00, 'SPIN_1000K', 1, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:50:03');

-- --------------------------------------------------------

--
-- Stand-in struktur untuk tampilan `spin_prizes_with_percentage`
-- (Lihat di bawah untuk tampilan aktual)
--
CREATE TABLE `spin_prizes_with_percentage` (
`amount` decimal(15,2)
,`chance_percentage` decimal(16,2)
,`chance_weight` int
,`code` varchar(20)
,`id` int unsigned
,`status` enum('Active','Inactive')
);

-- --------------------------------------------------------

--
-- Struktur dari tabel `tasks`
--

CREATE TABLE `tasks` (
  `id` int NOT NULL,
  `name` varchar(100) NOT NULL,
  `reward` decimal(15,2) NOT NULL,
  `required_level` int NOT NULL,
  `required_active_members` bigint NOT NULL,
  `status` enum('Active','Inactive') DEFAULT 'Active',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Dumping data untuk tabel `tasks`
--

INSERT INTO `tasks` (`id`, `name`, `reward`, `required_level`, `required_active_members`, `status`, `created_at`, `updated_at`) VALUES
(1, 'Tugas Perekrutan 1', 15000.00, 1, 5, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(2, 'Tugas Perekrutan 2', 35000.00, 1, 10, 'Active', '2025-09-08 03:57:01', '2025-09-11 22:07:23'),
(3, 'Tugas Perekrutan 3', 200000.00, 1, 50, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(4, 'Tugas Perekrutan 4', 450000.00, 1, 100, 'Active', '2025-09-08 03:57:01', '2025-09-08 03:57:01'),
(5, 'Tugas Perekrutan 5', 1000000.00, 1, 200, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(6, 'Tugas Perekrutan 6', 2750000.00, 1, 500, 'Active', '2025-09-08 03:57:01', '2025-09-08 03:57:01'),
(7, 'Tugas Perekrutan 7', 6000000.00, 1, 1000, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(8, 'Tugas Perekrutan 8', 16000000.00, 1, 2000, 'Active', '2025-09-08 03:57:01', '2025-09-08 04:00:03'),
(9, 'Tugas Perekrutan 9', 35000000.00, 1, 3000, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(10, 'Tugas Perekrutan 10', 80000000.00, 1, 5000, 'Active', '2025-09-08 03:57:01', '2025-09-08 03:57:01');

-- --------------------------------------------------------

--
-- Struktur dari tabel `transactions`
--

CREATE TABLE `transactions` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `charge` decimal(15,2) NOT NULL DEFAULT '0.00',
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `transaction_flow` enum('debit','credit') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'debit=money out, credit=money in',
  `transaction_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'deposit, withdraw, transfer, refund, bonus, penalty, etc',
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `status` enum('Success','Pending','Failed') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User transaction records';

-- --------------------------------------------------------

--
-- Struktur dari tabel `users`
--

CREATE TABLE `users` (
  `id` int NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `number` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `reff_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `reff_by` bigint UNSIGNED DEFAULT NULL,
  `balance` decimal(15,2) DEFAULT '0.00',
  `income` decimal(15,2) DEFAULT '0.00',
  `level` bigint NOT NULL DEFAULT '0' COMMENT 'VIP level (0-5)',
  `total_invest` decimal(15,2) DEFAULT '0.00' COMMENT 'Total all investments',
  `total_invest_vip` decimal(15,2) DEFAULT '0.00' COMMENT 'Total locked category investments for VIP level calculation',
  `spin_ticket` bigint DEFAULT '0',
  `status` enum('Active','Inactive','Suspend') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'Active',
  `investment_status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'Inactive',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `users`
--

INSERT INTO `users` (`id`, `name`, `number`, `password`, `reff_code`, `reff_by`, `balance`, `income`,  `level`, `total_invest`, `total_invest_vip`, `spin_ticket`, `status`, `investment_status`, `created_at`, `updated_at`) VALUES
(1, 'Ciroos Users Management', '812345678', '$2y$10$fa5X/6ZfpaNZsa07TyzO3ukL/AtxtGLv.6erFIw9KmXFNYyFbE656', 'CIROOS', 0, 0.00, 0.00, 0, 0.00, 0.00, 100, 'Active', 'Active', '2025-01-01 00:00:00.000', '2025-01-01 00:00:00.000');

-- --------------------------------------------------------

--
-- Struktur dari tabel `user_spins`
--

CREATE TABLE `user_spins` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `prize_id` int UNSIGNED NOT NULL COMMENT 'Reference to won prize',
  `amount` decimal(15,2) NOT NULL COMMENT 'Amount yang dimenangkan',
  `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Code hadiah yang dimenangkan',
  `won_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User spin wheel history and claims';

-- --------------------------------------------------------

--
-- Struktur dari tabel `tutorials`
--

CREATE TABLE `tutorials` (
  `id` int UNSIGNED NOT NULL,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `image` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `link` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `user_tasks`
--

CREATE TABLE `user_tasks` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  `task_id` int NOT NULL,
  `claimed_at` datetime DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `binary_nodes`
--

CREATE TABLE `binary_nodes` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int UNSIGNED NOT NULL,
  `left_id` int UNSIGNED DEFAULT NULL COMMENT 'User ID di sisi kiri',
  `right_id` int UNSIGNED DEFAULT NULL COMMENT 'User ID di sisi kanan',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Binary tree structure kiri-kanan';

-- --------------------------------------------------------

--
-- Struktur dari tabel `rewards`
--

CREATE TABLE `rewards` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `omset_target` decimal(15,2) NOT NULL COMMENT 'Target omset untuk mendapatkan reward',
  `reward_desc` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Deskripsi reward (untuk manual distribution)',
  `duration` int NOT NULL COMMENT 'Durasi dalam hari',
  `is_accumulative` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1 = akumulasi, 0 = reset',
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Definisi reward yang tersedia';

-- --------------------------------------------------------

--
-- Struktur dari tabel `reward_progress`
--

CREATE TABLE `reward_progress` (
  `id` int UNSIGNED NOT NULL,
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
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Progress reward untuk setiap user';

-- --------------------------------------------------------

--
-- Struktur dari tabel `withdrawals`
--

CREATE TABLE `withdrawals` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `bank_account_id` int UNSIGNED NOT NULL COMMENT 'Reference to user linked bank account',
  `amount` decimal(15,2) NOT NULL,
  `charge` decimal(15,2) NOT NULL DEFAULT '0.00',
  `final_amount` decimal(15,2) NOT NULL COMMENT 'amount - charge, calculated amount user receives',
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('Success','Pending','Failed') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User withdrawal requests';

--
-- Trigger `withdrawals`
--
DELIMITER $$
CREATE TRIGGER `withdrawals_calculate_final_amount` BEFORE INSERT ON `withdrawals` FOR EACH ROW BEGIN
    SET NEW.final_amount = NEW.amount - NEW.charge;
END
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `withdrawals_update_final_amount` BEFORE UPDATE ON `withdrawals` FOR EACH ROW BEGIN
    IF NEW.amount != OLD.amount OR NEW.charge != OLD.charge THEN
        SET NEW.final_amount = NEW.amount - NEW.charge;
    END IF;
END
$$
DELIMITER ;

--
-- Indexes for dumped tables
--

--
-- Indeks untuk tabel `admins`
--
ALTER TABLE `admins`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id` (`id`),
  ADD UNIQUE KEY `uni_admins_username` (`username`),
  ADD UNIQUE KEY `uni_admins_email` (`email`);

--
-- Indeks untuk tabel `banks`
--
ALTER TABLE `banks`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `code` (`code`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_code` (`code`);

--
-- Indeks untuk tabel `bank_accounts`
--
ALTER TABLE `bank_accounts`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_user_account` (`user_id`,`bank_id`,`account_number`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_bank_id` (`bank_id`);

--
-- Indeks untuk tabel `categories`
--
ALTER TABLE `categories`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_status` (`status`);

--
-- Indeks untuk tabel `forums`
--
ALTER TABLE `forums`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indeks untuk tabel `investments`
--
ALTER TABLE `investments`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_product_id` (`product_id`),
  ADD KEY `idx_category_id` (`category_id`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_next_return_at` (`next_return_at`);

--
-- Indeks untuk tabel `deposits`
--
ALTER TABLE `deposits`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`),
  ADD KEY `idx_deposits_user_id` (`user_id`),
  ADD KEY `idx_deposits_status` (`status`);

--
-- Indeks untuk tabel `payments`
--
ALTER TABLE `payments`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id` (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`);

--
-- Indeks untuk tabel `payment_settings`
--
ALTER TABLE `payment_settings`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_products_status` (`status`),
  ADD KEY `idx_products_category_id` (`category_id`),
  ADD KEY `idx_products_required_vip` (`required_vip`);

--
-- Indeks untuk tabel `refresh_tokens`
--
ALTER TABLE `refresh_tokens`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_refresh_user` (`user_id`),
  ADD KEY `idx_refresh_tokens_user_id` (`user_id`);

--
-- Indeks untuk tabel `revoked_tokens`
--
ALTER TABLE `revoked_tokens`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `settings`
--
ALTER TABLE `settings`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id` (`id`);

--
-- Indeks untuk tabel `spin_prizes`
--
ALTER TABLE `spin_prizes`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `code` (`code`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_code` (`code`),
  ADD KEY `idx_chance_weight` (`chance_weight`);

--
-- Indeks untuk tabel `spin_prizes_with_percentage`
--
ALTER TABLE `spin_prizes_with_percentage`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `tasks`
--
ALTER TABLE `tasks`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `transactions`
--
ALTER TABLE `transactions`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_order_id` (`order_id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_order_id` (`order_id`),
  ADD KEY `idx_transaction_flow` (`transaction_flow`),
  ADD KEY `idx_transaction_type` (`transaction_type`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_created_at` (`created_at`),
  ADD KEY `idx_user_status_created` (`user_id`,`status`,`created_at`),
  ADD KEY `idx_user_type_created` (`user_id`,`transaction_type`,`created_at`);

--
-- Indeks untuk tabel `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `idx_users_number` (`number`),
  ADD UNIQUE KEY `idx_users_reff_code` (`reff_code`),
  ADD KEY `idx_users_reff_by` (`reff_by`);

--
-- Indeks untuk tabel `user_spins`
--
ALTER TABLE `user_spins`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_won_at` (`won_at`),
  ADD KEY `fk_spins_prize` (`prize_id`);

--
-- Indeks untuk tabel `tutorials`
--
ALTER TABLE `tutorials`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_status` (`status`);

--
-- Indeks untuk tabel `user_tasks`
--
ALTER TABLE `user_tasks`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_user_task` (`user_id`,`task_id`),
  ADD KEY `task_id` (`task_id`);

--
-- Indeks untuk tabel `binary_nodes`
--
ALTER TABLE `binary_nodes`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `user_id` (`user_id`),
  ADD KEY `left_id` (`left_id`),
  ADD KEY `right_id` (`right_id`);

--
-- Indeks untuk tabel `rewards`
--
ALTER TABLE `rewards`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_status` (`status`);

--
-- Indeks untuk tabel `reward_progress`
--
ALTER TABLE `reward_progress`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_reward_id` (`reward_id`),
  ADD KEY `idx_expires_at` (`expires_at`),
  ADD KEY `idx_user_reward` (`user_id`,`reward_id`);

--
-- Indeks untuk tabel `withdrawals`
--
ALTER TABLE `withdrawals`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_bank_account_id` (`bank_account_id`),
  ADD KEY `idx_order_id` (`order_id`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_created_at` (`created_at`);

--
-- AUTO_INCREMENT untuk tabel yang dibuang
--

--
-- AUTO_INCREMENT untuk tabel `admins`
--
ALTER TABLE `admins`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `bank_accounts`
--
ALTER TABLE `bank_accounts`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `categories`
--
ALTER TABLE `categories`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `deposits`
--
ALTER TABLE `deposits`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `investments`
--
ALTER TABLE `investments`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `tutorials`
--
ALTER TABLE `tutorials`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `binary_nodes`
--
ALTER TABLE `binary_nodes`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `rewards`
--
ALTER TABLE `rewards`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `reward_progress`
--
ALTER TABLE `reward_progress`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `transactions`
--
ALTER TABLE `transactions`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `user_spins`
--
ALTER TABLE `user_spins`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `users`
--
ALTER TABLE `users`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `banks`
--
ALTER TABLE `banks`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `forums`
--
ALTER TABLE `forums`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `payments`
--
ALTER TABLE `payments`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `payment_settings`
--
ALTER TABLE `payment_settings`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `products`
--
ALTER TABLE `products`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `settings`
--
ALTER TABLE `settings`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `spin_prizes`
--
ALTER TABLE `spin_prizes`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `spin_prizes_with_percentage`
--
ALTER TABLE `spin_prizes_with_percentage`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `tasks`
--
ALTER TABLE `tasks`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `user_tasks`
--
ALTER TABLE `user_tasks`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `withdrawals`
--
ALTER TABLE `withdrawals`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;
