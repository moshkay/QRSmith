#!/usr/bin/env bash
# Obtain a Let's Encrypt certificate for QRForge and enable HTTPS.
# Run once on the EC2 host, from the deploy directory (e.g. ~/qrforge):
#
#   ./scripts/init-tls.sh you@example.com
#
# Prerequisites:
#   - DNS A record for the domain points at this host's public IP.
#   - Inbound TCP 80 AND 443 are open in the security group.
#   - The stack has been deployed at least once (docker compose up).
set -euo pipefail

DOMAIN="qr.moshcore.com.ng"
EMAIL="${1:-}"

if [ -z "$EMAIL" ]; then
  echo "usage: $0 <email-for-letsencrypt>" >&2
  exit 1
fi

echo "==> Ensuring the stack is up (HTTP must be reachable for the ACME challenge)"
docker compose up -d qrforge nginx

echo "==> Requesting certificate for ${DOMAIN}"
docker compose --profile tools run --rm certbot \
  certonly --webroot -w /var/www/certbot \
  -d "${DOMAIN}" \
  --email "${EMAIL}" --agree-tos --no-eff-email --non-interactive

echo "==> Installing HTTPS server block into the nginx_ssl volume"
docker compose exec -T nginx sh -eu -c 'cat > /etc/nginx/ssl-enabled/qrforge-ssl.conf' <<EOF
server {
    listen 443 ssl;
    listen [::]:443 ssl;
    http2 on;
    server_name ${DOMAIN};

    ssl_certificate     /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    client_max_body_size 8m;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        proxy_pass http://qrforge_backend;
        proxy_http_version 1.1;
        proxy_set_header Host              \$host;
        proxy_set_header X-Real-IP         \$remote_addr;
        proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection        "";

        proxy_connect_timeout 5s;
        proxy_send_timeout    60s;
        proxy_read_timeout    60s;
    }
}
EOF

echo "==> Validating and reloading nginx"
docker compose exec nginx nginx -t
docker compose exec nginx nginx -s reload

echo "==> Done. HTTPS is live at https://${DOMAIN}"
echo "    (HTTP on port 80 still proxies; see README to force an HTTPS redirect.)"
