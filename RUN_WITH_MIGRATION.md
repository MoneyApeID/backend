# How to Run with Auto-Migration

## Masalah: required_vip dan purchase_limit selalu 0

### Root Cause:
AutoMigrate di `main.go` tidak mencakup model `Product`, `Category`, `User`, dan `Investment`, sehingga kolom-kolom baru tidak dibuat di database.

### Solusi: ✅ SUDAH DIPERBAIKI!

File `main.go` sudah diupdate untuk include semua models:
```go
db.AutoMigrate(
    &models.Admin{}, 
    &models.RefreshToken{}, 
    &models.User{},          // ✅ ADDED
    &models.Category{},      // ✅ ADDED  
    &models.Product{},       // ✅ ADDED
    &models.Investment{},    // ✅ ADDED
    &models.UserSpin{}, 
    &models.Setting{}, 
    &models.Payment{}, 
    &models.PaymentSettings{},
)
```

---

## Cara Menjalankan

### 1. Set Environment ke Development

**Windows PowerShell:**
```powershell
$env:ENV="development"
$env:DB_HOST="localhost"
$env:DB_PORT="3306"
$env:DB_USER="root"
$env:DB_PASS="your_password"
$env:DB_NAME="your_database_name"

go run main.go
```

**Linux/Mac:**
```bash
export ENV=development
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASS=your_password
export DB_NAME=your_database_name

go run main.go
```

### 2. Atau Gunakan .env File

Buat file `.env`:
```env
ENV=development
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=your_password
DB_NAME=ciroos_db

# LinkQu deposit configuration
LINKQU_BASE_URL=https://partner-api.linkqu.id
LINKQU_USERNAME=your_linkqu_username
LINKQU_PIN=your_linkqu_pin
LINKQU_CLIENT_ID=your_linkqu_client_id
LINKQU_CLIENT_SECRET=your_linkqu_client_secret
LINKQU_CALLBACK_PAYMENT=https://your-domain.com/api/payments/linkqu/callback

# JWT
JWT_SECRET=your_jwt_secret

# Cron
CRON_KEY=your_cron_key
```

Lalu jalankan dengan godotenv atau langsung:
```bash
go run main.go
```

---

## Apa yang Akan Terjadi

Saat `ENV=development`, aplikasi akan:

1. ✅ **Auto-create** kolom `required_vip` di table products
2. ✅ **Auto-create** kolom `purchase_limit` di table products
3. ✅ **Auto-create** kolom `total_invest_vip` di table users
4. ✅ **Auto-create** table `categories` jika belum ada
5. ✅ **Auto-update** schema untuk semua model

### Output di Console:
```
[database] using DSN: root:******@tcp(localhost:3306)/ciroos_db?...
Running in development mode - performing auto-migration
Auto-migration completed successfully ✅
Server is running on port 8080
```

---

## Verifikasi Setelah Run

### 1. Cek Log Console
Harus ada pesan: `Auto-migration completed successfully`

### 2. Test API Products
```bash
curl http://localhost:8080/api/products
```

Response harus include:
```json
{
  "data": {
    "Monitor": [{
      "required_vip": 0,      // ✅ Harus muncul!
      "purchase_limit": 0     // ✅ Harus muncul!
    }]
  }
}
```

### 3. Test Admin Update Product
```bash
curl -X PUT http://localhost:8080/api/admin/products/8 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"required_vip": 1, "purchase_limit": 1}'
```

Response harus show updated values.

---

## Production Deployment

### Option A: Manual Migration (Recommended for Production)

Jangan gunakan AutoMigrate di production. Jalankan migration manual:

```bash
# 1. Backup dulu
mysqldump -u root -p ciroos_db > backup_before_v2.sql

# 2. Run migration
mysql -u root -p ciroos_db < migrations/add_vip_and_purchase_limit.sql

# 3. Verify
mysql -u root -p ciroos_db -e "DESCRIBE products;"

# 4. Set ENV=production lalu run
export ENV=production
go run main.go
```

### Option B: Full DB Reset (Jika OK kehilangan data test)

```bash
# Drop & recreate database
mysql -u root -p -e "DROP DATABASE IF EXISTS ciroos_db;"
mysql -u root -p -e "CREATE DATABASE ciroos_db;"

# Import schema baru
mysql -u root -p ciroos_db < database/db.sql

# Run aplikasi
ENV=production go run main.go
```

---

## Troubleshooting

### Masalah: Kolom masih belum muncul setelah run

**Solusi:**
1. Pastikan ENV=development
2. Restart aplikasi (stop & start ulang)
3. Cek console log ada "Auto-migration completed"
4. Manual check database: `SHOW COLUMNS FROM products;`

### Masalah: Error saat AutoMigrate

**Solusi:**
1. Cek database credentials benar
2. Cek user database punya permission ALTER TABLE
3. Jalankan manual migration sebagai fallback

### Masalah: required_vip masih 0 di database

**Solusi:**
```sql
-- Update manual setelah kolom ada
UPDATE products SET required_vip = 1 WHERE id = 8;
UPDATE products SET required_vip = 2 WHERE id = 9;
UPDATE products SET required_vip = 3 WHERE id = 10;
UPDATE products SET required_vip = 4 WHERE id = 11;
UPDATE products SET required_vip = 5 WHERE id = 12;
UPDATE products SET required_vip = 3 WHERE id IN (13, 14, 15, 16);
```

### Masalah: Konfigurasi pembayaran

**Solusi:**
1. Pastikan variabel LINKQU_* sudah di-set di environment / `.env`
2. Callback URL harus mengarah ke `/api/payments/linkqu/callback`
3. Gunakan endpoint `/api/users/deposits` untuk membuat kode pembayaran baru

---

## Quick Start (Development)

```powershell
# Windows PowerShell
cd "C:\Users\USER\Website Client\CiroosAI\BackEnd-V3(Final)"
$env:ENV="development"
$env:DB_NAME="ciroos_db"
$env:DB_USER="root"
$env:DB_PASS="your_password"
go run main.go
```

Tunggu hingga muncul:
```
Auto-migration completed successfully ✅
Server is running on port 8080
```

Lalu test:
```
http://localhost:8080/api/products
```

---

## Summary

✅ **Fixed:** main.go AutoMigrate now includes all models
✅ **Created:** Migration script for manual deployment
✅ **Created:** This guide for running with migration

**Next Steps:**
1. Set `ENV=development`
2. Run `go run main.go`
3. Check console for "Auto-migration completed"
4. Test API response includes required_vip
5. Try updating product from admin panel

---

**Date:** October 12, 2025  
**Status:** Ready to run with auto-migration

