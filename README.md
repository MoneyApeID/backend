# StoneForm Backend API

## 🚀 Production Ready Go API

StoneForm is a production-ready RESTful API for investment, banking, and user management. Built with Go, MySQL, Redis, and Docker.

### ✨ Features
- **JWT Authentication** with refresh tokens
- **Rate Limiting** and security middleware
- **Investment Products** with daily returns
- **Bank Account Management**
- **Forum & Task System**
- **Payment Integration** (LinkQu QRIS & Virtual Account deposits)
- **Production-ready** Docker deployment
- **Security hardened** with best practices

### 🏗️ Architecture
- **Backend**: Go 1.25 with Gin framework
- **Database**: MySQL 8.0 with connection pooling
- **Cache**: Redis 7 with persistence
- **Deployment**: Docker Compose with Nginx reverse proxy
- **Security**: TLS, rate limiting, input validation

## 📚 Documentation

- **[API Documentation](#api-documentation)** - Complete endpoint reference
- **[Production Deployment](how-to-run.md)** - Step-by-step deployment guide
- **[Security Recommendations](SECURITY-RECOMMENDATIONS.md)** - Security best practices
- **[Database Hardening](DATABASE-HARDENING.md)** - Database security guidelines

## 🚀 Quick Start

### Development
```bash
# Clone repository
git clone <your-repo-url>
cd stoneform-backend

# Copy environment template
cp env.example .env

# Start development environment
docker compose -f docker-compose.dev.yml up -d

# Run application locally
go run main.go
```

### Production Deployment
```bash
# Setup production environment
cp env.example .env
# Edit .env with production values

# Deploy with Docker Compose
docker compose up -d --build

# Check status
docker compose ps
docker compose logs -f
```

📖 **For detailed deployment instructions, see [how-to-run.md](how-to-run.md)**

## 🔧 Configuration

### Required Environment Variables
```bash
# Application
ENV=production
PORT=8080

# Database
DB_HOST=db
DB_USER=your_db_user
DB_PASS=your_secure_password
DB_NAME=your_database_name

# Security
JWT_SECRET=your_very_secure_jwt_secret_key_minimum_32_characters

# LinkQu Payments
LINKQU_BASE_URL=https://partner-api.linkqu.id
LINKQU_USERNAME=your_linkqu_username
LINKQU_PIN=your_linkqu_pin
LINKQU_CLIENT_ID=your_linkqu_client_id
LINKQU_CLIENT_SECRET=your_linkqu_client_secret
LINKQU_CALLBACK_PAYMENT=https://your-domain.com/api/payments/linkqu/callback

# Redis
REDIS_ADDR=redis:6379
REDIS_PASS=your_redis_password
```

📋 **See [env.example](env.example) for complete configuration options**

## 📊 API Documentation

### Base Information
- **Base URL:** `https://api.yourdomain.com/api` (production) or `http://localhost:8080/api` (development)
- **Authentication:** JWT Bearer tokens
- **Rate Limiting:** Implemented per endpoint
- **Response Format:** JSON with `success`, `message`, `data` fields

## .env Exampe

```
PORT=8080

JWT_SECRET=supersecretjwtkey
WITHDRAWAL_CHARGE_PERCENT=10.0
CRON_KEY=supersecretcronkey

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=
DB_NAME=sf

KYTAPAY_CLIENT_ID=
KYTAPAY_CLIENT_SECRET=
NOTIFY_URL=https://yourdomain.com/api/callback/payment
SUCCESS_URL=https://yourdomain.com/payment/success
FAILED_URL=https://yourdomain.com/payment/failed

LINKQU_BASE_URL=
LINKQU_USERNAME=
LINKQU_PIN=
LINKQU_CLIENT_ID=
LINKQU_CLIENT_SECRET=
LINKQU_CALLBACK_PAYMENT=https://yourdomain.com/api/payments/linkqu/callback

# Optional: full DSN (overrides DB_HOST/PORT/USER/PASS/NAME if set)
# Example for Docker: root:123456789@tcp(db:3306)/v1?charset=utf8mb4&parseTime=True&loc=Local
DB_DSN=
```

## List of Endpoints

| Method | Endpoint                              | Brief Description                       |
|--------|---------------------------------------|-----------------------------------------|
| POST   | /register                             | Register new user and returns JWT token |
| POST   | /login                                | Login, returns JWT token                |
| GET    | /ping                                 | Protected ping (JWT required)           |
| GET    | /users/info                           | Get user info (JWT required)            |
| POST   | /users/change-password                | Change password (JWT required)          |
| GET    | /products                             | List investment products                |
| POST   | /users/investments                    | Create investment (JWT required)        |
| GET    | /users/investments                    | List user investments (JWT required)    |
| GET    | /users/investments/{id}               | Get investment detail (JWT required)    |
| POST   | /users/deposits                       | Create deposit payment (JWT required)   |
| POST   | /users/withdrawal                     | Withdraw funds (JWT required)           |
| GET    | /users/bank                           | List user bank accounts (JWT required)  |
| POST   | /users/bank                           | Add bank account (JWT required)         |
| PUT    | /users/bank                           | Edit bank account (JWT required)        |
| DELETE | /users/bank                           | Delete bank account (JWT required)      |
| GET    | /bank                                 | List supported banks (JWT required)     |
| GET    | /users/task                           | List user tasks (JWT required)          |
| POST   | /users/task/submit                    | Submit task (JWT required)              |
| GET    | /users/forum                          | List forum posts (JWT required)         |
| POST   | /users/forum/submit                   | Submit forum post (JWT required)        |
| POST   | /payments/linkqu/callback             | LinkQu payment callback (no auth)       |
| POST   | /cron/daily-returns                   | Cron: process daily returns (X-CRON-KEY)|

## Endpoint Details

### Register
**POST /api/register**
- Registers a new user.
- **Headers:** `Content-Type: application/json`
- **Request Body:**
```json
{
  "name": "John Doe",
  "number": "0812000000011",
  "password": "secret12",
  "password_confirmation": "secret12",
  "referral_code": "" // optional
}
```
- **Success Response:**
```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "token": "<jwt-token>",
    "expired_at": "2025-08-28T01:23:45Z",
    "data": {
        "id": 1,
        "name": "John Doe",
        "number": "0812000000011",
        ...
    }
  }
}
```
- **Error Response:**
```json
{
  "success": false,
  "message": "number already registered"
}
```

### Login
**POST /api/login**
- Authenticates user and returns JWT token.
- **Headers:** `Content-Type: application/json`
- **Request Body:**
```json
{
  "number": "0812000000011",
  "password": "secret12"
}
```
- **Success Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "<jwt-token>",
    "expired_at": "2025-08-28T01:23:45Z",
    "data": {
        "id": 1,
        "name": "John Doe",
        "number": "0812000000011",
        ...
    }
  }
}
```
- **Error Response:**
```json
{
  "success": false,
  "message": "Invalid number or password"
}
```

### Get User Info
**GET /api/users/info**
- Returns user information.
- **Headers:** `Authorization: Bearer <token>`
- **Success Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "John Doe",
    "number": "0812000000011",
    ...
  }
}
```

