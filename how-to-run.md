# 🚀 VPS Deployment Guide - Go Backend & Next.js Frontend

Dokumentasi lengkap untuk menjalankan aplikasi Go Backend dan Next.js Frontend di VPS Ubuntu/Debian dengan setup domain terpisah.

## 📋 **PREREQUISITES**

- VPS Ubuntu 20.04+ atau Debian 11+
- Root access atau user dengan sudo privileges
- 2 Domain yang sudah siap: `web.com` dan `api.web.com`
- Minimal 2GB RAM, 2 CPU cores, 20GB storage
- Akses ke DNS management (Cloudflare recommended)

## 🌐 **STEP 1: SETUP DOMAIN & DNS**

### 1.1 Point Domain ke VPS
```bash
# Di DNS provider (Cloudflare/Namecheap/etc), tambahkan A records:
# web.com        -> IP_VPS_ANDA
# api.web.com    -> IP_VPS_ANDA
# www.web.com    -> IP_VPS_ANDA (optional)
```

### 1.2 Verifikasi DNS Propagation
```bash
# Cek DNS propagation
nslookup web.com
nslookup api.web.com

# Atau gunakan online tools:
# https://dnschecker.org/
```

## 🔧 **STEP 2: SETUP VPS**

### 2.1 Update System
```bash
sudo apt update && sudo apt upgrade -y
```

### 2.2 Install Dependencies
```bash
sudo apt install -y curl wget git nginx certbot python3-certbot-nginx htop ufw
```

### 2.3 Install Docker & Docker Compose
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Logout dan login ulang untuk apply Docker group
newgrp docker
```

### 2.4 Install Node.js untuk Frontend
```bash
# Install Node.js 18.x
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install PM2 untuk process management
sudo npm install -g pm2
```

## 🔒 **STEP 3: SETUP SSL CERTIFICATES**

### 3.1 Generate SSL untuk kedua domain
```bash
# Stop nginx
sudo systemctl stop nginx
sudo systemctl disable nginx
# Generate SSL certificate untuk main domain
sudo certbot certonly --standalone -d web.com -d www.web.com

# Generate SSL certificate untuk API domain
sudo certbot certonly --standalone -d api.web.com

# Verifikasi certificates
sudo certbot certificates
```

### 3.2 Setup Auto-renewal
```bash
# Test renewal
sudo certbot renew --dry-run

# Auto renewal sudah otomatis via systemd timer
sudo systemctl status certbot.timer

# Enable nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

## ⚙️ **STEP 4: SETUP BACKEND (Go + Docker)**

### 4.1 Clone Backend Repository
```bash
# Clone repository ke direktori backend
git clone https://github.com/your-username/backend-repo.git /home/$USER/backend
cd /home/$USER/backend
```

### 4.2 Setup Environment Variables
```bash
# Copy environment template
cp .env.example .env

# Edit environment variables
nano .env
```

**Isi .env:**
```env
# Database Configuration
DB_HOST=db
DB_PORT=3306
DB_USER=app_user
DB_PASS=your_secure_password_here
DB_NAME=app_database
DB_ROOT_PASSWORD=your_root_password_here
DB_TLS=false
DB_TLS_VERIFY=false
DB_PARAMS=charset=utf8mb4&parseTime=True&loc=Local&tls=false&timeout=10s&readTimeout=10s&writeTimeout=10s

# Redis Configuration
REDIS_ADDR=redis:6379
REDIS_PASS=your_redis_password_here
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_minimum_32_characters
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# S3/MinIO Configuration (Optional)
S3_ENDPOINT=https://s3.amazonaws.com
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=your_bucket_name
S3_REGION=us-east-1

# Payment Gateway Configuration
PAKASIR_API_KEY=your_pakasir_api_key
PAKASIR_PROJECT=your_project_name
PAKASIR_BASE_URL=https://app.pakasir.com

KLIKPAY_API_KEY=your_klikpay_api_key
KLIKPAY_PROJECT=your_project_name
KLIKPAY_BASE_URL=https://app.klikpay.com

# Application Configuration
ENV=production
PORT=8080
APP_PORT=8080
API_HOST=https://api.web.com

# Cron Configuration
CRON_KEY=your_cron_secret_key_here
```

