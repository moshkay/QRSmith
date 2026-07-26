# QRForge

A modular Go service: a **QR Code Generator** and **URL Shortener** in one.
Generate beautiful, scannable QR codes for 9 content types, style them, and
download as **PNG or SVG** — plus shorten long URLs with click tracking.

It ships with a clean web UI and a JSON API following the standard Dojah
response envelope.

## Features

- **9 QR content types**: URL, Text, WiFi, Business card, Contact, Email, Phone,
  SMS, and Location — each encoded in the correct format (vCard, `WIFI:`,
  `mailto:`, `tel:`, `SMSTO:`, `geo:`, etc.).
- **PNG & SVG downloads**: crisp raster PNGs for screens, scalable SVGs for print.
- **Beautify**: foreground/background colors, linear gradients, square or
  rounded-dot modules, adjustable border (quiet zone), output size, and error
  correction level.
- **Center logo** embedding (PNG/JPEG). Automatically bumps error correction to
  the highest level so the code stays scannable.
- **Preset themes** (Classic, Midnight, Ocean, Sunset, Forest, Neon, Grape).
- **Contrast guard**: rejects color combinations too low-contrast to scan.
- **URL shortener** with `/s/{code}` redirects and click tracking. In the
  generator, "shorten first" produces a smaller, trackable code.
- Rate limiting, security headers, graceful shutdown, and a `/healthz` probe.

## Architecture

The app is split into focused, independently testable services:

| Package              | Responsibility                                                             |
| -------------------- | -------------------------------------------------------------------------- |
| `internal/qr`        | Core rendering engine: matrix → styled PNG **and SVG** (colors, gradient). |
| `internal/beautify`  | Styling service: presets, validation, contrast checks, logo decode.        |
| `internal/qrcontent` | Builds the encoded payload for each of the 9 content types.                |
| `internal/shortener` | In-memory URL shortener with click tracking.                               |
| `internal/server`    | HTTP API, handlers, middleware (rate limit, logging, security).            |
| `internal/config`    | Environment-based configuration.                                           |
| `web/`               | Embedded single-page frontend (served via `go:embed`).                     |
| `cmd/server`         | Entrypoint with graceful shutdown.                                         |

## Run locally

```bash
go run ./cmd/server
# open http://localhost:8080
```

## Configuration

All settings come from environment variables (with sane defaults):

| Variable            | Default | Description                          |
| ------------------- | ------- | ------------------------------------ |
| `PORT`              | `8080`  | HTTP listen port                     |
| `READ_TIMEOUT`      | `15s`   | HTTP read timeout                    |
| `WRITE_TIMEOUT`     | `30s`   | HTTP write timeout                   |
| `SHUTDOWN_TIMEOUT`  | `10s`   | Graceful shutdown timeout            |
| `MAX_CONTENT_BYTES` | `2048`  | Max QR content length                |
| `MAX_LOGO_BYTES`    | `2097152` | Max decoded logo size (2 MiB)      |

## API

All JSON responses use the standard envelope:

```json
{ "entity": { }, "error": null, "status": true }
```

### `GET /healthz`
Liveness/readiness probe.

### `GET /api/v1/qr-presets`
Lists the built-in beautify presets.

### `POST /api/v1/qr-codes`
Generates a QR code and returns it as a base64 data URL.

Provide the payload in one of two ways:

- **Raw** — set `content` to the exact string to encode.
- **Structured** — set `type` (one of the 9 content types) plus a `data` object
  of fields; the server formats the correct payload.

```json
{
  "type": "wifi",
  "data": { "ssid": "Cafe", "password": "hunter2", "auth": "WPA" },
  "format": "svg",
  "preset": "ocean",
  "shape": "circle",
  "size": 512,
  "gradient": {
    "angle": 45,
    "stops": [
      { "offset": 0, "color": "#06B6D4" },
      { "offset": 1, "color": "#2563EB" }
    ]
  },
  "logoBase64": "iVBORw0KGgo..."
}
```

`format` is `"png"` (default) or `"svg"`. Style fields (`foreground`,
`background`, `transparent`, `shape`, `errorCorrection`, `size`, `borderWidth`,
`gradient`, `logoBase64`) and `preset` are all optional; explicit fields
override preset defaults.

Content types and their `data` fields:

| `type`     | Fields                                                                |
| ---------- | --------------------------------------------------------------------- |
| `url`      | `url`                                                                 |
| `text`     | `text`                                                                |
| `wifi`     | `ssid`, `password`, `auth` (WPA/WEP/nopass), `hidden`                 |
| `business` | `company`, `name`, `title`, `phone`, `email`, `website`, `address`    |
| `contact`  | `firstName`, `lastName`, `phone`, `email`, `organization`, `title`, `website` |
| `email`    | `to`, `subject`, `body`                                               |
| `phone`    | `number`                                                              |
| `sms`      | `number`, `message`                                                   |
| `location` | `latitude`, `longitude`                                               |

