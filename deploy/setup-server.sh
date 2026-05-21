#!/usr/bin/env bash
# AvtoMakon — UZINFOCOM/Ubuntu serverni BIR MARTA tayyorlash skripti.
# Ubuntu 22.04 / 24.04 (amd64) uchun. root yoki sudo bilan ishga tushiring:
#   sudo bash setup-server.sh
set -euo pipefail

GO_VERSION="1.23.4"

echo ">>> [1/8] Paketlarni yangilash..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -y && apt-get upgrade -y
apt-get install -y curl wget ca-certificates gnupg lsb-release ufw apt-transport-https

echo ">>> [2/8] PostgreSQL + PostGIS..."
apt-get install -y postgresql postgresql-contrib postgis || true
# PostGIS kengaytma versiyasiga mos paket (xato bo'lsa o'tkazib yuboriladi)
PG_VER="$(ls /usr/lib/postgresql 2>/dev/null | sort -n | tail -1 || true)"
[ -n "${PG_VER}" ] && apt-get install -y "postgresql-${PG_VER}-postgis-3" || true

echo ">>> [3/8] Go ${GO_VERSION}..."
wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tgz
rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh

echo ">>> [4/8] golang-migrate..."
curl -fsSL "https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz" \
	| tar xz -C /usr/local/bin migrate
chmod +x /usr/local/bin/migrate

echo ">>> [5/8] MinIO..."
wget -q "https://dl.min.io/server/minio/release/linux-amd64/minio" -O /usr/local/bin/minio
chmod +x /usr/local/bin/minio
id minio-user &>/dev/null || useradd -r minio-user -s /sbin/nologin
mkdir -p /var/lib/minio && chown -R minio-user:minio-user /var/lib/minio

echo ">>> [6/8] Caddy (avtomatik HTTPS)..."
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
	| gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
	> /etc/apt/sources.list.d/caddy-stable.list
apt-get update -y && apt-get install -y caddy

echo ">>> [7/8] avtomakon foydalanuvchi + papka..."
id avtomakon &>/dev/null || useradd -r -m -d /opt/avtomakon -s /bin/bash avtomakon
mkdir -p /opt/avtomakon && chown -R avtomakon:avtomakon /opt/avtomakon

echo ">>> [8/8] Firewall (faqat SSH/HTTP/HTTPS)..."
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo ""
echo ">>> TAYYOR. Keyingi qadamlar (DEPLOY runbook'da):"
echo "    1) DB yaratish + PostGIS kengaytma"
echo "    2) /opt/avtomakon/.env to'ldirish (.env.production.example dan)"
echo "    3) minio.service va avtomakon-api.service o'rnatish + ishga tushirish"
echo "    4) Caddyfile /etc/caddy/Caddyfile ga qo'yib, domenni yozish"
echo "    5) deploy.sh bilan build + migrate + restart"
