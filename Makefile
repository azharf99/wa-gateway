# Nama binary output
APP_NAME=wa-gateway
MAIN_FILE=cmd/api/main.go

.PHONY: all tidy run build build-linux clean reset-db

all: tidy build

# Merapikan dan mengunduh dependency (mirip npm install)
tidy:
	@echo "=> Merapikan go.mod..."
	go mod tidy

# Menjalankan aplikasi untuk testing di lokal
run:
	@echo "=> Menjalankan server lokal..."
	go run $(MAIN_FILE)

# Build binary untuk OS Windows lokal
build:
	@echo "=> Melakukan build untuk Windows..."
	go build -o $(APP_NAME).exe $(MAIN_FILE)

# Build khusus untuk di-deploy ke VPS Linux (Cross-Compilation)
# Gunakan target ini jika mengeksekusi lewat Git Bash / WSL / Terminal Linux
build-linux:
	@echo "=> Melakukan cross-compile untuk VPS Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux $(MAIN_FILE)

# Membersihkan file binary hasil build sebelumnya
clean:
	@echo "=> Membersihkan file build..."
	rm -f $(APP_NAME).exe $(APP_NAME)-linux

# Mereset database SQLite jika ingin mengulang proses History Sync WA dari awal
reset-db:
	@echo "=> Menghapus database sesi WhatsApp dan User..."
	rm -f sessions.db sessions.db-shm sessions.db-wal