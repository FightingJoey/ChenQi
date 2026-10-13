# Multi-stage: build frontend + backend, serve both from Go
FROM node:20-alpine AS fe
WORKDIR /fe
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS be
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=be /out/server /app/server
COPY --from=fe /fe/dist /app/static
ENV ADDR=:8080
ENV DB_PATH=/data/savers.db
ENV STATIC_DIR=/app/static
ENV TZ=Asia/Shanghai
ENV JWT_SECRET=change-me-in-production
EXPOSE 8080
VOLUME ["/data"]
CMD ["/app/server"]
