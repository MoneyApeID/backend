# Troubleshooting S3 Upload untuk Tutorials

## Error: "Gagal upload gambar"

### Checklist Environment Variables

Pastikan environment variables berikut sudah di-set di `.env` file:

```env
S3_BUCKET_SERVER=your-bucket-name
S3_REGION=ap-southeast-1
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BASE_URL=https://your-custom-domain.com  # Optional, jika menggunakan custom domain
```

### Verifikasi Environment Variables

1. **Cek di container Docker:**
   ```powershell
   docker exec vla-app env | grep S3
   ```

2. **Cek di .env file:**
   Pastikan file `.env` ada di root project dan berisi semua variable di atas.

### Common Issues

#### 1. Environment Variable Tidak Ter-Load
- Pastikan Docker Compose membaca `.env` file
- Restart container setelah mengubah `.env`:
  ```powershell
  docker-compose down
  docker-compose up -d
  ```

#### 2. S3 Credentials Tidak Valid
- Pastikan `S3_ACCESS_KEY` dan `S3_SECRET_KEY` benar
- Pastikan credentials memiliki permission untuk:
  - `s3:PutObject` (untuk upload)
  - `s3:DeleteObject` (untuk delete)
  - `s3:GetObject` (untuk read, optional)

#### 3. Bucket Tidak Ada atau Tidak Accessible
- Pastikan bucket `S3_BUCKET_SERVER` sudah dibuat
- Pastikan bucket berada di region yang sama dengan `S3_REGION`
- Pastikan bucket policy mengizinkan upload dari credentials Anda

#### 4. Region Tidak Sesuai
- Pastikan `S3_REGION` sesuai dengan region bucket Anda
- Contoh region: `ap-southeast-1` (Singapore), `us-east-1` (N. Virginia), dll.

#### 5. Network/Firewall Issues
- Pastikan container dapat mengakses S3 (tidak terblokir firewall)
- Untuk development, pastikan koneksi internet tersedia

### Debug Steps

1. **Cek Logs:**
   ```powershell
   docker logs vla-app | Select-String "S3"
   ```

2. **Test S3 Connection:**
   Buat script test sederhana untuk test koneksi S3 (optional)

3. **Cek Error Message:**
   Setelah update code terbaru, error message akan lebih detail di server logs:
   ```powershell
   docker logs -f vla-app
   ```

### S3 Bucket Policy Example

Jika menggunakan AWS S3, pastikan bucket policy mengizinkan upload:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::ACCOUNT_ID:user/YOUR_USER"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::YOUR_BUCKET_NAME/*"
    },
    {
      "Sid": "AllowDeleteObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::ACCOUNT_ID:user/YOUR_USER"
      },
      "Action": "s3:DeleteObject",
      "Resource": "arn:aws:s3:::YOUR_BUCKET_NAME/*"
    }
  ]
}
```

### Alternative: Menggunakan S3-Compatible Storage

Jika menggunakan S3-compatible storage (seperti MinIO, DigitalOcean Spaces, dll):
- Pastikan `S3_REGION` sesuai dengan endpoint
- Untuk custom endpoint, mungkin perlu modifikasi `getS3Config()` di `utils/s3.go`

### Quick Test

Setelah memperbaiki environment variables, test upload lagi:
1. Restart container: `docker-compose restart app`
2. Coba upload tutorial lagi melalui API
3. Cek logs untuk error detail: `docker logs -f vla-app`

