# Tchat - Self-Hosted Chat Application

A chat application with web frontend and WebSocket backend, all served from a single domain using Docker Compose and Caddy.

## Architecture

```
haadi-tunnel.online
    ↓
Caddy (Reverse Proxy)
    ├─ /ws → Go Server (WebSocket)
    ├─ /chatter-count → Go Server (REST API)
    └─ /* → Web App (React - Static Files)
```

## Quick Start

### 1. Build and Start Services

```bash
docker-compose up -d
```

This will:
- Build the React web app
- Build and start the Go server
- Start Caddy reverse proxy

### 2. Check Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f caddy
docker-compose logs -f server
docker-compose logs -f web
```

### 3. Stop Services

```bash
docker-compose down
```

## DNS Configuration

### On Namecheap:

1. Go to **Domain List** → **Manage** `haadi-tunnel.online`
2. Click **Advanced DNS** tab
3. Add **A Record**:
   - **Host**: `@`
   - **Value**: `75.19.28.10` (your home server IP)
   - **TTL**: Automatic
4. (Optional) Add **A Record** for `www` pointing to the same IP

### Port Forwarding

Make sure your router forwards:
- **Port 80** → Your server (for HTTP/Let's Encrypt)
- **Port 443** → Your server (for HTTPS)

## Rebuilding the Web App

If you make changes to the web app:

```bash
# Rebuild just the web service
docker-compose up -d --build web

# Or rebuild everything
docker-compose up -d --build
```

## Services

- **web**: Builds the React app and stores it in a volume
- **server**: Go WebSocket server on port 9002
- **caddy**: Reverse proxy serving everything on ports 80/443

## Troubleshooting

### Web app not loading
- Check web service logs: `docker-compose logs web`
- Verify build completed: `docker-compose exec web ls -la /output`

### WebSocket not connecting
- Check server logs: `docker-compose logs server`
- Verify Caddy is proxying correctly: `docker-compose logs caddy`

### SSL certificate issues
- Ensure ports 80 and 443 are open and forwarded
- Check Caddy logs for Let's Encrypt errors
- DNS must be pointing to your server IP

