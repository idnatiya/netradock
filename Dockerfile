FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/embed.go web/
COPY --from=web /src/web/dist web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /netradock ./cmd/web

# Runs as root because it talks to /var/run/docker.sock.
FROM gcr.io/distroless/static-debian12
COPY --from=server /netradock /netradock
EXPOSE 8080
ENTRYPOINT ["/netradock"]
