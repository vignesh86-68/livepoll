# ── Stage 1: Build React frontend ────────────────────────────────────────
FROM node:20-alpine AS frontend

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --production=false
COPY frontend/ ./
RUN npm run build

# ── Stage 2: Build Go backend ───────────────────────────────────────────
FROM golang:1.22-alpine AS backend

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum* ./
RUN go mod download
COPY backend/ ./
# CGO disabled for a static binary that runs on distroless/scratch.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# ── Stage 3: Final image ────────────────────────────────────────────────
# Distroless is Google's minimal container image: no shell, no package
# manager, nothing except the binary. Smaller attack surface and smaller
# image size (~20MB vs ~800MB for the Go build stage).
FROM gcr.io/distroless/static-debian12

COPY --from=backend /server /server
COPY --from=frontend /app/frontend/dist /web

# The Go server reads STATIC_DIR to find the built React app.
ENV STATIC_DIR=/web
ENV APP_ENV=production

EXPOSE 8080

ENTRYPOINT ["/server"]