Response `entity`:

```json
{ "image": "data:image/svg+xml;base64,...", "format": "svg", "width": 512 }
```

### `POST /api/v1/qr-codes/download`
Same request body as above, but responds with the raw image (`image/png` or
`image/svg+xml`) and a `Content-Disposition: attachment` header.

```bash
curl -X POST http://localhost:8080/api/v1/qr-codes/download \
  -H 'Content-Type: application/json' \
  -d '{"content":"https://dojah.io","preset":"ocean","format":"svg"}' \
  -o qrcode.svg
```

### URL shortener

- `POST /api/v1/short-links` — body `{ "url": "https://…" }` → creates a short
  link. Response `entity`: `{ code, shortUrl, url, clicks, createdAt }`.
- `GET /api/v1/short-links/{code}` — returns click stats for a code.
- `GET /s/{code}` — 302-redirects to the target URL and increments `clicks`.

```bash
curl -X POST http://localhost:8080/api/v1/short-links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/a/very/long/path"}'
```

> Short links are held in memory and do not survive a restart.

## Docker

```bash
docker build -t qrforge .
docker run -p 8080:8080 qrforge
```

The image is a multi-stage build, runs as a non-root user, and defines a
`HEALTHCHECK` against `/healthz`.

## Deploy to EC2 (CI/CD)

`/.github/workflows/deploy.yml` builds, tests, and deploys on every push to
`master` (and via manual "Run workflow"). It builds the Docker image on the
GitHub runner, ships it (plus the compose file) to the EC2 host over SSH, then
runs `docker compose up -d`. No container registry is required.

**Topology:** the existing **host nginx** (already serving your other app on
ports 80/443) terminates public traffic and reverse-proxies `qr.moshcore.com.ng`
to the `qrforge` container, which binds to **`127.0.0.1:8090`** only. QRForge is
never exposed to the public internet directly — all traffic goes through the
host nginx. This shares one reverse proxy across every site on the box.

### One-time setup

1. **Provision the EC2 host** — the host nginx already owns ports 80/443, so no
   new security-group ports are needed. Install Docker + the compose plugin:

   ```bash
   ssh ec2-user@<host> 'bash -s' < scripts/ec2-bootstrap.sh
   ```

2. **Add GitHub repository secrets** (Settings → Secrets and variables → Actions):

   | Secret        | Description                                                        |
   | ------------- | ------------------------------------------------------------------ |
   | `EC2_HOST`    | Public IP or DNS of the EC2 instance                               |
   | `EC2_USER`    | SSH user (e.g. `ec2-user` on Amazon Linux, `ubuntu` on Ubuntu)     |
   | `EC2_SSH_KEY` | **Private** SSH key (PEM contents) authorized on the instance      |

   Secrets are never hardcoded — the workflow reads them from GitHub Secrets and
   removes the deploy key from the runner after each run.

### Deploy

```bash
git push origin master      # triggers build → test → deploy
```

The app container listens on `127.0.0.1:8090`; the pipeline polls that port's
`/healthz` to confirm the rollout succeeded. Public access is served by the host
nginx vhost (below).

### Wire up the host nginx vhost + HTTPS

QRForge ships a ready-made host vhost at **`nginx/qr.moshcore.com.ng.conf`**. It
proxies `qr.moshcore.com.ng` → `127.0.0.1:8090`. Because your host nginx already
owns 80/443 (for your other app), you add this as one more site — no second
proxy, no port conflict.

1. Point the DNS **A record** for `qr.moshcore.com.ng` at the EC2 public IP.
2. Install the vhost on the host nginx (use whatever layout your other site
   uses — `conf.d` or `sites-available`):

   ```bash
   sudo cp ~/qrforge-src/nginx/qr.moshcore.com.ng.conf \
     /etc/nginx/conf.d/qr.moshcore.com.ng.conf
   sudo nginx -t && sudo systemctl reload nginx
   ```

   (The compose file doesn't ship this file to the host — copy it from your repo
   checkout, or paste its contents.)

3. Obtain TLS with the **host certbot** (same tool as your other site). The
   nginx plugin rewrites the vhost to add the 443 block + HTTP→HTTPS redirect:

   ```bash
   sudo certbot --nginx -d qr.moshcore.com.ng
   ```

   `https://qr.moshcore.com.ng` is then live, and certbot's systemd timer
   renews it automatically alongside your other certificates.

## Notes on security & privacy

- QR content may be sensitive (links, tokens); it is **never logged** — request
  logging records method, path, status, and duration only.
- Logos are validated for format (PNG/JPEG) and size before processing.
- Errors return generic messages with stable codes; no stack traces leak.
```
