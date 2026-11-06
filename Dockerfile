# One Dockerfile for all app images (D-008). Build context: the repo root.
#
#   docker build --target go-service --build-arg CMD=api .
#
#   docker build --target frontend .

# ---- Go binaries (cmd/api, cmd/mockprovider, later cmd/worker, cmd/seed) ----
FROM golang:1.27-alpine AS go-build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
# ARG CMD comes after the download, so all binaries share this cache layer.
RUN go mod download
COPY backend/ ./
ARG CMD
RUN test -n "$CMD" || (echo "build arg CMD is required, for example --build-arg CMD=api" >&2 && exit 1)
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${CMD}

FROM gcr.io/distroless/static-debian12:nonroot AS go-service
COPY --from=go-build /out/app /app
ENTRYPOINT ["/app"]

# ---- Frontend (frontend/, served by nginx) ----
FROM node:24-alpine AS frontend-build
WORKDIR /app
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
# The build context has no .env (see .dockerignore). Vite reads these instead.
ARG VITE_SERVER_API_URL=http://localhost:8080
ARG VITE_SERVER_API_PREFIX=/api
ARG VITE_SERVER_API_TIMEOUT=10000
ENV VITE_SERVER_API_URL=${VITE_SERVER_API_URL} \
    VITE_SERVER_API_PREFIX=${VITE_SERVER_API_PREFIX} \
    VITE_SERVER_API_TIMEOUT=${VITE_SERVER_API_TIMEOUT}
RUN pnpm build

FROM nginxinc/nginx-unprivileged:1.29-alpine AS frontend
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-build /app/dist /usr/share/nginx/html
