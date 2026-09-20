# Values from .env (if present) override the dev defaults below. Keep .env values unquoted.
-include .env
export

NETRADOCK_USERNAME ?= admin
NETRADOCK_PASSWORD ?= admin
NETRADOCK_SECRET ?= dev-only-secret
NETRADOCK_PORT ?= 18080
WEB_PORT ?= 5178

.PHONY: dev dev-api dev-web install build test

## dev: run the Go API and the Vite dev server together (Ctrl+C stops both)
dev: web/node_modules web/dist/index.html
	@echo "UI:  http://localhost:$(WEB_PORT)  (login $(NETRADOCK_USERNAME))"
	@echo "API: http://localhost:$(NETRADOCK_PORT)"
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory dev-api & \
	$(MAKE) --no-print-directory dev-web & \
	wait

dev-api: web/dist/index.html
	go run ./cmd/web

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