### List Products
**GET /api/products**
- Lists available investment products.
- **Success Response:**
```json
{
  "success": true,
  "data": {
    "products": [
      { "id": 1, "name": "Star 1", "min": 100000, "max": 1000000 },
      { "id": 2, "name": "Star 2", "min": 1000000, "max": 10000000 },
      ...
    ]
  }
}
```

### Create Investment
**POST /api/users/investments**
- Create a new investment for the user.
- **Headers:** `Authorization: Bearer <token>`, `Content-Type: application/json`
- **Request Body:**
```json
{
  "product_id": 3,
  "amount": 5000000,
  "payment_method": "BANK",
  "payment_channel": "BCA"
}
```
- **Success Response:**
```json
{
  "success": true,
  "data": {
    "investment_id": 123,
    "payment": {
      "account_number": "1234567890",
      "expiry": "2025-09-10T12:00:00Z"
    }
  }
}
```

### Withdraw
**POST /api/users/withdrawal**
- Withdraw funds to a bank account.
- **Headers:** `Authorization: Bearer <token>`, `Content-Type: application/json`
- **Request Body:**
```json
{
  "amount": 99000,
  "bank_account_id": 1
}
```
- **Success Response:**
```json
{
  "success": true,
  "message": "Withdrawal request submitted"
}
```

