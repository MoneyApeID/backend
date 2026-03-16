# Money Rich VPS Runbook

Dokumen ini dipakai untuk deploy proyek `Money Ape / Money Rich` ke VPS Ubuntu dengan domain:

- `moneyrich.co`
- `www.moneyrich.co`
- `api.moneyrich.co`

Struktur di VPS:

- `/var/www/moneyrich/backend`
- `/var/www/moneyrich/frontend`

Arsitektur production:

- Backend Go berjalan di Docker Compose pada port internal `8080`
- Frontend Next.js berjalan via PM2 pada port internal `3000`
- Nginx di host VPS menjadi reverse proxy untuk web dan API
- SSL memakai Let's Encrypt

## 1. DNS

Pastikan record berikut mengarah ke IP VPS `18.140.2.192`:

- `moneyrich.co` -> `18.140.2.192`
- `www.moneyrich.co` -> `18.140.2.192`
- `api.moneyrich.co` -> `18.140.2.192`

## 2. Login VPS

Dari laptop Windows:

```powershell
ssh -i C:\Users\USER\.ssh\siapkerja.pem ubuntu@18.140.2.192
```

## 3. Bootstrap VPS

Jalankan di VPS:

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y nginx certbot python3-certbot-nginx unzip curl wget rsync ufw
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker ubuntu
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
sudo npm install -g pm2
sudo mkdir -p /var/www/moneyrich/backend /var/www/moneyrich/frontend
sudo chown -R ubuntu:ubuntu /var/www/moneyrich
```

Logout lalu login lagi agar group Docker aktif.

## 4. Upload Project Langsung Dari Laptop

Dari folder lokal proyek:

```powershell
tar --exclude=BackEnd/.env --exclude=FrontEnd/.env --exclude=FrontEnd/.next --exclude=FrontEnd/node_modules -czf backend.tar.gz -C BackEnd .
tar --exclude=FrontEnd/.env --exclude=FrontEnd/.next --exclude=FrontEnd/node_modules -czf frontend.tar.gz -C FrontEnd .
scp -i C:\Users\USER\.ssh\siapkerja.pem backend.tar.gz ubuntu@18.140.2.192:/var/www/moneyrich/
scp -i C:\Users\USER\.ssh\siapkerja.pem frontend.tar.gz ubuntu@18.140.2.192:/var/www/moneyrich/
```

Lalu di VPS:

```bash
cd /var/www/moneyrich
rm -rf backend/* frontend/*
tar -xzf backend.tar.gz -C backend
tar -xzf frontend.tar.gz -C frontend
rm -f backend.tar.gz frontend.tar.gz
```

## 5. Backend Environment

Buat file `/var/www/moneyrich/backend/.env`:

```env
ENV=production
PORT=8080
APP_PORT=8080

DB_HOST=db
DB_PORT=3306
DB_USER=moneyrich
DB_PASS=ISI_PASSWORD_DB
DB_ROOT_PASSWORD=ISI_ROOT_PASSWORD_DB
DB_NAME=moneyrich
DB_TLS=false

REDIS_ADDR=redis:6379
REDIS_PASS=ISI_PASSWORD_REDIS
REDIS_DB=0

JWT_SECRET=ISI_JWT_SECRET_MIN_32_CHAR
CRON_KEY=ISI_CRON_KEY
SF_API_KEY=ISI_INTERNAL_API_KEY
JWT_AUD=
JWT_ISS=

CORS_ALLOWED_ORIGINS=https://moneyrich.co,https://www.moneyrich.co

PAKAILINK_BASE_URL=https://api.pakailink.id
PAKAILINK_CLIENT_KEY=ee7f8fc2564211f0a993fa163e117483
PAKAILINK_CLIENT_SECRET=921988da7032bc8683c795dba81e4e84
PAKAILINK_PARTNER_ID=PTR00000TI
PAKAILINK_MERCHANT_ID=1031332780
PAKAILINK_STORE_ID=GATEPAY
PAKAILINK_TERMINAL_ID=1701103543259012
PAKAILINK_CHANNEL_ID=95222
PAKAILINK_PRIVATE_KEY_PATH=./keys/rsa_private_key.pem
PAKAILINK_PAYMENT_CALLBACK_URL=https://api.moneyrich.co/api/callback/payments
PAKAILINK_PAYOUT_CALLBACK_URL=https://api.moneyrich.co/api/callback/payouts
PAKAILINK_CALLBACK_PUBLIC_KEY_PATH=
```

Catatan:

- File private key harus ada di `/var/www/moneyrich/backend/keys/rsa_private_key.pem`
- `PAKAILINK_CALLBACK_PUBLIC_KEY_PATH` boleh dikosongkan sampai public key callback resmi diberikan PakaiLink

## 6. Frontend Environment

Buat file `/var/www/moneyrich/frontend/.env`:

```env
NEXT_PUBLIC_API_URL=https://api.moneyrich.co/api
NEXT_PUBLIC_S3_ENDPOINT=s3.ap-southeast-1.amazonaws.com

S3_REGION=ap-southeast-1
S3_BUCKET=moneyrich
S3_BUCKET_SERVER=moneyrich
S3_ACCESS_KEY=AKIAVTZGAV4ZU7SFELPB
S3_SECRET_KEY=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
```

Penting:

- Credential S3 sengaja memakai env server-only, bukan `NEXT_PUBLIC_*`
- Jangan menaruh `S3_ACCESS_KEY` atau `S3_SECRET_KEY` di variabel public

## 7. Start Backend

Di VPS:

```bash
cd /var/www/moneyrich/backend
docker compose up -d --build
docker compose ps
docker compose logs -f app
```

Import schema awal:

```bash
cd /var/www/moneyrich/backend
docker exec -i moneyrich-mysql mysql -u root -p${DB_ROOT_PASSWORD} ${DB_NAME} < database/db.sql
```

Kalau perintah shell di atas tidak membaca `${DB_ROOT_PASSWORD}`, jalankan versi literal:

```bash
docker exec -i moneyrich-mysql mysql -u root -pYOUR_ROOT_PASSWORD moneyrich < /var/www/moneyrich/backend/database/db.sql
```

## 8. Start Frontend

Di VPS:

```bash
cd /var/www/moneyrich/frontend
npm ci
npm run build
pm2 delete frontend || true
pm2 start npm --name frontend -- start -- --hostname 127.0.0.1 --port 3000
pm2 save
pm2 startup systemd -u ubuntu --hp /home/ubuntu
```

## 9. Nginx

Buat file `/etc/nginx/sites-available/moneyrich.conf`:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name moneyrich.co www.moneyrich.co;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

server {
    listen 80;
    listen [::]:80;
    server_name api.moneyrich.co;

    client_max_body_size 20M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Aktifkan:

```bash
sudo ln -sf /etc/nginx/sites-available/moneyrich.conf /etc/nginx/sites-enabled/moneyrich.conf
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

## 10. SSL

Jalankan:

```bash
sudo certbot --nginx -d moneyrich.co -d www.moneyrich.co -d api.moneyrich.co
sudo certbot renew --dry-run
```

## 11. Verifikasi

Backend:

```bash
curl https://api.moneyrich.co/health
curl https://api.moneyrich.co/api/health
```

Frontend:

```bash
curl -I https://moneyrich.co
curl -I https://www.moneyrich.co
```

Proses:

```bash
docker ps
pm2 status
sudo systemctl status nginx
```

## 12. Redeploy Berikutnya

Dari laptop:

```powershell
tar --exclude=BackEnd/.env --exclude=FrontEnd/.env --exclude=FrontEnd/.next --exclude=FrontEnd/node_modules -czf backend.tar.gz -C BackEnd .
tar --exclude=FrontEnd/.env --exclude=FrontEnd/.next --exclude=FrontEnd/node_modules -czf frontend.tar.gz -C FrontEnd .
scp -i C:\Users\USER\.ssh\siapkerja.pem backend.tar.gz ubuntu@18.140.2.192:/var/www/moneyrich/
scp -i C:\Users\USER\.ssh\siapkerja.pem frontend.tar.gz ubuntu@18.140.2.192:/var/www/moneyrich/
```

Di VPS:

```bash
cd /var/www/moneyrich
tar -xzf backend.tar.gz -C backend
tar -xzf frontend.tar.gz -C frontend
rm -f backend.tar.gz frontend.tar.gz

cd /var/www/moneyrich/backend
docker compose up -d --build

cd /var/www/moneyrich/frontend
npm ci
npm run build
pm2 restart moneyrich-frontend
```

## 13. Operasional Singkat

Lihat log backend:

```bash
cd /var/www/moneyrich/backend
docker compose logs -f app
```

Lihat log frontend:

```bash
pm2 logs moneyrich-frontend
```

Rollback cepat frontend:

```bash
pm2 restart moneyrich-frontend
```

Restart backend:

```bash
cd /var/www/moneyrich/backend
docker compose restart app
```
