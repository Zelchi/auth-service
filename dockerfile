FROM node:24-alpine AS frontend

WORKDIR /app/frontend

COPY frontend/package.json ./

RUN yarn install 

COPY frontend/ ./

RUN yarn build

FROM golang:1.26-alpine AS backend

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./

RUN go mod download

COPY backend/ ./

COPY --from=frontend /app/frontend/dist ./cmd/dist

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=backend /app/server ./server

EXPOSE 8888

CMD ["./server"]