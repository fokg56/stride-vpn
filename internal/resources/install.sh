#!/usr/bin/env bash
set -euo pipefail

# ══════════════════════════════════════════════════════════════════
# Stride VPN — server auto‑installer for Ubuntu 22.04+
# Xray-core (VLESS + WebSocket + TLS) behind Nginx on port 443
# with dynamic secret path and HTTPS camouflage
# ══════════════════════════════════════════════════════════════════

DOMAIN="${1:-}"
if [[ -z "$DOMAIN" ]]; then
  echo "Usage: $0 your-domain.com"
  exit 1
fi

# ── colours ──
RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC}  $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
err()   { echo -e "${RED}[ERR]${NC}   $*"; exit 1; }

[[ $EUID -eq 0 ]] || err "Run as root"
command -v apt-get &>/dev/null || err "apt-get required"

# ────────────────────────────────────────────────────────────────
# 1.  GENERATE SECRETS
# ────────────────────────────────────────────────────────────────
UUID=$(cat /proc/sys/kernel/random/uuid)
# short hash for path uniqueness
HASH=$(echo "$UUID$DOMAIN" | sha256sum | head -c 10)
WS_PATH="/vless-ws-${HASH}"

XRAY_PORT=10005
XRAY_VER="1.8.24"

info "Domain:              $DOMAIN"
info "Generated UUID:      $UUID"
info "Secret WS path:      $WS_PATH"
info "Xray internal port:  $XRAY_PORT"

# ────────────────────────────────────────────────────────────────
# 2.  INSTALL SYSTEM PACKAGES
# ────────────────────────────────────────────────────────────────
apt-get update -qq
apt-get install -y -qq curl wget unzip nginx certbot python3-certbot-nginx openssl

# ────────────────────────────────────────────────────────────────
# 3.  INSTALL XRAY
# ────────────────────────────────────────────────────────────────
if ! command -v xray &>/dev/null; then
  info "Installing Xray-core v${XRAY_VER} …"
  bash <(curl -fsSL https://github.com/XTLS/Xray-install/raw/main/install-release.sh) install -u root
  systemctl stop xray
  ok "Xray installed"
fi

# ────────────────────────────────────────────────────────────────
# 4.  XRAY CONFIG  (/usr/local/etc/xray/config.json)
# ────────────────────────────────────────────────────────────────
mkdir -p /usr/local/etc/xray
cat > /usr/local/etc/xray/config.json <<XRAY_EOF
{
  "log": { "loglevel": "warning" },
  "inbounds": [
    {
      "port": $XRAY_PORT,
      "listen": "127.0.0.1",
      "protocol": "vless",
      "settings": {
        "decryption": "none",
        "clients": [{"id": "$UUID"}]
      },
      "streamSettings": {
        "network": "ws",
        "wsSettings": {
          "path": "$WS_PATH"
        }
      },
      "sniffing": {
        "enabled": true,
        "destOverride": ["http","tls","quic"]
      }
    }
  ],
  "outbounds": [
    {"protocol": "freedom", "settings": {}},
    {"protocol": "blackhole", "settings": {}, "tag": "block"}
  ],
  "routing": {
    "domainStrategy": "IPIfNonMatch",
    "rules": [
      {"type": "field", "ip": ["geoip:private"], "outboundTag": "block"}
    ]
  }
}
XRAY_EOF

systemctl enable --now xray 2>/dev/null || true
systemctl restart xray
ok "Xray configured + restarted"

# ────────────────────────────────────────────────────────────────
# 5.  NGINX CONFIG  —  reverse proxy + camouflage
# ────────────────────────────────────────────────────────────────

# --- camouflage HTML (served to censors at /) ---
mkdir -p /var/www/stride
cat > /var/www/stride/index.html <<'HTML'
<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Welcome</title>
<meta name="robots" content="noindex, nofollow">
<style>body{font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#f5f5f5;color:#333}h1{font-weight:300;font-size:2.5rem}</style>
</head>
<body><h1>Welcome to nginx!</h1></body>
</html>
HTML

# --- Nginx site config ---
cat > /etc/nginx/sites-available/stride.conf <<NGINX
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $DOMAIN;

    # SSL (certbot fills these)
    ssl_certificate     /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    # ── Camouflage root ───────────────────────────────────
    # Any request that is NOT the secret WS path → real HTML
    root /var/www/stride;
    index index.html;

    location / {
        try_files \$uri \$uri/ =404;
    }

    # ── Secret WebSocket proxy ────────────────────────────
    location $WS_PATH {
        proxy_pass http://127.0.0.1:$XRAY_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$http_host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # ── Security headers ──────────────────────────────────
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;
}
NGINX

ln -sf /etc/nginx/sites-available/stride.conf /etc/nginx/sites-enabled/stride.conf
rm -f /etc/nginx/sites-enabled/default

# ────────────────────────────────────────────────────────────────
# 6.  SSL VIA LET'S ENCRYPT
# ────────────────────────────────────────────────────────────────
if [[ ! -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]]; then
  info "Obtaining SSL certificate for $DOMAIN …"
  certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos \
          --email "admin@$DOMAIN" --redirect || true
fi

nginx -t && systemctl reload nginx || err "nginx config error"
ok "Nginx configured + reloaded"

# ────────────────────────────────────────────────────────────────
# 7.  OUTPUT vless:// LINK
# ────────────────────────────────────────────────────────────────
SERVER_IP=$(curl -fsSL --max-time 5 ifconfig.me 2>/dev/null || echo "$DOMAIN")

# IMPORTANT: the vless link MUST contain the exact secret path
VLESS_LINK="vless://${UUID}@${SERVER_IP}:443?security=tls&encryption=none&type=ws&host=${DOMAIN}&path=${WS_PATH}&sni=${DOMAIN}#Stride-${DOMAIN}"

cat <<RESULT

╔══════════════════════════════════════════════════════════════════╗
║                    ✅  SERVER READY                            ║
╠══════════════════════════════════════════════════════════════════╣
║  Domain      │  $DOMAIN
║  IP          │  $SERVER_IP
║  Port        │  443 (TLS)
║  UUID        │  $UUID
║  Secret path │  $WS_PATH
║  Transport   │  WebSocket + TLS
║  Camouflage  │  ✅ real HTML served on GET /
║                                                               ║
║  ── vless:// link  (paste into client) ──                    ║
║  $VLESS_LINK
╚══════════════════════════════════════════════════════════════════╝

RESULT
