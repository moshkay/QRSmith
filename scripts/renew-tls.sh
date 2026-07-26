#!/usr/bin/env bash
# Renew the Let's Encrypt certificate and reload nginx.
# Run from the deploy directory (e.g. ~/qrforge). Safe to run on a schedule;
# certbot only renews when the certificate is near expiry.
#
# Example cron entry (twice daily, as the deploy user):
#   0 3,15 * * * cd ~/qrforge && ./scripts/renew-tls.sh >> ~/qrforge/renew.log 2>&1
set -euo pipefail

docker compose --profile tools run --rm certbot renew
docker compose exec nginx nginx -s reload
echo "==> Renewal check complete"
