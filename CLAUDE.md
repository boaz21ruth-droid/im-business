# im-business — Go Backend

Custom Go service that handles auth, user profiles, and bridges the Flutter app to OpenIM.

> **Upstream dependencies (DO NOT MODIFY):** `../open-im-server/`, `../openim-sdk-core/`

---

## File Structure

```
im-business/
├── cmd/server/main.go              ← entry point
├── internal/
│   ├── config/config.go            ← Viper config loading
│   ├── db/db.go                    ← GORM postgres + redis init
│   ├── model/user.go               ← User GORM model (AutoMigrate on startup)
│   ├── repo/user.go                ← user repository (CRUD, search)
│   ├── service/
│   │   ├── openim.go               ← OpenIM admin API client (token cached 23h)
│   │   ├── account.go              ← register, login, reset/change password
│   │   └── user.go                 ← profile, search
│   ├── handler/
│   │   ├── router.go               ← Gin router, all routes registered here
│   │   ├── account.go              ← HTTP handlers for /account/*
│   │   └── user.go                 ← HTTP handlers for /user/*
│   └── middleware/
│       ├── auth.go                 ← JWT auth middleware (reads "token" header)
│       └── logger.go               ← Zap request logger
├── pkg/
│   ├── jwt/jwt.go                  ← HS256 sign/verify
│   ├── code/code.go                ← Redis-backed verification codes (mock: "123456")
│   └── resp/resp.go                ← unified JSON envelope {errCode, errMsg, data}
├── config/
│   └── config.yaml                 ← single source of truth (gitignored).
│                                     Docker-only values overridden via env vars
│                                     in docker-compose.yaml
├── Dockerfile
└── docker-compose.yaml
```

---

## Key Implementation Notes

**OpenIM admin client (`internal/service/openim.go`):**
- Every request MUST include `operationID` header (nanosecond timestamp string)
- Admin token: `POST /auth/get_admin_token`
- User token: `POST /auth/get_user_token` (NOT `/auth/user_token`)
- Admin token cached in-memory 23h (OpenIM issues 24h tokens)

**User ID:** Snowflake (node=1), stored as decimal string

**Account lookup:** `FindByAccount` matches `account OR email OR phone_number`

**Verification codes:** MVP stores "123456" in Redis, 5-min TTL. Key: `im:vcode:<email_or_phone>`

**errCode 1501** → client auto-logout (see `pkg/resp/resp.go` `ErrUnauthorized`)

**Password flow:** Flutter sends `MD5(raw_password)` → backend stores `bcrypt(MD5_password)`

---

## API Routes

```
GET  /healthz
GET  /app/check
POST /client_config/get

POST /account/code/send          ← unauthenticated
POST /account/code/verify
POST /account/register           ← returns {userID, imToken, chatToken}
POST /account/login              ← returns {userID, imToken, chatToken}
POST /account/password/reset     ← unauthenticated, requires code
POST /account/password/change    ← requires chatToken header

POST /user/update                ← requires chatToken
POST /user/find/full             ← requires chatToken
POST /user/search/full           ← requires chatToken
POST /friend/search              ← alias for search, requires chatToken
```

---

## Local Dev

### Prerequisites

```bash
# Create PostgreSQL user + DB (one-time)
psql postgres -c "CREATE USER im_business WITH PASSWORD 'im_business123';"
psql postgres -c "CREATE DATABASE im_business OWNER im_business;"

# OpenIM infrastructure must be running first
cd ../open-im-server && docker compose up -d
```

### Config (`config/config.yaml`)

```yaml
server:
  port: 10008
  mode: debug
postgres:
  dsn: "host=localhost user=im_business password=im_business123 dbname=im_business port=5432 sslmode=disable TimeZone=Asia/Shanghai"
redis:
  addr: "localhost:16379"   # OpenIM Redis exposed on 16379 on host
  password: "openIM123"
  db: 5                     # db 0–4 used by OpenIM
jwt:
  secret: "CHANGE-THIS-IN-PRODUCTION-32CHARS"
  expire_hours: 168
openim:
  api_url: "http://localhost:10002"
  admin_user_id: "imAdmin"
  secret: "openIM123"
```

### Run

```bash
go run cmd/server/main.go -config config/config.yaml
```

### Smoke Test

```bash
curl http://localhost:10008/healthz

curl -s -X POST http://localhost:10008/account/code/send \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","usedFor":1}'

# MD5("test") = 098f6bcd4621d373cade4e832627b4f6
curl -s -X POST http://localhost:10008/account/register \
  -H "Content-Type: application/json" \
  -d '{"verifyCode":"123456","platform":2,"user":{"nickname":"Alice","email":"alice@example.com","password":"098f6bcd4621d373cade4e832627b4f6","gender":1}}'

curl -s -X POST http://localhost:10008/account/login \
  -H "Content-Type: application/json" \
  -d '{"account":"alice@example.com","password":"098f6bcd4621d373cade4e832627b4f6","platform":2}'
```

---

## Docker

```bash
# Joins the 'openim' docker network; mounts config/config.yaml and overrides
# container hostnames + mode + webhook URLs via env vars (see docker-compose.yaml).
docker compose up -d --build
```

Container hostnames overridden via env in docker-compose.yaml:
`postgres-business`, `redis:6379`, `http://openim-server:10002`.
