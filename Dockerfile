# One Dockerfile for all app images (D-008). Build context: the repo root.
#
#   docker build --target go-service --build-arg CMD=api .
#
# Slice 01-frontend-trim adds a frontend target.

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
