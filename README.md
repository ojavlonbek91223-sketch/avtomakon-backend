# AvtoMakon Backend

Go 1.22 + Fiber v2 + PostgreSQL (PostGIS) + Redis + MinIO bilan REST API va WebSocket xizmati.

## Struktura

```
avtomakon-backend/
├── cmd/
│   └── api/
│       └── main.go               # Entry point
├── internal/
│   ├── config/                   # Environment, konfiguratsiya
│   ├── server/                   # Fiber app, routelar, middleware bog'lash
│   ├── middleware/               # Auth, CORS, rate limit, logging
│   ├── handler/                  # HTTP handler'lar (controller analog)
│   ├── service/                  # Biznes logikasi
│   ├── repository/postgres/      # DB so'rovlar (pgx)
│   ├── domain/                   # Modellar (User, Post, Business...)
│   ├── websocket/                # WS hub, chat eventlari
│   └── pkg/
│       ├── jwt/                  # JWT manager
│       ├── hash/                 # Argon2id parol hash
│       └── validator/            # Request validatsiya
├── migrations/                   # SQL migratsiyalar (golang-migrate)
├── scripts/                      # Yordamchi skriptlar
├── docs/
├── Dockerfile
├── Makefile
├── go.mod
└── .env.example
```

## Talab qilinadigan dasturlar

- [Go 1.22+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

## Boshlash

```bash
# 1. Bog'liqliklarni o'rnatish
go mod tidy

# 2. .env yaratish
cp .env.example .env
# Keyin .env faylida JWT_SECRET ni almashtiring (kamida 32 belgi)

# 3. PostgreSQL + Redis + MinIO ishga tushirish
cd ..
docker-compose up -d

# 4. Migratsiyalarni qo'llash
cd avtomakon-backend
make migrate-up
# yoki: migrate -path ./migrations -database "postgres://avtomakon:avtomakon_dev_password@localhost:5432/avtomakon?sslmode=disable" up

# 5. API'ni ishga tushirish
make run
# yoki: go run ./cmd/api
```

API `http://localhost:8000` da ishga tushadi.

Tekshirish:
```bash
curl http://localhost:8000/health
```

## Loyihalash tamoyillari

### Qatlamlar

```
handler → service → repository → database
   ▲          │
   └──────────┴── domain (model'lar har joyda ishlatiladi)
```

- **handler** — HTTP so'rov/javob, validatsiya. Ichida biznes logika **yo'q**.
- **service** — biznes qoidalar, tranzaksiyalar, boshqa servislarni chaqirish.
- **repository** — faqat DB. SQL so'rovlar shu yerda.
- **domain** — sof modellar, biror tashqi bog'liqliksiz.

### Xavfsizlik

- ✅ Argon2id parol heshlash (`pkg/hash`)
- ✅ JWT access (15 daq) + refresh token (alohida jadval)
- ✅ Rate limiting (`middleware`)
- ✅ CORS sozlamasi
- ✅ pgx prepared statements (SQL injection bloklangan)
- ✅ Audit log jadvali (har bir muhim amal)
- ✅ Soft delete (tasodifiy o'chirishdan himoya)
- 🚧 TODO: 2FA (TOTP), Cloudflare WAF integratsiyasi, Vault sirlar

### Real-time

WebSocket `/ws` endpoint. Chat, like, online indikator uchun. Redis pub/sub bilan multi-instance.

### Saqlash

Rasm/video MinIO'ga yuklanadi (S3-compatible). Production'da to'g'ridan Cloudflare R2 yoki AWS S3'ga o'tish mumkin (kod o'zgarmaydi).

## API hujjati

Hozircha `internal/server/routes.go` da endpoint ro'yxati bor. Keyin Swagger/OpenAPI yoziladi.

## Migratsiyalar

Yangi migratsiya yaratish:
```bash
make migrate-create
# Migration nomi: add_user_status
```

Bekor qilish:
```bash
make migrate-down
```

## Test

```bash
go test ./...
go test -v -race -cover ./...
```

## Production deploy

UZINFOCOM/EPAM serverda Docker bilan:
```bash
docker build -t avtomakon-api .
docker run -d --env-file .env -p 8000:8000 avtomakon-api
```

Cloudflare oldidan reverse proxy (Nginx) va SSL.
