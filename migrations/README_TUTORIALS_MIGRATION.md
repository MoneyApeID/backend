# Migration Tutorials Table

Dokumentasi untuk menambahkan tabel `tutorials` ke database menggunakan Docker.

## Opsi 1: Menggunakan File Migration (Recommended)

### PowerShell (Windows)

```powershell
# Pastikan container MySQL sudah running
docker ps | Select-String "vla-mysql"

# Import migration file
Get-Content .\migrations\create_tutorials_table.sql | docker exec -i vla-mysql mysql -u root -pvlaroot vla-db
```

### Bash (Linux/Mac)

```bash
# Pastikan container MySQL sudah running
docker ps | grep vla-mysql

# Import migration file
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db < migrations/create_tutorials_table.sql
```

### Atau menggunakan script PowerShell

```powershell
.\migrations\run_tutorials_migration.ps1
```

### Atau menggunakan script Bash

```bash
chmod +x migrations/run_tutorials_migration.sh
./migrations/run_tutorials_migration.sh
```

## Opsi 2: Menjalankan SQL Langsung

### PowerShell (Windows)

```powershell
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "
CREATE TABLE IF NOT EXISTS \`tutorials\` (
  \`id\` int UNSIGNED NOT NULL,
  \`title\` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  \`image\` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  \`link\` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  \`status\` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  \`created_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  \`updated_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE \`tutorials\`
  ADD PRIMARY KEY (\`id\`),
  ADD KEY \`idx_status\` (\`status\`);

ALTER TABLE \`tutorials\`
  MODIFY \`id\` int UNSIGNED NOT NULL AUTO_INCREMENT;
"
```

### Bash (Linux/Mac)

```bash
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "
CREATE TABLE IF NOT EXISTS \`tutorials\` (
  \`id\` int UNSIGNED NOT NULL,
  \`title\` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  \`image\` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  \`link\` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  \`status\` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  \`created_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  \`updated_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE \`tutorials\`
  ADD PRIMARY KEY (\`id\`),
  ADD KEY \`idx_status\` (\`status\`);

ALTER TABLE \`tutorials\`
  MODIFY \`id\` int UNSIGNED NOT NULL AUTO_INCREMENT;
"
```

## Opsi 3: Menggunakan Docker Compose Exec

```powershell
# PowerShell
docker-compose exec db mysql -u root -pvlaroot vla-db -e "SOURCE /path/to/migrations/create_tutorials_table.sql"
```

## Verifikasi

Setelah migration berhasil, verifikasi dengan command berikut:

### PowerShell

```powershell
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "SHOW TABLES LIKE 'tutorials';"
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "DESCRIBE tutorials;"
```

### Bash

```bash
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "SHOW TABLES LIKE 'tutorials';"
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "DESCRIBE tutorials;"
```

## Catatan

- Ganti `vlaroot` dengan password root MySQL Anda jika berbeda
- Ganti `vla-db` dengan nama database Anda jika berbeda
- Ganti `vla-mysql` dengan nama container MySQL Anda jika berbeda
- Pastikan container MySQL sudah running sebelum menjalankan migration

## Troubleshooting

### Error: Container tidak ditemukan
```powershell
# Cek container yang running
docker ps

# Start container jika belum running
docker-compose up -d db
```

### Error: Access denied
```powershell
# Pastikan menggunakan user dan password yang benar
# Atau gunakan user dari environment variable
docker exec -i vla-mysql mysql -u vlauser -p${DB_PASS} vla-db < migrations/create_tutorials_table.sql
```

### Error: Table already exists
Jika tabel sudah ada, migration akan di-skip karena menggunakan `CREATE TABLE IF NOT EXISTS`. 
Untuk menghapus dan membuat ulang:

```powershell
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "DROP TABLE IF EXISTS tutorials;"
# Kemudian jalankan migration lagi
```

