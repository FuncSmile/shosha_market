# Shosha Mart POS (Offline-First)

Stack mengikuti techspec: Electron (main), Vue 3 + Vite (TypeScript) + Tailwind untuk renderer, serta sidecar backend Go (Gin + GORM + SQLite) yang siap sinkron ke PostgreSQL. Semua data ditulis ke SQLite lebih dulu; flag `synced` disiapkan untuk antrean upload saat koneksi tersedia.

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
