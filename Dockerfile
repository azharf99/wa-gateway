# ==========================================
# STAGE 1: Builder
# ==========================================
FROM golang:1.26-alpine AS builder

# Set working directory di dalam container
WORKDIR /app

# Install tzdata (SANGAT PENTING untuk Gocron agar zona waktu akurat)
RUN apk add --no-cache tzdata

# Copy file dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary aplikasi
# CGO_ENABLED=0 aman karena kita menggunakan modernc.org/sqlite
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wa-gateway cmd/api/main.go

# ==========================================
# STAGE 2: Runner (Image final yang sangat kecil)
# ==========================================
FROM alpine:latest

WORKDIR /app

# Copy konfigurasi zona waktu dari builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Jakarta

# Copy file binary dari stage builder
COPY --from=builder /app/wa-gateway .

# Buat direktori khusus untuk menyimpan database SQLite agar persisten
RUN mkdir -p /app/data

# Expose port internal aplikasi (Gin default jalan di 8080)
EXPOSE 8080

# Jalankan aplikasi
CMD ["./wa-gateway"]