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
make dev     # Go API on :18080 + Vite on http://localhost:5173 (login admin / admin), Ctrl+C stops both
make test    # go test ./...
make build   # production binary ./netradock with the UI embedded
```

Override with env or `.env`: `NETRADOCK_PORT=19000 WEB_PORT=5174 make dev`. The API runs via `go run`, so restart `make dev` after Go changes; Vite hot-reloads the UI.