### Add Bank Account
**POST /api/users/bank**
- Add a new bank account for the user.
- **Headers:** `Authorization: Bearer <token>`, `Content-Type: application/json`
- **Request Body:**
```json
{
  "bank_id": 2,
  "account_name": "John Doe",
  "account_number": "757654123611"
}
```
- **Success Response:**
```json
{
  "success": true,
  "message": "Bank account added"
}
```

### Delete Bank Account
**DELETE /api/users/bank**
- Delete a user's bank account.
- **Headers:** `Authorization: Bearer <token>`, `Content-Type: application/json`
- **Request Body:**
```json
{
  "id": 1
}
```
- **Success Response:**
```json
{
  "success": true,
  "message": "Bank account deleted"
}
```
- **Error Response:**
```json
{
  "success": false,
  "message": "Unauthorized"
}
```

### Webhook (Kytapay)
**POST /api/payments/kyta/webhook**
- Payment provider callback. No authentication required.
- **Request Body:**
```json
{
  "reference_id": "PUT_ORDER_ID_HERE",
  "status": "PAID",
  "response_code": "2000500",
  "id": "payment-123"
}
```
- **Success Response:**
```json
{
  "success": true
}
```

### Cron: Daily Returns
**POST /api/cron/daily-returns**
- Internal cron endpoint. Requires `X-CRON-KEY` header.
- **Headers:** `X-CRON-KEY: <your_cron_key>`
- **Success Response:**
```json
{
  "success": true,
  "message": "Processed daily returns"
}
```

## Rate Limiting

- **Register/Login:** 10 requests/minute/IP
- **User endpoints (read):** 120 requests/minute/user (e.g., GET /users/info, /users/investments)
- **User endpoints (write):** 60 requests/minute/user (e.g., POST/PUT/DELETE)
- **Withdraw/Transfer:** 20 requests/minute/user (if implemented separately)
- **Webhook:** 500 requests/hour/IP (no auth, sliding window, whitelisted IPs unlimited)
- **Cron:** 1000 requests/hour/IP

## Example Workflow

1. **Register**
2. **Login** (get JWT token)
3. **Get user info** (use token)
4. **List products**
5. **Create investment**
6. **Add bank account**
7. **Withdraw**

### Example
```bash
# Register
curl -X POST https://api.domain.com/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John","number":"081200000001","password":"secret12","password_confirmation":"secret12"}'

# Login
curl -X POST https://api.domain.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"number":"081200000001","password":"secret12"}'

# Use token for next requests
curl https://api.domain.com/api/users/info \
  -H "Authorization: Bearer <jwt-token>"

# List products
curl https://api.domain.com/api/products

# Create investment
curl -X POST https://api.domain.com/api/users/investments \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"product_id":3,"amount":5000000,"payment_method":"BANK","payment_channel":"BCA"}'

# Add bank account
curl -X POST https://api.domain.com/api/users/bank \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"bank_id":2,"account_name":"John Doe","account_number":"757654123611"}'

# Withdraw
curl -X POST https://api.domain.com/api/users/withdrawal \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"amount":99000,"bank_account_id":1}'
```

## Error Handling

Common error codes and responses:

| Code | Meaning                | Example Response |
|------|------------------------|------------------|
| 400  | Bad Request            | `{ "success": false, "message": "Invalid request" }` |
| 401  | Unauthorized           | `{ "success": false, "message": "Unauthorized" }` |
| 403  | Forbidden              | `{ "success": false, "message": "Forbidden" }` |
| 429  | Too Many Requests      | `{ "success": false, "message": "Too many requests. Please try again later." }` |
| 500  | Server Error           | `{ "success": false, "message": "Internal server error" }` |

