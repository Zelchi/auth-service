ARG NODE_IMAGE=node:24-alpine
ARG GO_IMAGE=golang:1.26-alpine
ARG RUNTIME_IMAGE=alpine:3.21

FROM ${NODE_IMAGE} AS frontend

WORKDIR /app/frontend

COPY frontend/package.json frontend/yarn.lock ./

RUN yarn install --frozen-lockfile

COPY frontend/ ./

ARG VITE_AUTH_BRIDGE_ORIGINS
ENV VITE_AUTH_BRIDGE_ORIGINS=${VITE_AUTH_BRIDGE_ORIGINS}

RUN yarn build

FROM ${GO_IMAGE} AS backend

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./

RUN go mod download

COPY backend/ ./

COPY --from=frontend /app/frontend/dist ./cmd/dist

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/main.go

FROM ${RUNTIME_IMAGE}

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app && mkdir -p /data && chown -R app:app /data

COPY --from=backend /app/server ./server

USER app

EXPOSE 8888

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8888/healthz || exit 1

CMD ["./server"]