### 4.3 Start Backend Services
```bash
# Build dan start services
docker compose up -d --build

# Verify services
docker compose ps

# Check logs
docker compose logs -f app
```

### 4.4 Verify Backend Database
```bash
# Check database connection
docker exec backend-mysql-1 mysql -u root -p${DB_ROOT_PASSWORD} -e "SHOW DATABASES;"

##Create migration table
docker exec -i vla-mysql mysql -u root -pvlaroot vla-ciroos < ./database/db.sql

# Verify tables
docker exec vla-mysql mysql -u root -pvlaroot vla-db -e "SHOW TABLES;"
```

## 🎨 **STEP 5: SETUP FRONTEND (Next.js)**

### 5.1 Clone Frontend Repository
```bash
# Clone frontend repository
git clone https://github.com/your-username/frontend-repo.git /home/$USER/frontend
cd /home/$USER/frontend
```

### 5.2 Install Dependencies
```bash
# Install NPM dependencies
npm install

# atau jika menggunakan yarn
# yarn install
```

### 5.3 Setup Environment Variables
```bash
#Copy .env.example to .env
cp .env.example .env

# Create environment file
nano .env
```

**Isi .env.local:**
```env
# API Configuration
NEXT_PUBLIC_API_URL=https://api.web.com
NEXT_PUBLIC_API_VERSION=v1

# Application Configuration
NEXT_PUBLIC_APP_NAME=YourAppName
NEXT_PUBLIC_APP_VERSION=1.0.0
NEXT_PUBLIC_APP_DESCRIPTION=Your App Description

# Frontend URL
NEXT_PUBLIC_FRONTEND_URL=https://web.com

# Feature Flags (optional)
NEXT_PUBLIC_ENABLE_ANALYTICS=true
NEXT_PUBLIC_ENABLE_PWA=false

# Environment
NODE_ENV=production
```

### 5.4 Update Next.js Configuration
```bash
# Edit next.config.js
nano next.config.js
```

**Isi next.config.js:**
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  
  // Image optimization
  images: {
    domains: ['api.web.com'],
    unoptimized: false
  },
  
  // Build configuration
  eslint: {
    ignoreDuringBuilds: false,
  },
  typescript: {
    ignoreBuildErrors: false,
  },
  
  // Performance optimization
  experimental: {
    optimizeCss: true,
    optimizePackageImports: ['lucide-react', '@heroicons/react']
  },
  
  // Security headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin',
          },
        ],
      },
    ]
  },
  
  // Redirects
  async redirects() {
    return [
      {
        source: '/admin',
        destination: '/admin/login',
        permanent: false,
      },
    ]
  },
}

module.exports = nextConfig
```

### 5.5 Build dan Start Frontend
```bash
# Build aplikasi
npm run build

# Start dengan PM2
pm2 start npm --name "frontend" -- start

# Save PM2 configuration
pm2 save

# Setup PM2 startup script
pm2 startup
# Jalankan command yang muncul dari output diatas
```

### 5.6 Verify Frontend
```bash
# Check PM2 status
pm2 status

# Check logs
pm2 logs frontend

# Monitor resources
pm2 monit
```

## 🌐 **STEP 6: SETUP NGINX REVERSE PROXY**

### 6.1 Create API Domain Configuration
```bash
# Create nginx config untuk API domain
sudo nano /etc/nginx/sites-available/api.web.com
```

**Isi api.web.com config:**
```nginx
# HTTP redirect to HTTPS
server {
    listen 80;
    server_name api.web.com;
    return 301 https://$host$request_uri;
}