All error responses are JSON with `success: false` and a `message` field.
1) Copy `.env.example` to `.env`, ensure DB settings match docker-compose.
2) Start services:
   ```bash
   docker compose up -d --build
   ```
3) Verify:
   ```bash
   docker compose ps
   docker compose logs -f app
   ```
4) Test endpoints:
   - Register:
     ```bash
     curl -X POST http://localhost:8080/api/register \
       -H "Content-Type: application/json" \
       -d '{"name":"John","number":"081200000001","password":"secret12","password_confirmation":"secret12"}'
     ```
   - Login:
     ```bash
     curl -X POST http://localhost:8080/api/login \
       -H "Content-Type: application/json" \
       -d '{"number":"081200000001","password":"secret12"}'
     ```
     Successful response:
     ```json
     {
       "success": true,
       "message": "Login successful",
       "data": {
         "token": "<jwt-token>",
         "expired_at": "2025-08-28T01:23:45Z"
       }
     }
     ```
   - Use token on next request (protected example):
     ```bash
     curl http://localhost:8080/api/ping \
       -H "Authorization: Bearer <jwt-token>"
     ```

## JWT Notes
- Token contains claims: id, name, exp.
- Expiration is 6 hours from login; after that, token is invalid.
- Middleware reads Authorization: Bearer <token>, verifies with JWT_SECRET.
- Expired token returns:
  ```json
  { "success": false, "message": "Token expired, please login again" }
  ```

### Access-token Revocation (optional Redis)

- The server supports immediate access-token revocation by storing the token's `jti` in a revocation store.
- If `REDIS_ADDR` is configured the server will use Redis as the revocation store and check the key `jwt:blacklist:<jti>` on each request. If that key is present the access token is rejected.
- If Redis is not configured, the server will fall back to checking a `revoked_tokens` DB table.
- Use the helper `utils.RevokeJTI(jti, ttl)` in logout or revoke endpoints. If Redis is used this sets a key with the provided TTL; otherwise it inserts into `revoked_tokens`.

Example (server-side):
1. Parse access token and extract `jti` and `exp`.
2. Compute `ttl := time.Until(exp)`.
3. Call `utils.RevokeJTI(jti, ttl)` to ensure the token is rejected until expiry.


## Environment
- Set `JWT_SECRET` in your `.env` (required for token signing/verification).
- Database config via `.env`: DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME (or DB_DSN).


# Stoneform Investment API Additions

This update replaces the old deposit top-up flow with direct Investments using fixed Products (Star 1, Star 2, Star 3). Payments still use Kytapay (QRIS / Virtual Account). Daily returns are processed via a cron endpoint.

## Database Migrations
Run the SQL files in the `migrations` folder in order (ensure you already created `users`, `transactions`, `banks`, etc.):
- create_products_table.sql
- seed_products.sql
- create_investments_table.sql
- 20250907_alter_investments_add_payment_fields.sql

These create and seed:
- products: Star 1/2/3 with min/max, percentage, duration
- investments: tracks user investments, daily profit, status, schedule

