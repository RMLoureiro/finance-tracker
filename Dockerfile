# --- frontend: export estático do Next ---
FROM node:22-alpine AS webbuild
WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
ENV BUILD_EXPORT=1
RUN npm run build

# --- backend: binário Go ---
FROM golang:1.25-alpine AS gobuild
WORKDIR /build/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

# --- imagem final: um binário + ficheiros estáticos ---
FROM alpine:3.21
RUN adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=gobuild /out/server ./server
COPY --from=webbuild /build/web/out ./web-dist
ENV PORT=8080 WEB_DIR=/app/web-dist
EXPOSE 8080
ENTRYPOINT ["/app/server"]
