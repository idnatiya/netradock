# Netradock

<p align="center">
  <strong>A lightweight, ultra-fast, single-binary Docker control plane and real-time telemetry dashboard for your VPS.</strong>
</p>

<p align="center">
  <a href="#features"><img src="https://img.shields.io/badge/Docker-Control%20Plane-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker"></a>
  <a href="#development"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Version"></a>
  <a href="#development"><img src="https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat-square&logo=vuedotjs&logoColor=white" alt="Vue Version"></a>
  <a href="#development"><img src="https://img.shields.io/badge/Tailwind-v4-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white" alt="Tailwind CSS"></a>
  <a href="#license"><img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License"></a>
</p>

---

## 💡 Why Netradock?

Managing Docker containers on a VPS often means choosing between heavy management suites that consume hundreds of megabytes of RAM or resorting solely to SSH and terminal commands.

**Netradock** bridges this gap:
- **Ultra Lightweight:** Consumes only ~20–30 MB of RAM under typical workloads.
- **Single Static Binary:** The entire Vue 3 single-page app is compiled directly into the Go binary using `go:embed`. No Node.js runtime or external asset hosting needed.
- **Zero External Database:** Stateless architecture with zero database dependencies (no PostgreSQL, Redis, or SQLite required).
- **Real-Time Telemetry:** Native WebSocket pipelines deliver live container resource metrics (CPU & Memory), continuous log streaming, and an interactive shell terminal powered by [xterm.js](https://xtermjs.org/).
- **Production-Grade Security:** Hardened Content Security Policy, CSRF protection, strict Same-Origin validation, brute-force rate limiting, and constant-time bcrypt authentication.

---

## ✨ Features

### 📦 Container Management
- **Lifecycle Control:** Start, stop, restart, and remove containers with a single click.
- **Interactive Web Terminal:** Open an interactive shell terminal (`sh` / `bash`) directly inside any running container via WebSockets.
- **Live Log Streaming:** Real-time log monitoring with automatic auto-scroll and quick filtering.
- **Real-Time Resource Telemetry:** Live CPU and Memory usage tracking streamed continuously without polling lag.
- **Deep Inspection:** View container configuration, mounted volumes, port mappings, and environment metadata in clean JSON format.

### 🖼️ Docker Image Management
- **Inspect Images:** Browse locally stored images, tags, creation dates, and virtual sizes.
- **Pull from Registries:** Pull images from Docker Hub or public/private registries with live download feedback.
- **Cleanup & Removal:** Delete unused or dangling images to reclaim host disk space.

### 💾 Volumes & Networks
- **Volumes:** Inspect persistent storage volumes, volume drivers, and mount points; delete unused volumes.
- **Networks:** Inspect Docker bridge, host, and custom networks alongside connected container IPs.

### 📊 System Overview
- Daemon status, Docker engine version, operating system, and architecture.
- Container status counters (running, paused, stopped).

---

## ⚡ Quick Start

### Option 1: Docker Compose (Recommended)

1. Clone the repository and navigate into the project directory:
   ```sh
   git clone https://github.com/idnatiya/netradock.git
   cd netradock
   ```

2. Copy the example environment configuration:
   ```sh
   cp .env.example .env
   ```

3. Configure your credentials and secret:
   - Generate a secure password hash:
     ```sh
     make hash
     ```
     Paste the generated `NETRADOCK_PASSWORD_HASH=...` into your `.env` file.
   - Generate a 32-byte session secret:
     ```sh
     openssl rand -hex 32
     ```
     Paste the output as `NETRADOCK_SECRET=...` in your `.env` file.

4. Start Netradock using Docker Compose:
   ```sh
   docker compose up -d --build
   ```

5. Access the web interface at `http://127.0.0.1:8080`.

> [!IMPORTANT]
> By default, Netradock binds to `127.0.0.1:8080`. For security reasons, do not expose this port directly to the public internet without an HTTPS reverse proxy (such as Caddy or Nginx).

---

### Option 2: Docker CLI (`docker run`)

You can run Netradock directly with `docker run`:

```sh
docker run -d \
  --name netradock \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e NETRADOCK_USERNAME=admin \
  -e NETRADOCK_PASSWORD="your-secure-password" \
  -e NETRADOCK_SECRET="$(openssl rand -hex 32)" \
  netradock
```

---

### Option 3: Standalone Binary (Host Service)

Build and run Netradock as a native host binary:

```sh
# Prerequisites: Go 1.26+, Node.js 22+
make install
make build
```

This compiles a single standalone binary `./netradock` with the web frontend embedded. Run it directly:

```sh
NETRADOCK_USERNAME=admin \
NETRADOCK_PASSWORD="your-secure-password" \
NETRADOCK_SECRET="$(openssl rand -hex 32)" \
./netradock
```

---

## 🌐 Reverse Proxy Setup

Because Netradock handles root-equivalent Docker daemon operations, **always serve it behind an HTTPS reverse proxy**.

### Caddy (Recommended)

Caddy automatically provisions TLS certificates and handles WebSocket upgrades:

```caddy
docker.yourdomain.com {
    reverse_proxy 127.0.0.1:8080
}
```

In `.env`, set:
```ini
NETRADOCK_PROXY_HEADER=X-Forwarded-For
NETRADOCK_SECURE_COOKIE=true
```

---

### Nginx

Make sure to pass WebSocket headers (`Upgrade`, `Connection`) and configure client IP forwarding:

```nginx
server {
    listen 443 ssl http2;
    server_name docker.yourdomain.com;

    ssl_certificate     /etc/letsencrypt/live/docker.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/docker.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Forwarded headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Disable response buffering for live streaming
        proxy_buffering off;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }
}
```

In `.env`, set:
```ini
NETRADOCK_PROXY_HEADER=X-Forwarded-For
NETRADOCK_SECURE_COOKIE=true
```

---

## ⚙️ Configuration Reference

Netradock is configured via environment variables or a `.env` file:

| Variable | Type | Default | Required | Description |
| :--- | :--- | :--- | :--- | :--- |
| `NETRADOCK_USERNAME` | string | `admin` | **Yes** | Username required to log into the web console. |
| `NETRADOCK_PASSWORD` | string | `change-me` | Conditional | Plaintext password. Used only if `NETRADOCK_PASSWORD_HASH` is unset. |
| `NETRADOCK_PASSWORD_HASH` | string | _None_ | **Recommended** | Bcrypt hash of the login password (generate with `make hash`). Prevents plaintext password exposure in `docker inspect`. |
| `NETRADOCK_SECRET` | string | _Random on boot_ | **Recommended** | 32-byte secret (hex string) used to HMAC-sign session cookies. If unset, a random key is generated on boot and all sessions expire upon restart. |
| `NETRADOCK_SECURE_COOKIE` | boolean | `true` | No | Sets the `Secure` attribute on the session cookie. Set to `false` only for local unencrypted HTTP development. |
| `NETRADOCK_PROXY_HEADER` | string | `""` | **Recommended** | Header containing real client IP behind a reverse proxy (`X-Forwarded-For` or `X-Real-IP`). Required for accurate login rate limiting. |
| `NETRADOCK_PORT` | integer | `8080` | No | Port on which the HTTP server listens. |
| `NETRADOCK_SESSION_HOURS` | integer | `12` | No | Lifetime of a user session token in hours. |
| `NETRADOCK_LOG_LEVEL` | string | `info` | No | Logrus logging level: `debug`, `info`, `warn`, `error`. |

---

## 🔒 Security Architecture

### 1. Docker Socket Access
> [!CAUTION]
> Anyone authenticated into Netradock possesses root-equivalent privileges on the host system via the `/var/run/docker.sock` socket, including the capability to attach shells to any container or mount host paths. Protect your credentials accordingly.

### 2. Stateless HMAC Tokens
- Authentication produces a signed, stateless HMAC token held in an `HttpOnly`, `SameSite=Lax`, and `Secure` cookie.
- No server-side session database is maintained.
- If `NETRADOCK_USERNAME`, the password, or `NETRADOCK_SECRET` is modified, all existing session tokens are invalidated instantaneously.
- Active WebSocket streams (stats, logs, exec terminal) re-verify token validity every 30 seconds and disconnect immediately if verification fails.

### 3. Rate Limiting & CSRF Hardening
- **Login Rate Limiter:** Restricted to **5 attempts per minute per IP** and **30 attempts per minute globally** to prevent brute-force attacks.
- **CSRF Defense:** All mutation endpoints (`POST`, `DELETE`) and WebSocket upgrades enforce strict same-origin origin validation.
- **HTTP Security Headers:** Protected by Fiber Helmet with a strict `Content-Security-Policy`, `X-Frame-Options: DENY`, `no-referrer`, and `HSTS`.

---

## 🛠️ Development & Building

### Prerequisites
- **Go:** 1.26 or higher
- **Node.js:** 22 or higher (with npm)
- **Docker:** Running local Docker daemon with socket access

### Development Workflow

Start the backend API and frontend Vite dev server concurrently with hot-reloading:

```sh
make dev
```
- **Web UI:** [http://localhost:5178](http://localhost:5178)
- **API Server:** `http://localhost:18080`
- **Default Credentials:** `admin` / `admin`

### Useful Make Commands

| Command | Description |
| :--- | :--- |
| `make dev` | Starts Go backend on `:18080` & Vite frontend on `:5178` with hot reload. |
| `make build` | Builds the production Vue frontend and compiles the single `./netradock` binary. |
| `make test` | Runs Go automated unit tests (`go test ./...`). |
| `make hash` | Securely prompts for a password and outputs the bcrypt `NETRADOCK_PASSWORD_HASH`. |
| `make audit` | Runs `govulncheck` on Go code and `npm audit` on frontend dependencies. |
| `make install` | Downloads Go modules and installs frontend npm packages. |

---

## 🤝 Contributing

Contributions are warmly welcomed! Whether you are reporting a bug, proposing a feature, or submitting a pull request:

1. **Fork** the repository on GitHub.
2. **Create a branch** for your feature or bugfix (`git checkout -b feature/awesome-feature`).
3. **Write tests** and ensure existing tests pass (`make test`).
4. **Commit your changes** with clear, descriptive commit messages.
5. **Open a Pull Request** explaining your motivation and changes.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