## Environment Variables
Add the following to your `.env`:
- KYTAPAY_BASE_URL (default: https://api.kytapay.com/v2)
- KYTAPAY_CLIENT_ID
- KYTAPAY_CLIENT_SECRET
- NOTIFY_URL  (Kytapay webhook URL -> e.g. https://yourdomain/api/payments/kyta/webhook)
- SUCCESS_URL (redirect after successful payment)
- FAILED_URL  (redirect after failed payment)
- CRON_KEY    (secret used by the cron endpoint)

## New Endpoints
- GET /api/products
  - Public. Lists active products (Star 1/2/3).

- POST /api/users/investments (protected)
  - Body: { product_id, amount, payment_method: "QRIS"|"BANK", payment_channel: "BCA"|"BRI"|"BNI"|"MANDIRI"|"PERMATA"|"BNC" (if BANK) }
  - Validates amount within product min/max, creates Kytapay payment, creates Investment (Pending) + a Transaction (Pending), and returns payment details (qr_string or account_number) and expiry.

- GET /api/users/investments (protected)
  - List user investments.

- GET /api/users/investments/{id} (protected)
  - Get single investment detail.

- POST /api/payments/kyta/webhook
  - Kytapay callback (no auth). On success, mark investment Running, set next_return_at to +24h, mark the related transaction Success, and increment user.total_invest.

- POST /api/cron/daily-returns
  - Cron endpoint protected via header: X-CRON-KEY: <CRON_KEY>
  - Processes due investments (status Running, next_return_at <= now). Credits daily profit to user balance, adds a Success transaction of type investment_profit, updates schedule and marks Completed when total_paid == duration.

## Notes
- The old deposit route is removed from the router. Payment utilities from deposit code are reused internally for investments.
- Transaction types used: "investment" for the initial top-up and "investment_profit" for daily returns.
- Daily profit formula: amount * (percentage/100) / duration, rounded to 2 decimals.
- Payment gateway details are stored with each investment for later access: payment_method, payment_channel, payment_code (qr_string or VA number), payment_link, expired_at.

## Testing
Use `test.http` for example requests:
- Register/Login and capture token.
- GET /api/products
- POST /api/users/investments (QRIS or BANK)
- Simulate webhook: POST /api/payments/kyta/webhook
- Trigger cron: POST /api/cron/daily-returns with X-CRON-KEY header.

## 🚀 Production Deployment

### Prerequisites
- VPS with Ubuntu 20.04+ or Debian 11+
- Docker & Docker Compose installed
- Domain name (optional, for HTTPS)

### Quick Deploy
```bash
# 1. Clone repository
git clone <your-repo-url>
cd stoneform-backend

# 2. Setup environment
cp env.example .env
nano .env  # Edit with production values

# 3. Deploy
docker compose up -d --build

# 4. Check status
docker compose ps
curl http://localhost:8080/health
```

### With Nginx Reverse Proxy
```bash
# Deploy with Nginx
docker compose --profile nginx up -d --build

# Setup SSL (optional)
sudo certbot certonly --standalone -d yourdomain.com
```

### Monitoring & Maintenance
```bash
# View logs
docker compose logs -f

# Backup database
docker compose exec db mysqldump -u root -p$DB_ROOT_PASSWORD $DB_NAME > backup.sql

# Update application
git pull && docker compose up -d --build
```

📖 **Complete deployment guide: [how-to-run.md](how-to-run.md)**

## 🔒 Security

This application implements production-grade security features:

- **Authentication**: JWT with refresh tokens
- **Authorization**: Role-based access control
- **Rate Limiting**: Per-endpoint and per-user limits
- **Input Validation**: Comprehensive request validation
- **Security Headers**: CORS, CSRF protection, HSTS
- **Database Security**: TLS connections, connection pooling
- **Container Security**: Non-root user, minimal base image

📋 **Security best practices: [SECURITY-RECOMMENDATIONS.md](SECURITY-RECOMMENDATIONS.md)**

## 🛠️ Development

### Local Development
```bash
# Start development environment
docker compose -f docker-compose.dev.yml up -d

# Run application locally
go run main.go

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Project Structure
```
├── controllers/     # HTTP handlers
├── models/         # Database models
├── middleware/     # HTTP middleware
├── routes/         # Route definitions
├── database/       # Database connection
├── utils/          # Utility functions
├── migrations/     # Database migrations
└── scripts/        # Development scripts
```

## 📞 Support

- **Documentation**: See individual `.md` files
- **Issues**: Create GitHub issue
- **Security**: Report to security@yourcompany.com

## 📄 License

[Your License Here]

---

**🎉 StoneForm Backend - Production Ready Go API**
