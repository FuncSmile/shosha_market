# Shosha Mart POS (Offline-First)

Stack mengikuti techspec: Electron (main), Vue 3 + Vite (TypeScript) + Tailwind untuk renderer, serta sidecar backend Go (Gin + GORM + SQLite) yang siap sinkron ke PostgreSQL. Semua data ditulis ke SQLite lebih dulu; flag `synced` disiapkan untuk antrean upload saat koneksi tersedia.

## 📥 Download Aplikasi

**[⬇️ Download Latest Release](https://github.com/FuncSmile/shosha_market/releases/latest)**

Atau pilih versi spesifik:
- **Windows**: [ShoshaMart-POS-Setup.exe](https://github.com/FuncSmile/shosha_market/releases)
- **Linux**: [ShoshaMart-POS.AppImage](https://github.com/FuncSmile/shosha_market/releases)
- **macOS**: [ShoshaMart-POS.dmg](https://github.com/FuncSmile/shosha_market/releases)

## Struktur
- `renderer/` – UI desktop (Vue 3, TypeScript, Tailwind).
- `backend/` – Sidecar Go (Gin + GORM + SQLite + Excelize) di `127.0.0.1:8080`.
- `electron-main/` – Main process Electron yang men‐spawn backend dan memuat renderer.

## Menjalankan
1) **Backend (dev)**  
   ```bash
   cd backend
   go run main.go
   ```
   Endpoint contoh: `GET /api/health`, CRUD `/api/products`, `/api/branches`, `POST /api/sales`, `POST /api/stock-opname`, export `GET /api/reports/sales` dan `/api/stock-opname/:id/report`.

2) **Renderer (UI)**  
   ```bash
   cd renderer
   npm install
   npm run dev
   ```
   Buka `http://localhost:5173`. Build produksi: `npm run build`.

3) **Electron (dev, menjalankan backend otomatis)**  
   ```bash
   cd electron-main
   npm install
   npm start
   ```
   `main.js` akan mencoba menjalankan binary `backend/server` jika ada, fallback ke `go run main.go`. Jendela Electron akan memuat dev server Vite (`http://localhost:5173`).

## Catatan Tech
- Paket dipilih yang masih aktif/LTS: Electron `^30`, Vite `^6`, Vue `3.5`, TypeScript `~5.6`, Tailwind `3.4`, Gin `1.10`, GORM `1.25`, Excelize `2.8`.
- Backend mengikat ke `127.0.0.1:8080` untuk keamanan desktop dan mencatat model: Product, Branch, Sale (+ items), StockOpname (+ items), StoreProfile.
- Export laporan menggunakan Excelize dan dikirim sebagai lampiran (renderer mengubah ke unduhan blob).

## Sinkronisasi (garis besar)
- Setiap write memberi `synced=false` untuk antrean upload ke server pusat.
- SQLite menjadi sumber kebenaran saat offline; endpoint disiapkan agar dapat dipakai rutin dan diintegrasikan ke mekanisme sync (belum ada worker sync di branch ini).

---

## 📦 Build & Distribution

### Docker Deployment
Lihat dokumentasi lengkap di **[DOCKER.md](./DOCKER.md)**

```bash
# Production
docker-compose up -d

# Development (dengan hot-reload)
docker-compose -f docker-compose.dev.yml up
```

### Electron Desktop App

#### Option 1: Build Lokal (Linux)

```bash
# Build renderer
cd renderer
npm install
npm run build

# Build backend binary
cd ../backend
go build -o server main.go

# Copy ke electron-main
cp server ../electron-main/resources/backend/

# Build Electron app
cd ../electron-main
npm install
npm run dist
```

**Output**: `release/ShoshaMart-POS-0.1.0.AppImage` (~145 MB)

**Dokumentasi lengkap**: [BUILD_COMPLETE.md](./BUILD_COMPLETE.md)

#### Option 2: GitHub Actions (Recommended) ⭐

**Cara paling mudah untuk build multi-platform (Linux, Windows, macOS)**:

```bash
# 1. Push code ke GitHub
git add .
git commit -m "Ready for production"
git push origin fadli/dev

# 2. Buat release tag
git tag v0.1.0
git push origin v0.1.0

# 3. GitHub Actions otomatis build semua platform (~15 menit)
```

**Download installer**:
- Buka: `https://github.com/FuncSmile/shosha_market/releases/tag/v0.1.0`
- Download:
  - **Windows**: `ShoshaMart-POS-0.1.0-Setup.exe` (~100 MB)
  - **Linux**: `ShoshaMart-POS-0.1.0.AppImage` (~145 MB)
  - **macOS**: `ShoshaMart-POS-0.1.0.dmg` (~150 MB)

**Keuntungan GitHub Actions**:
- ✅ CGO otomatis enabled (SQLite support)
- ✅ Build native per platform (tidak perlu cross-compile)
- ✅ Multi-platform sekaligus
- ✅ Free untuk public repository
- ✅ Clean & reproducible builds

**Dokumentasi lengkap**: [GITHUB_ACTIONS_INSTALL.md](./GITHUB_ACTIONS_INSTALL.md)

### Instalasi untuk End User

#### Windows
1. Download `ShoshaMart-POS-0.1.0-Setup.exe`
2. Double-click installer
3. Next → Install → Done
4. App siap pakai (shortcut di Desktop)

#### Linux
1. Download `ShoshaMart-POS-0.1.0.AppImage`
2. Beri permission: `chmod +x ShoshaMart-POS-0.1.0.AppImage`
3. Jalankan: `./ShoshaMart-POS-0.1.0.AppImage`

#### macOS
1. Download `ShoshaMart-POS-0.1.0.dmg`
2. Open DMG
3. Drag app ke Applications folder

---

## 📚 Dokumentasi Lengkap

- **[PROJECT_ANALYSIS.md](./PROJECT_ANALYSIS.md)** - Analisis arsitektur project
- **[BUILD_COMPLETE.md](./BUILD_COMPLETE.md)** - Panduan build lokal lengkap
- **[BUILD_SOLUTIONS.md](./BUILD_SOLUTIONS.md)** - Troubleshooting build issues
- **[DOCKER.md](./DOCKER.md)** - Docker deployment guide
- **[GITHUB_ACTIONS_INSTALL.md](./GITHUB_ACTIONS_INSTALL.md)** - CI/CD & distribusi via GitHub
- **[FIX_PRODUCTION_PATH.md](./FIX_PRODUCTION_PATH.md)** - Fix production path loading
- **[FIX_WINDOWS_CGO.md](./FIX_WINDOWS_CGO.md)** - Fix Windows CGO compilation
