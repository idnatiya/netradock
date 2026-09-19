# Netradock

Web UI for the Docker daemon on a VPS: containers (start, stop, restart, remove, logs, live stats, shell), images (pull, remove), volumes and networks.

Go Fiber backend (layout follows [golang-clean-architecture](https://github.com/khannedy/golang-clean-architecture)), Vue frontend embedded into the binary.

## Run on a VPS

```sh
cp .env.example .env   # set NETRADOCK_USERNAME, NETRADOCK_PASSWORD, NETRADOCK_SECRET
docker compose up -d --build
```

The app listens on `127.0.0.1:8080`. Serve it through an HTTPS reverse proxy (Caddy, nginx) and set `NETRADOCK_SECURE_COOKIE=true` and `NETRADOCK_PROXY_HEADER=X-Forwarded-For`. The proxy must forward WebSocket upgrades for `/ws/` and keep the original `Host` header.

**Security:** anyone logged in has root-equivalent access to the host through the Docker socket, including a shell inside any container. Use a long password and do not expose the port publicly without TLS.

## Development

```sh
cd web && npm install && npm run build && cd ..   # go:embed needs web/dist to exist
NETRADOCK_USERNAME=admin NETRADOCK_PASSWORD=dev go run ./cmd/web
cd web && npm run dev                              # http://localhost:5173, proxies /api and /ws to :8080
```

`NETRADOCK_BACKEND=http://localhost:18080 npm run dev` points the Vite proxy at another port.

Tests: `go test ./...`
