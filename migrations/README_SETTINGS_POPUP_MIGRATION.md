# Migration Settings Popup

Dokumentasi untuk mengupdate tabel `settings` - mengubah `logo` menjadi `popup` dan menambahkan `popup_title`, `created_at`, `updated_at`.

## Perubahan Database

1. **Rename column**: `logo` → `popup`
2. **Add column**: `popup_title` (varchar(255), nullable)
3. **Add column**: `created_at` (datetime, default CURRENT_TIMESTAMP)
4. **Add column**: `updated_at` (datetime, default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)

## Command Docker untuk Migration

### PowerShell (Windows)

```powershell
# Pastikan container MySQL sudah running
docker ps | Select-String "vla-mysql"

# Import migration file
Get-Content .\migrations\update_settings_popup.sql | docker exec -i vla-mysql mysql -u root -pvlaroot vla-db
```

### Bash (Linux/Mac)

```bash
# Pastikan container MySQL sudah running
docker ps | grep vla-mysql

# Import migration file
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db < migrations/update_settings_popup.sql
```

### Atau menggunakan script PowerShell

```powershell
.\migrations\run_update_settings_popup.ps1
```

## Opsi 2: Menjalankan SQL Langsung

### PowerShell (Windows)

```powershell
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "
ALTER TABLE \`settings\` CHANGE COLUMN \`logo\` \`popup\` text DEFAULT NULL;
ALTER TABLE \`settings\` ADD COLUMN \`popup_title\` varchar(255) DEFAULT NULL AFTER \`popup\`;
ALTER TABLE \`settings\` ADD COLUMN \`created_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER \`link_app\`;
ALTER TABLE \`settings\` ADD COLUMN \`updated_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER \`created_at\`;
"
```

### Bash (Linux/Mac)

```bash
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "
ALTER TABLE \`settings\` CHANGE COLUMN \`logo\` \`popup\` text DEFAULT NULL;
ALTER TABLE \`settings\` ADD COLUMN \`popup_title\` varchar(255) DEFAULT NULL AFTER \`popup\`;
ALTER TABLE \`settings\` ADD COLUMN \`created_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER \`link_app\`;
ALTER TABLE \`settings\` ADD COLUMN \`updated_at\` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER \`created_at\`;
"
```

## Verifikasi

Setelah migration berhasil, verifikasi dengan command berikut:

### PowerShell

```powershell
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "DESCRIBE settings;"
```

### Bash

```bash
docker exec -i vla-mysql mysql -u root -pvlaroot vla-db -e "DESCRIBE settings;"
```

## Catatan

- Ganti `vlaroot` dengan password root MySQL Anda jika berbeda
- Ganti `vla-db` dengan nama database Anda jika berbeda
- Ganti `vla-mysql` dengan nama container MySQL Anda jika berbeda
- Pastikan container MySQL sudah running sebelum menjalankan migration
- Jika kolom sudah ada, migration akan gagal dengan error "Duplicate column name" - ini normal dan bisa diabaikan

## Troubleshooting

### Error: Duplicate column name
Jika kolom `popup_title`, `created_at`, atau `updated_at` sudah ada, migration akan gagal. Ini normal. Anda bisa:
1. Skip error tersebut (kolom sudah ada)
2. Atau hapus kolom terlebih dahulu jika perlu

### Error: Unknown column 'logo'
Jika kolom `logo` tidak ada, berarti sudah diubah sebelumnya. Anda bisa skip bagian `CHANGE COLUMN`.

