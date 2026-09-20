# Values from .env (if present) override the dev defaults below. Keep .env values unquoted.
# Exception: make eats the `$` in a bcrypt hash, so NETRADOCK_PASSWORD_HASH is read raw in dev-api.
-include .env
export

NETRADOCK_USERNAME ?= admin
NETRADOCK_PASSWORD ?= admin
# A fresh key per `make dev`: a committed secret lets anyone forge a session token.
ifeq ($(origin NETRADOCK_SECRET), undefined)
NETRADOCK_SECRET := $(shell openssl rand -hex 32)
endif
# The dev server speaks plain HTTP, so the browser would drop a Secure cookie.
NETRADOCK_SECURE_COOKIE ?= false
NETRADOCK_PORT ?= 18080
WEB_PORT ?= 5178
DOCS_PORT ?= 8000

.PHONY: dev dev-api dev-web install build test audit hash docs view-docs

## dev: run the Go API and the Vite dev server together (Ctrl+C stops both)
dev: web/node_modules web/dist/index.html
	@echo "UI:  http://localhost:$(WEB_PORT)  (login $(NETRADOCK_USERNAME))"
	@echo "API: http://localhost:$(NETRADOCK_PORT)"
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory dev-api & \
	$(MAKE) --no-print-directory dev-web & \
	wait

dev-api: web/dist/index.html
	@NETRADOCK_PASSWORD_HASH="$$(sed -n 's/^NETRADOCK_PASSWORD_HASH=//p' .env 2>/dev/null)" go run ./cmd/web

dev-web: web/node_modules
	cd web && NETRADOCK_BACKEND=http://localhost:$(NETRADOCK_PORT) npx vite --port $(WEB_PORT) --strictPort

install: web/node_modules
	go mod download

web/node_modules: web/package.json web/package-lock.json
	cd web && npm install
	@touch $@

# go:embed needs at least one file in web/dist; in dev the UI is served by Vite instead.
web/dist/index.html:
	mkdir -p web/dist
	echo '<!doctype html><title>Netradock</title><p>Dev mode: open the Vite URL, or run make build.' > $@

build: web/node_modules
	cd web && npm run build
	go build -o netradock ./cmd/web

test:
	go test ./...

## hash: prompt for a password and print the NETRADOCK_PASSWORD_HASH line for .env
hash:
	@printf 'Password: ' >&2; \
	stty -echo 2>/dev/null; read -r pass; stty echo 2>/dev/null; printf '\n' >&2; \
	printf '%s' "$$pass" | go run ./cmd/hash | sed 's/^/NETRADOCK_PASSWORD_HASH=/'

## audit: check Go and npm dependencies for known vulnerabilities (needs network)
audit: web/node_modules
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	cd web && npm audit

## docs: preview landing page locally at http://localhost:$(DOCS_PORT)
docs:
	@echo "Docs: http://localhost:$(DOCS_PORT)"
	@(sleep 1 && (open "http://localhost:$(DOCS_PORT)" 2>/dev/null || xdg-open "http://localhost:$(DOCS_PORT)" 2>/dev/null || true)) & \
	if command -v python3 >/dev/null 2>&1; then \
		python3 -m http.server $(DOCS_PORT) -d docs; \
	elif command -v npx >/dev/null 2>&1; then \
		npx --yes serve -p $(DOCS_PORT) docs; \
	else \
		open docs/index.html 2>/dev/null || xdg-open docs/index.html 2>/dev/null; \
	fi

view-docs: docs
