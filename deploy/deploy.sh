#!/usr/bin/env bash
# AvtoMakon — kodni build qilib, migratsiya qilib, xizmatni restart qilish.
# Serverda ishga tushiriladi. Kod /opt/avtomakon/src ichida bo'lishi kerak
# (git clone yoki scp orqali).
#   sudo bash /opt/avtomakon/src/deploy/deploy.sh
set -euo pipefail

SRC="/opt/avtomakon/src"
APP_DIR="/opt/avtomakon"
GO="/usr/local/go/bin/go"

cd "${SRC}"

echo ">>> Kodni yangilash (git bo'lsa)..."
git pull --ff-only 2>/dev/null || echo "   (git yo'q — kod scp bilan yuklangan deb hisoblanadi)"

echo ">>> Binary build qilish..."
CGO_ENABLED=0 GOOS=linux "${GO}" build -ldflags="-w -s" -o "${APP_DIR}/api" ./cmd/api
chown avtomakon:avtomakon "${APP_DIR}/api"

echo ">>> Migratsiyalar..."
set -a; source "${APP_DIR}/.env"; set +a
migrate -path "${SRC}/migrations" \
	-database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
	up

echo ">>> Xizmatni restart qilish..."
systemctl restart avtomakon-api
sleep 2
systemctl --no-pager status avtomakon-api | head -n 12

echo ">>> Sog'liq tekshiruvi..."
curl -fsS http://127.0.0.1:8000/health && echo "  <- OK"
