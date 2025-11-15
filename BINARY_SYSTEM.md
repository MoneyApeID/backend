# Sistem Binary Kiri-Kanan dengan Reward

## Overview
Sistem binary kiri-kanan menggantikan sistem task sebelumnya. Setiap user yang memiliki investasi aktif dan memiliki downline akan mendapatkan reward berdasarkan omset dari level 1-3 di binary tree mereka.

## Struktur Binary
- Setiap user memiliki 2 posisi: **Kiri** dan **Kanan**
- Saat user baru register dengan referral code, mereka akan di-assign ke posisi kiri atau kanan dari upline mereka
- Prioritas: Kiri dulu, baru kanan
- Jika kedua sisi sudah terisi, user baru akan di-assign ke level berikutnya

## Omset Calculation
- **Omset** = Total `total_returned` (penghasilan dari return harian yang sudah dibayar) dari semua investasi aktif (status `Running`) dari semua downline di level 1-3
- Contoh: Bawahan investasi 100k dengan daily profit 1k, jika sudah dibayar 10 hari, maka omset = 10k (1k x 10 hari)
- Omset dihitung terpisah untuk sisi **Kiri** dan **Kanan**
- **Total Omset** = Omset Kiri + Omset Kanan

## Reward System
Ada 5 jenis reward dengan target omset dan durasi berbeda:

| Reward | Target Omset | Hadiah | Durasi | Reset |
|--------|-------------|--------|--------|-------|
| Money Reward 1 | Rp5.000.000 | Rp500.000 (tunai) | 30 hari | Reset tiap bulan |
| Money Reward 2 | Rp15.000.000 | Ponsel/Gadget (Rp1,5–2 juta) | 30 hari | Reset tiap bulan |
| Money Reward 3 | Rp30.000.000 | Kulkas kecil/TV (Rp2,5–3 juta) | 60 hari | Reset tiap 2 bulan |
| Money Reward 4 | Rp75.000.000 | Uang tunai 10.000.000 | 90 hari | Akumulasi (tidak reset) |
| Money Reward 5 | Rp200.000.000 | Rp20.000.000 (tunai) | 150 hari | Akumulasi (tidak reset) |

### Reset vs Akumulasi
- **Reset**: Omset akan di-reset ke 0 setelah periode berakhir, dan periode baru dimulai
- **Akumulasi**: Omset terus bertambah tanpa reset sampai mencapai target

## API Endpoints

## Admin API Endpoints

### 1. GET /api/admin/binary
Melihat semua binary structure dari semua user (atau filter by user_id)

**Query Parameters:**
- `user_id` (optional): Filter by specific user ID

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "binary_structures": [
      {
        "user_id": 1,
        "user_name": "John Doe",
        "user_number": "081234567890",
        "left_id": 2,
        "right_id": 3,
        "left_name": "Jane Doe",
        "right_name": "Bob Smith",
        "omset_left": 10000000,
        "omset_right": 15000000,
        "total_omset": 25000000,
        "level1_count": 2,
        "level2_count": 4,
        "level3_count": 8
      }
    ]
  }
}
```

### 2. GET /api/admin/binary/rewards
Melihat semua reward progress dari semua user

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "reward_progress": [
      {
        "id": 1,
        "user_id": 1,
        "user_name": "John Doe",
        "user_number": "081234567890",
        "reward_id": 1,
        "reward_name": "Money Reward 1",
        "omset_target": 5000000,
        "reward_desc": "Rp500.000 (tunai)",
        "omset_left": 2000000,
        "omset_right": 1500000,
        "total_omset": 3500000,
        "is_completed": false,
        "is_claimed": false,
        "started_at": "2025-01-01T00:00:00Z",
        "expires_at": "2025-01-31T23:59:59Z",
        "progress": 70.0
      }
    ]
  }
}
```

## User API Endpoints

### 1. GET /api/users/binary/structure
Melihat struktur binary user (kiri/kanan)

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "rewards": [
      {
        "id": 1,
        "name": "Money Reward 1",
        "omset_target": 5000000,
        "reward_desc": "Rp500.000 (tunai)",
        "duration": 30,
        "is_accumulative": false,
        "omset_left": 2000000,
        "omset_right": 1500000,
        "total_omset": 3500000,
        "is_completed": false,
        "is_claimed": false,
        "started_at": "2025-01-01T00:00:00Z",
        "expires_at": "2025-01-31T23:59:59Z",
        "progress": 70.0
      }
    ]
  }
}
```

## Database Schema

### binary_nodes
Menyimpan struktur binary kiri-kanan untuk setiap user.

### rewards
Menyimpan definisi reward yang tersedia.

### reward_progress
Menyimpan progress reward untuk setiap user, termasuk:
- Omset kiri dan kanan
- Total omset
- Status completed dan claimed
- Periode mulai dan berakhir

## Automatic Updates

1. **Saat User Register**: User baru akan di-assign ke binary tree upline mereka
2. **Saat User Investasi**: 
   - Reward progress di-initialize untuk user yang baru aktif
   - Reward progress di-update untuk user dan semua upline yang terpengaruh (level 1-3)
3. **Cron Job** (opsional): Bisa ditambahkan untuk update reward progress secara berkala

## Manual Claim
Reward di-claim secara manual oleh admin. Sistem hanya men-track progress, tidak otomatis memberikan reward.

## Migration

1. Import `db.sql` untuk membuat tabel `binary_nodes`, `rewards`, dan `reward_progress`
2. Run seed data untuk rewards:
   ```sql
   source migrations/seed_rewards.sql
   ```

## Notes
- Hanya user dengan `investment_status = 'Active'` yang akan di-track reward progress-nya
- Omset hanya dihitung dari investasi dengan status `Running`
- Reward progress akan otomatis di-reset untuk reward yang memiliki `is_accumulative = false` setelah periode berakhir

