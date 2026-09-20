# Netradock

Web UI for the Docker daemon on a VPS: containers (start, stop, restart, remove, logs, live stats, shell), images (pull, remove), volumes and networks.

Go Fiber backend (layout follows [golang-clean-architecture](https://github.com/khannedy/golang-clean-architecture)), Vue frontend embedded into the binary.

## Run on a VPS

```sh
cp .env.example .env   # set NETRADOCK_USERNAME, NETRADOCK_PASSWORD, NETRADOCK_SECRET
make hash              # optional: prints NETRADOCK_PASSWORD_HASH=... to paste into .env
docker compose up -d --build
```

The app listens on `127.0.0.1:8080`. Serve it through an HTTPS reverse proxy (Caddy, nginx). The proxy must forward WebSocket upgrades for `/ws/`, keep the original `Host` header, and overwrite `X-Forwarded-For` rather than pass a client-supplied one.

Set `NETRADOCK_PROXY_HEADER=X-Forwarded-For` so the login rate limit sees real client IPs. Leave `NETRADOCK_SECURE_COOKIE` at its default (`true`); setting it to `false` sends the session cookie over plain HTTP.

**Security:** anyone logged in has root-equivalent access to the host through the Docker socket, including a shell inside any container. Use a long random password, prefer `NETRADOCK_PASSWORD_HASH` (run `make hash`) over `NETRADOCK_PASSWORD`, and do not expose the port publicly without TLS.

Sessions are stateless tokens with no server-side store, so there is no way to revoke one individually. Changing `NETRADOCK_USERNAME`, the password, or `NETRADOCK_SECRET` invalidates all of them at once; live WebSocket streams re-check the session every 30 seconds and close when it stops verifying.

Run `make audit` to check Go and npm dependencies for known vulnerabilities.

## Development

```sh
make dev     # Go API on :18080 + Vite on http://localhost:5178 (login admin / admin), Ctrl+C stops both
make test    # go test ./...
make build   # production binary ./netradock with the UI embedded
```

Override with env or `.env`: `NETRADOCK_PORT=19000 WEB_PORT=5174 make dev`. The API runs via `go run`, so restart `make dev` after Go changes; Vite hot-reloads the UI.