# HTTPS API Server
server {
    listen 443 ssl http2;
    server_name api.web.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/api.web.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.web.com/privkey.pem;
    
    # SSL Security
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers off;
    
    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    
    # Handle preflight requests
    location / {
        # Handle OPTIONS requests first
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "https://web.com";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Accept, Authorization, Cache-Control, Content-Type, DNT, If-Modified-Since, Keep-Alive, Origin, User-Agent, X-Requested-With";
            add_header Access-Control-Max-Age 1728000;
            add_header Content-Type "text/plain; charset=utf-8";
            add_header Content-Length 0;
            return 204;
        }
        
        # Proxy to Go backend
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Static files caching
    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
        proxy_pass http://127.0.0.1:8080;
        expires 1M;
        add_header Cache-Control "public, immutable";
    }
}
```

### 6.2 Create Frontend Domain Configuration
```bash
# Create nginx config untuk frontend domain
sudo nano /etc/nginx/sites-available/web.com
```

**Isi web.com config:**
```nginx
# HTTP redirect to HTTPS
server {
    listen 80;
    server_name web.com www.web.com;
    return 301 https://web.com$request_uri;
}

# HTTPS redirect www to non-www
server {
    listen 443 ssl http2;
    server_name www.web.com;
    
    ssl_certificate /etc/letsencrypt/live/web.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/web.com/privkey.pem;
    
    return 301 https://web.com$request_uri;
}

