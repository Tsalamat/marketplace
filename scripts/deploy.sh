#!/usr/bin/env bash
set -euo pipefail

DOMAIN="404tears.kz"
PRIMARY_DOMAIN="www.$DOMAIN"
EMAIL="${CERTBOT_EMAIL:-admin@404tears.kz}"
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PROJECT_NAME="${COMPOSE_PROJECT_NAME:-student-marketplace}"
LETSENCRYPT_VOLUME="${PROJECT_NAME}_letsencrypt"
CERTBOT_WEBROOT_VOLUME="${PROJECT_NAME}_certbot_webroot"

cd "$PROJECT_DIR"

# ── 1. Check .env ─────────────────────────────────────────────────────────────
if [[ ! -f .env ]]; then
  echo "[INFO] .env not found. Generating production defaults..."
  umask 077
  cat > .env <<EOF
APP_ENV=production
APP_PORT=8080
FRONTEND_URL=https://$PRIMARY_DOMAIN,https://$DOMAIN

DB_NAME=studentmarketplace
DB_USER=smuser
DB_PASSWORD=$(openssl rand -hex 24)

REDIS_PASSWORD=$(openssl rand -hex 24)

JWT_ACCESS_SECRET=$(openssl rand -hex 32)
JWT_REFRESH_SECRET=$(openssl rand -hex 32)
JWT_ACCESS_EXPIRE=15m
JWT_REFRESH_EXPIRE=720h

MINIO_USER=minioadmin
MINIO_PASSWORD=$(openssl rand -hex 24)
MINIO_BUCKET=student-marketplace

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=

GOOGLE_CLIENT_ID_WEB=
GOOGLE_CLIENT_SECRET_WEB=
GOOGLE_CALLBACK_URL=https://$PRIMARY_DOMAIN/api/v1/auth/google/callback

MINIO_PUBLIC_URL=https://$PRIMARY_DOMAIN/minio
VITE_API_URL=https://$PRIMARY_DOMAIN
VITE_WS_URL=wss://$PRIMARY_DOMAIN

PLATFORM_FEE_PERCENT=10
CERTBOT_EMAIL=$EMAIL
EOF
fi

set_env() {
  local key="$1"
  local value="$2"
  local escaped
  escaped="$(printf '%s\n' "$value" | sed 's/[\/&]/\\&/g')"
  if grep -q "^${key}=" .env; then
    sed -i "s/^${key}=.*/${key}=${escaped}/" .env
  else
    printf '%s=%s\n' "$key" "$value" >> .env
  fi
}

set_env FRONTEND_URL "https://$PRIMARY_DOMAIN,https://$DOMAIN"
set_env GOOGLE_CALLBACK_URL "https://$PRIMARY_DOMAIN/api/v1/auth/google/callback"
set_env MINIO_PUBLIC_URL "https://$PRIMARY_DOMAIN/minio"
set_env VITE_API_URL "https://$PRIMARY_DOMAIN"
set_env VITE_WS_URL "wss://$PRIMARY_DOMAIN"

# shellcheck disable=SC1091
source .env
EMAIL="${CERTBOT_EMAIL:-$EMAIL}"

# ── 2. Pull images and build ──────────────────────────────────────────────────
echo "[1/5] Building images..."
docker compose pull --ignore-buildable
docker compose build --parallel

# ── 3. Start HTTP-only nginx for ACME challenge ───────────────────────────────
echo "[2/5] Starting nginx (HTTP-only) for Let's Encrypt challenge..."
docker volume create "$LETSENCRYPT_VOLUME" >/dev/null
docker volume create "$CERTBOT_WEBROOT_VOLUME" >/dev/null

docker compose stop nginx certbot >/dev/null 2>&1 || true
docker rm -f sm_nginx_tmp >/dev/null 2>&1 || true

docker run --rm -d \
  --name sm_nginx_tmp \
  -p 80:80 \
  -v "$PROJECT_DIR/nginx/acme.conf:/etc/nginx/nginx.conf:ro" \
  -v "$CERTBOT_WEBROOT_VOLUME:/var/www/certbot" \
  nginx:alpine
cleanup_tmp_nginx() { docker stop sm_nginx_tmp >/dev/null 2>&1 || true; }
trap cleanup_tmp_nginx EXIT

echo "[3/5] Obtaining/updating SSL certificate for $DOMAIN and $PRIMARY_DOMAIN..."
docker run --rm \
  -v "$LETSENCRYPT_VOLUME:/etc/letsencrypt" \
  -v "$CERTBOT_WEBROOT_VOLUME:/var/www/certbot" \
  certbot/certbot certonly \
    --webroot \
    --webroot-path /var/www/certbot \
    --email "$EMAIL" \
    --agree-tos \
    --no-eff-email \
    --non-interactive \
    --cert-name "$DOMAIN" \
    --expand \
    -d "$DOMAIN" \
    -d "$PRIMARY_DOMAIN"

docker stop sm_nginx_tmp 2>/dev/null || true
trap - EXIT

# ── 4. Start all services ────────────────────────────────────────────────────
echo "[4/5] Starting all services..."
docker compose up -d

# ── 5. Health check ───────────────────────────────────────────────────────────
echo "[5/5] Waiting for backend health check..."
for i in {1..30}; do
  if curl -sf "https://$PRIMARY_DOMAIN/health" > /dev/null 2>&1; then
    echo ""
    echo "✓ Deployed successfully: https://$PRIMARY_DOMAIN"
    exit 0
  fi
  printf "."
  sleep 3
done

echo ""
echo "[WARN] Health check timed out. Check logs: docker compose logs backend"
