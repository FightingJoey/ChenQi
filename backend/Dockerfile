FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/server /app/server
ENV ADDR=:8080
ENV DB_PATH=/data/savers.db
ENV TZ=Asia/Shanghai
EXPOSE 8080
VOLUME ["/data"]
CMD ["/app/server"]
