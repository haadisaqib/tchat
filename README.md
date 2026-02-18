# Tchat - Self-Hosted Chat Application

A chat application with web frontend and WebSocket backend, served over **Cloudflare Tunnel** (no Caddy, no port forwarding). Runs on a Raspberry Pi or any machine.

## Architecture

```
haadi-tunnel.online
    ↓
Cloudflare Tunnel (cloudflared)
    ↓
Go server :8080
    ├─ /ws → WebSocket
    ├─ /chatter-count → REST API
    └─ /* → React app (static files)
```

One image: web app is built at **image build time** and baked into the server. No long-lived build container.

## Quick Start

### 1. Cloudflare Tunnel

In [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) → Access → Tunnels, configure your tunnel’s **Public Hostname** so it points to the app:

- **Service type**: HTTP
- **URL**: `server:8080` (same Docker network as the tunnel)

If the tunnel was previously pointing at `caddy:80`, change it to `server:8080`.

### 2. Build and run

```bash
# From repo root
docker compose up -d --build
```

This builds the React app and Go server into a single image, then starts the server and the tunnel. No Caddy, no zombie build container.

### 3. Logs

```bash
docker compose logs -f server
docker compose logs -f tunnel
```

### 4. Stop

```bash
docker compose down
```

## Rebuilding after changes

```bash
docker compose up -d --build
```

Rebuilds the image (web + server) and restarts. No separate “web” service.

## Local development

- **Frontend**: `cd web && npm run dev` (Vite dev server).
- **Backend**: `cd server && PORT=9002 go run .` (no `STATIC_DIR`; frontend talks to `localhost:9002`).

The app uses `localhost` to decide between dev and production URLs.

## Services

- **server**: Go app on port 8080; serves static files and `/ws`, `/chatter-count`. Built from repo root `Dockerfile` (web build + server build in one image).
- **tunnel**: Cloudflared, connects to Cloudflare and forwards traffic to `server:8080`.

## Troubleshooting

- **Site not loading**: In Zero Trust, confirm the tunnel’s HTTP service is `server:8080`.
- **WebSocket fails**: Same as above; `/ws` is served by the Go server.
- **Count not persisting**: The compose file mounts `./server/chatterCount.json`; ensure the file exists and is writable.