# HTTPS Frontend Server
server {
    listen 443 ssl http2;
    server_name web.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/web.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/web.com/privkey.pem;
    
    # SSL Security
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers off;
    
    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' https://api.web.com http: https: data: blob: 'unsafe-inline'" always;
    
    # Proxy to Next.js frontend
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Next.js static files
    location /_next/static {
        proxy_pass http://127.0.0.1:3000;
        expires 365d;
        add_header Cache-Control "public, immutable";
    }
    
    # Next.js images
    location /_next/image {
        proxy_pass http://127.0.0.1:3000;
        expires 365d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 6.3 Enable Sites dan Test Configuration
```bash
# Enable sites
sudo ln -s /etc/nginx/sites-available/api.web.com /etc/nginx/sites-enabled/
sudo ln -s /etc/nginx/sites-available/web.com /etc/nginx/sites-enabled/

# Remove default nginx site
sudo rm /etc/nginx/sites-enabled/default

# Test nginx configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
sudo systemctl enable nginx
```

## ✅ **STEP 7: VERIFICATION & TESTING**

### 7.1 Test Backend API
```bash
# Test health endpoint
curl https://api.web.com/health

# Test API endpoints
curl https://api.web.com/api/v1/ping

# Test with authentication (jika ada)
curl -H "Authorization: Bearer YOUR_TOKEN" https://api.web.com/api/v1/protected
```

### 7.2 Test Frontend
```bash
# Test frontend loading
curl -I https://web.com

# Test specific pages
curl -I https://web.com/admin
curl -I https://web.com/login
```

### 7.3 Test in Browser
1. Open `https://web.com` - Should load frontend
2. Open `https://api.web.com/health` - Should return API health status
3. Open `https://www.web.com` - Should redirect to `https://web.com`
4. Check console for any CORS or API connection errors

### 7.4 Performance Testing
```bash
# Install Apache Bench untuk testing
sudo apt install apache2-utils

# Test API performance
ab -n 100 -c 10 https://api.web.com/health

# Test frontend performance
ab -n 100 -c 10 https://web.com/
```

## ☁️ **STEP 8: CLOUDFLARE SETUP (OPTIONAL)**

### 8.1 Add Site to Cloudflare
1. Login ke Cloudflare Dashboard
2. Click "Add Site" dan masukkan `web.com`
3. Choose plan (Free plan sudah cukup)
4. Cloudflare akan scan DNS records

### 8.2 Update DNS Records
```
Type    Name        Content         Proxy Status
A       web.com     YOUR_VPS_IP     Proxied (Orange Cloud)
A       api.web.com YOUR_VPS_IP     Proxied (Orange Cloud)
CNAME   www         web.com         Proxied (Orange Cloud)
```

### 8.3 Configure Cloudflare Settings
**SSL/TLS Settings:**
- SSL/TLS encryption mode: `Full (strict)`
- Always Use HTTPS: `On`
- Minimum TLS Version: `1.2`

**Security Settings:**
- Security Level: `Medium`
- Bot Fight Mode: `On`
- Browser Integrity Check: `On`

**Speed Settings:**
- Auto Minify: Enable `HTML`, `CSS`, `JS`
- Brotli: `On`
- Early Hints: `On`

**Page Rules (Optional):**
1. `www.web.com/*` → `https://web.com/$1` (301 Redirect)
2. `web.com/*` → Cache Level: `Cache Everything`

### 8.4 Update Nginx untuk Cloudflare
```bash
# Edit nginx config untuk real IP
sudo nano /etc/nginx/nginx.conf
```

Add dalam `http` block:
```nginx
# Cloudflare real IP
set_real_ip_from 103.21.244.0/22;
set_real_ip_from 103.22.200.0/22;
set_real_ip_from 103.31.4.0/22;
set_real_ip_from 104.16.0.0/13;
set_real_ip_from 104.24.0.0/14;
set_real_ip_from 108.162.192.0/18;
set_real_ip_from 131.0.72.0/22;
set_real_ip_from 141.101.64.0/18;
set_real_ip_from 162.158.0.0/15;
set_real_ip_from 172.64.0.0/13;
set_real_ip_from 173.245.48.0/20;
set_real_ip_from 188.114.96.0/20;
set_real_ip_from 190.93.240.0/20;
set_real_ip_from 197.234.240.0/22;
set_real_ip_from 198.41.128.0/17;
set_real_ip_from 2400:cb00::/32;
set_real_ip_from 2606:4700::/32;
set_real_ip_from 2803:f800::/32;
set_real_ip_from 2405:b500::/32;
set_real_ip_from 2405:8100::/32;
set_real_ip_from 2c0f:f248::/32;
set_real_ip_from 2a06:98c0::/29;

real_ip_header CF-Connecting-IP;
```

## 🔧 **STEP 9: MONITORING & MAINTENANCE**

### 9.1 Setup System Monitoring
```bash
# Install monitoring tools
sudo apt install -y htop iotop nethogs

# Check system resources
htop
df -h
free -h
```

### 9.2 Setup Log Monitoring
```bash
# Monitor logs in real-time
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# Backend logs
docker compose logs -f app

# Frontend logs
pm2 logs frontend
```

### 9.3 Setup Automated Backups
```bash
# Create backup script
sudo nano /usr/local/bin/backup.sh
```

**Isi backup.sh:**
```bash
#!/bin/bash

BACKUP_DIR="/backup"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
docker exec backend-mysql-1 mysqldump -u root -p$DB_ROOT_PASSWORD --all-databases > $BACKUP_DIR/db_backup_$DATE.sql

# Backup application files
tar -czf $BACKUP_DIR/app_backup_$DATE.tar.gz /home/$USER/backend /home/$USER/frontend

# Keep only last 7 days of backups
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
```

```bash
# Make executable
sudo chmod +x /usr/local/bin/backup.sh

# Add to crontab (daily backup at 2 AM)
echo "0 2 * * * /usr/local/bin/backup.sh >> /var/log/backup.log 2>&1" | sudo crontab -
```

### 9.4 Setup Health Checks
```bash
# Create health check script
nano /home/$USER/health_check.sh
```

**Isi health_check.sh:**
```bash
#!/bin/bash

# Check backend
if curl -f https://api.web.com/health > /dev/null 2>&1; then
    echo "Backend: OK"
else
    echo "Backend: FAILED"
    # Restart backend if needed
    cd /home/$USER/backend && docker compose restart app
fi

# Check frontend
if curl -f https://web.com > /dev/null 2>&1; then
    echo "Frontend: OK"
else
    echo "Frontend: FAILED"
    # Restart frontend if needed
    pm2 restart frontend
fi

# Check disk space
DISK_USAGE=$(df / | grep -vE '^Filesystem|tmpfs|cdrom' | awk '{print $5}' | sed 's/%//g')
if [ $DISK_USAGE -gt 80 ]; then
    echo "WARNING: Disk usage is ${DISK_USAGE}%"
fi
```

```bash
# Make executable dan add to cron (every 5 minutes)
chmod +x /home/$USER/health_check.sh
echo "*/5 * * * * /home/$USER/health_check.sh >> /var/log/health_check.log 2>&1" | crontab -
```

## 🔐 **STEP 10: SECURITY HARDENING**

### 10.1 Setup Firewall
```bash
# Configure UFW firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH, HTTP, HTTPS
sudo ufw allow 22
sudo ufw allow 80
sudo ufw allow 443

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

### 10.2 Setup Fail2Ban
```bash
# Install fail2ban
sudo apt install fail2ban

# Configure jail
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo nano /etc/fail2ban/jail.local
```

Configure nginx jail:
```ini
[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
port = http,https  
logpath = /var/log/nginx/error.log
maxretry = 10
```

```bash
# Restart fail2ban
sudo systemctl restart fail2ban
sudo systemctl enable fail2ban
```

### 10.3 Secure SSH
```bash
# Edit SSH config
sudo nano /etc/ssh/sshd_config
```

Update:
```
Port 2222                    # Change default port
PermitRootLogin no          # Disable root login
PasswordAuthentication no   # Use keys only
PubkeyAuthentication yes
```

```bash
# Restart SSH
sudo systemctl restart sshd

# Update UFW rules
sudo ufw delete allow 22
sudo ufw allow 2222
```

## 🚨 **TROUBLESHOOTING**

### Common Issues & Solutions:

1. **SSL Certificate Issues**
   ```bash
   # Check certificate status
   sudo certbot certificates
   
   # Renew certificates
   sudo certbot renew --force-renewal
   
   # Check nginx SSL config
   sudo nginx -t
   ```

2. **Backend Connection Failed**
   ```bash
   # Check Docker containers
   docker compose ps
   
   # Check backend logs
   docker compose logs app
   
   # Restart backend
   docker compose restart app
   ```

3. **Frontend Not Loading**
   ```bash
   # Check PM2 status
   pm2 status
   
   # Check frontend logs
   pm2 logs frontend
   
   # Restart frontend
   pm2 restart frontend
   ```

4. **Database Connection Issues**
   ```bash
   # Check database logs
   docker compose logs mysql
   
   # Access database directly
   docker exec -it backend-mysql-1 mysql -u root -p
   ```

5. **CORS Issues**
   ```bash
   # Check nginx configuration
   sudo nginx -t
   
   # Verify CORS headers
   curl -H "Origin: https://web.com" -H "Access-Control-Request-Method: GET" -X OPTIONS https://api.web.com/api/v1/test
   ```

6. **High CPU/Memory Usage**
   ```bash
   # Monitor system resources
   htop
   
   # Check Docker stats
   docker stats
   
   # Check PM2 monitoring
   pm2 monit
   ```

### Emergency Commands:
```bash
# Quick restart all services
sudo systemctl restart nginx
docker compose restart
pm2 restart all

# Check all service status
sudo systemctl status nginx
docker compose ps
pm2 status

# View all logs
sudo tail -f /var/log/nginx/error.log &
docker compose logs -f &
pm2 logs &
```

## 📊 **PERFORMANCE OPTIMIZATION**

### 10.1 Database Optimization
```sql
-- Add to MySQL configuration
[mysqld]
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
max_connections = 100
query_cache_type = 1
query_cache_size = 64M
```

### 10.2 Nginx Optimization
```bash
# Edit nginx.conf
sudo nano /etc/nginx/nginx.conf
```

Add optimizations:
```nginx
worker_processes auto;
worker_connections 1024;

# Enable gzip
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

# Enable caching
open_file_cache max=1000 inactive=20s;
open_file_cache_valid 30s;
open_file_cache_min_uses 2;
open_file_cache_errors on;
```

### 10.3 PM2 Optimization
```bash
# Update PM2 config untuk cluster mode
pm2 delete frontend
pm2 start npm --name "frontend" -i max -- start

# Save configuration
pm2 save
```

## 📝 **DEPLOYMENT CHECKLIST**

### Pre-Deployment:
- [ ] VPS setup dengan spesifikasi minimal
- [ ] Domain `web.com` dan `api.web.com` sudah point ke VPS
- [ ] DNS propagation sudah selesai
- [ ] SSL certificates sudah generated
- [ ] Repository backend dan frontend sudah siap

### During Deployment:
- [ ] System dependencies terinstall
- [ ] Docker dan Docker Compose terinstall
- [ ] Node.js dan PM2 terinstall
- [ ] Backend environment variables dikonfigurasi
- [ ] Frontend environment variables dikonfigurasi
- [ ] Database migrations berhasil
- [ ] Nginx reverse proxy dikonfigurasi
- [ ] SSL certificates applied ke nginx

### Post-Deployment:
- [ ] API endpoints bisa diakses via `https://api.web.com`
- [ ] Frontend bisa diakses via `https://web.com`
- [ ] CORS policy working correctly
- [ ] Database connection stable
- [ ] Monitoring dan logging setup
- [ ] Backup system configured
- [ ] Security hardening implemented

### Testing:
- [ ] API health check: `curl https://api.web.com/health`
- [ ] Frontend loading: `curl https://web.com`
- [ ] SSL certificates valid: Browser security indicator
- [ ] CORS working: Frontend can call API
- [ ] Performance acceptable: Page load < 3 seconds
- [ ] Mobile responsive: Test pada berbagai device

## 📞 **SUPPORT & MAINTENANCE**

### Daily Monitoring:
```bash
# Quick health check
curl https://api.web.com/health && curl https://web.com
pm2 status
docker compose ps
df -h
```

### Weekly Tasks:
```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Check SSL certificate expiry
sudo certbot certificates

# Review logs for errors
sudo tail -100 /var/log/nginx/error.log
```

### Monthly Tasks:
```bash
# Review and rotate logs
sudo logrotate /etc/logrotate.conf

# Update Docker images
cd /home/$USER/backend
docker compose pull
docker compose up -d --build

# Update frontend dependencies (check for security updates)
cd /home/$USER/frontend
npm audit
npm update

# Database maintenance
docker exec backend-mysql-1 mysql -u root -p -e "OPTIMIZE TABLE database_name.*;"

# Clean up old Docker images
docker system prune -f

# Review disk usage and clean up if needed
sudo du -sh /var/log/*
sudo find /var/log -name "*.log" -mtime +30 -delete
```

## 🎯 **PRODUCTION-READY FEATURES**

### 11.1 Rate Limiting dengan Nginx
```nginx
# Add to nginx.conf dalam http block
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
    
    # Dalam server block api.web.com
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        # ... existing proxy config
    }
    
    location /api/auth/login {
        limit_req zone=login burst=5 nodelay;
        # ... existing proxy config
    }
}
```

### 11.2 Redis Caching untuk Performance
```bash
# Verify Redis container
docker exec backend-redis-1 redis-cli ping

# Monitor Redis
docker exec backend-redis-1 redis-cli monitor
```

### 11.3 Database Connection Pooling
Pastikan dalam backend Go code menggunakan connection pooling:
```go
// Example dalam main.go atau database config
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 11.4 Static File Optimization
```nginx
# Add dalam nginx server block untuk frontend
location ~* \.(jpg|jpeg|png|gif|ico|svg|woff|woff2|ttf|eot)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
    add_header Vary Accept-Encoding;
    access_log off;
}

location ~* \.(css|js)$ {
    expires 1M;
    add_header Cache-Control "public";
    add_header Vary Accept-Encoding;
}
```

## 🔄 **CONTINUOUS DEPLOYMENT**

### 12.1 Automated Deployment Script
```bash
# Create deployment script
nano /home/$USER/deploy.sh
```

**Isi deploy.sh:**
```bash
#!/bin/bash

set -e  # Exit on any error

echo "🚀 Starting deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function untuk logging
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Backup before deployment
log "Creating backup..."
/usr/local/bin/backup.sh

# Update backend
log "Updating backend..."
cd /home/$USER/backend
git fetch origin
BACKEND_CURRENT=$(git rev-parse HEAD)
git pull origin main

if [ "$BACKEND_CURRENT" != "$(git rev-parse HEAD)" ]; then
    log "Backend updated, rebuilding..."
    docker compose up -d --build
    
    # Wait for backend to be healthy
    log "Waiting for backend to be healthy..."
    for i in {1..30}; do
        if curl -f https://api.web.com/health > /dev/null 2>&1; then
            log "Backend is healthy!"
            break
        fi
        if [ $i -eq 30 ]; then
            error "Backend health check failed after 30 attempts"
            exit 1
        fi
        sleep 2
    done
else
    log "Backend already up to date"
fi

# Update frontend
log "Updating frontend..."
cd /home/$USER/frontend
git fetch origin
FRONTEND_CURRENT=$(git rev-parse HEAD)
git pull origin main

if [ "$FRONTEND_CURRENT" != "$(git rev-parse HEAD)" ]; then
    log "Frontend updated, rebuilding..."
    npm install
    npm run build
    pm2 restart frontend
    
    # Wait for frontend to be healthy
    log "Waiting for frontend to be healthy..."
    for i in {1..30}; do
        if curl -f https://web.com > /dev/null 2>&1; then
            log "Frontend is healthy!"
            break
        fi
        if [ $i -eq 30 ]; then
            error "Frontend health check failed after 30 attempts"
            exit 1
        fi
        sleep 2
    done
else
    log "Frontend already up to date"
fi

# Reload nginx
log "Reloading nginx configuration..."
sudo nginx -t && sudo systemctl reload nginx

# Final health check
log "Running final health checks..."
if curl -f https://api.web.com/health > /dev/null 2>&1 && curl -f https://web.com > /dev/null 2>&1; then
    log "🎉 Deployment successful!"
else
    error "❌ Deployment failed - health checks failed"
    exit 1
fi

log "Deployment completed at $(date)"
```

```bash
# Make executable
chmod +x /home/$USER/deploy.sh

# Test deployment
./deploy.sh
```

### 12.2 Zero-Downtime Deployment untuk Frontend
```bash
# Create blue-green deployment script untuk frontend
nano /home/$USER/deploy-frontend-zero-downtime.sh
```

```bash
#!/bin/bash

# Blue-Green deployment untuk Next.js
CURRENT_PORT=$(pm2 show frontend | grep -o 'port.*[0-9]*' | grep -o '[0-9]*')
NEW_PORT=$((CURRENT_PORT == 3000 ? 3001 : 3000))

echo "Current port: $CURRENT_PORT, New port: $NEW_PORT"

# Update code
cd /home/$USER/frontend
git pull origin main
npm install
npm run build

# Start new instance
PORT=$NEW_PORT pm2 start npm --name "frontend-new" -- start

# Wait for new instance to be ready
sleep 10

# Test new instance
if curl -f http://localhost:$NEW_PORT > /dev/null 2>&1; then
    echo "New instance healthy, switching traffic..."
    
    # Update nginx to point to new port
    sudo sed -i "s/127.0.0.1:$CURRENT_PORT/127.0.0.1:$NEW_PORT/g" /etc/nginx/sites-available/web.com
    sudo nginx -t && sudo systemctl reload nginx
    
    # Stop old instance
    pm2 delete frontend
    pm2 restart frontend-new --name frontend
    pm2 delete frontend-new
    
    echo "Zero-downtime deployment completed!"
else
    echo "New instance failed, rolling back..."
    pm2 delete frontend-new
    exit 1
fi
```

## 📈 **SCALING STRATEGIES**

### 13.1 Database Scaling
```bash
# Setup MySQL Master-Slave (jika diperlukan)
# Create additional database container untuk read replica
```

```yaml
# Add to docker-compose.yml
  mysql-slave:
    image: mysql:8.0
    container_name: vla-mysql-slave
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: ${DB_NAME}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASS}
    volumes:
      - mysql_slave_data:/var/lib/mysql
      - ./database/slave.cnf:/etc/mysql/conf.d/slave.cnf:ro
    ports:
      - "3307:3306"
    restart: unless-stopped
    depends_on:
      - db
```

### 13.2 Load Balancer Setup (Nginx)
```nginx
# Setup upstream untuk multiple backend instances
upstream backend_servers {
    least_conn;
    server 127.0.0.1:8080 weight=3;
    server 127.0.0.1:8081 weight=2;
    # server 127.0.0.1:8082 weight=1;
}

upstream frontend_servers {
    least_conn;
    server 127.0.0.1:3000 weight=3;
    server 127.0.0.1:3001 weight=2;
}

# Update proxy_pass
location / {
    proxy_pass http://backend_servers;
    # ... existing config
}
```

### 13.3 CDN Integration
```bash
# Setup CDN untuk static assets (jika menggunakan Cloudflare)
# Buat subdomain untuk assets: assets.web.com
# Upload static files ke CDN atau S3
```

## 🔍 **MONITORING & ANALYTICS**

### 14.1 Application Performance Monitoring (APM)
```bash
# Install Prometheus dan Grafana (optional)
mkdir -p /home/$USER/monitoring
cd /home/$USER/monitoring

# Create docker-compose untuk monitoring
nano docker-compose.monitoring.yml
```

```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
    restart: unless-stopped

  grafana:
    image: grafana/grafana
    container_name: grafana
    ports:
      - "3005:3000"
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=your_grafana_password
    restart: unless-stopped

volumes:
  prometheus_data:
  grafana_data:
```

### 14.2 Error Tracking dengan Sentry (Optional)
```bash
# Tambahkan Sentry ke frontend Next.js
cd /home/$USER/frontend
npm install @sentry/nextjs
```

## 💰 **COST OPTIMIZATION**

### 15.1 Resource Usage Optimization
```bash
# Monitor resource usage
echo "=== CPU Usage ===" && top -bn1 | grep load
echo "=== Memory Usage ===" && free -h
echo "=== Disk Usage ===" && df -h
echo "=== Network Usage ===" && iostat -n 1 1

# Docker resource limits
# Add to docker-compose.yml
services:
  app:
    # ... existing config
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 15.2 Log Management
```bash
# Setup log rotation untuk semua services
sudo nano /etc/logrotate.d/docker-containers
```

```bash
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    size=1M
    missingok
    delaycompress
    copytruncate
}
```

## 🎓 **BEST PRACTICES SUMMARY**

### Development:
1. **Environment Separation**: Gunakan environment variables berbeda untuk dev/staging/prod
2. **Git Workflow**: Gunakan feature branches dan pull requests
3. **Code Quality**: Implement linting, testing, dan code reviews
4. **Database Migrations**: Selalu test migrations di staging dulu

### Security:
1. **Regular Updates**: Update system dan dependencies secara berkala
2. **Backup Strategy**: Automated daily backups dengan offsite storage
3. **Access Control**: Minimal privileges principle
4. **Monitoring**: Real-time monitoring untuk security events

### Performance:
1. **Caching Strategy**: Implement multiple layers of caching
2. **Database Optimization**: Index optimization dan query analysis
3. **CDN Usage**: Static assets delivery via CDN
4. **Monitoring**: Continuous performance monitoring

### Reliability:
1. **Health Checks**: Comprehensive health monitoring
2. **Graceful Degradation**: Handle service failures gracefully
3. **Circuit Breakers**: Implement circuit breaker pattern
4. **Disaster Recovery**: Regular disaster recovery testing

---

## 📋 **FINAL DEPLOYMENT CHECKLIST**

### Pre-Production:
- [ ] Load testing completed
- [ ] Security audit passed
- [ ] Backup/restore procedures tested
- [ ] Monitoring dashboards configured
- [ ] SSL certificates valid for 90+ days
- [ ] DNS configuration verified
- [ ] CDN/Cloudflare configured (if using)

### Go-Live:
- [ ] Database migrations applied
- [ ] Environment variables set correctly
- [ ] All services healthy
- [ ] SSL certificates applied
- [ ] Domain routing working
- [ ] API endpoints responding correctly
- [ ] Frontend loading properly
- [ ] Mobile responsiveness verified

### Post-Production:
- [ ] Monitor error rates first 24 hours
- [ ] Verify backup jobs running
- [ ] Check performance metrics
- [ ] Validate security headers
- [ ] Test disaster recovery procedures
- [ ] Document any issues and solutions

---

**🎉 Selamat! Aplikasi Anda sudah siap production dengan arsitektur yang scalable dan maintainable.**

**Untuk support dan maintenance, pastikan untuk:**
- Monitor logs secara berkala
- Update dependencies dan security patches
- Review performance metrics mingguan
- Test backup dan disaster recovery bulanan
- Scale resources sesuai traffic growth

**Happy Deploying! 🚀**