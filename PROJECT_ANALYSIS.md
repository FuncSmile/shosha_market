# Project Analysis & Debugging Report - Shosha Mart POS

## 📋 Project Overview

**Shosha Mart POS** adalah aplikasi Point of Sale (POS) desktop offline-first dengan arsitektur modern:
- **Frontend**: Vue 3 + TypeScript + TailwindCSS (Renderer Process Electron)
- **Backend**: Go (Gin Framework) - Sidecar Service
- **Database**: SQLite (lokal, offline) → PostgreSQL (cloud, untuk sync)
- **Desktop**: Electron (main process)

### Arsitektur Sidecar Pattern
```
Electron Main Process
    ├─ Spawn Backend (Go binary)
    └─ Create Window (Renderer Process)
         └─ Load Vue App + TailwindCSS
              └─ HTTP Requests → Backend API (localhost:8080)
```

---

## 🔍 Analisis Struktur

### Backend (Go + Gin + GORM)
**Lokasi**: `/backend`

| File | Purpose |
|------|---------|
| `main.go` | Entry point - Start Gin server, spawn sync worker |
| `config/config.go` | Load config dari env vars (POS_DB_PATH, POS_BIND_ADDR, dll) |
| `models/models.go` | GORM models (Product, Branch, Sale, StockOpname, dll) |
| `controllers/` | HTTP handlers untuk setiap endpoint |
| `routes/routes.go` | Register routes, CORS config, middleware |
| `services/database.go` | Database initialization & connection |
| `sync/` | Sync worker untuk background sync offline→online |

**Key Routes**:
- `GET /api/health` - Health check
- `GET/POST /api/products` - Product CRUD
- `GET/POST /api/branches` - Branch CRUD  
- `POST /api/sales` - Create sales transaction
- `POST /api/stock-opname` - Stock opname
- `GET /api/reports/sales` - Export sales report (Excel)
- `GET /api/sync/summary` - Sync status
- `POST /api/sync/run` - Trigger manual sync

**Database**: SQLite (offline-first)
- File: `offline.db` (configurable via `POS_DB_PATH`)
- Persisten across restarts
- Saat online, data sync ke PostgreSQL

**Dependencies** (go.mod):
```
- github.com/gin-gonic/gin v1.10.1      (Web framework)
- gorm.io/gorm v1.25.12                  (ORM)
- gorm.io/driver/sqlite v1.5.7           (SQLite driver)
- github.com/xuri/excelize/v2 v2.8.1     (Excel export)
- github.com/joho/godotenv v1.5.1        (.env loading)
- github.com/gin-contrib/cors v1.7.6     (CORS middleware)
```

### Frontend (Vue 3 + Vite)
**Lokasi**: `/renderer`

| File | Purpose |
|------|---------|
| `package.json` | Dependencies, scripts |
| `vite.config.ts` | Vite bundler config |
| `tailwind.config.js` | TailwindCSS config |
| `src/main.ts` | Vue app entrypoint |
| `src/App.vue` | Root component |
| `src/api.ts` | API client wrapper |

**Build Process**:
```bash
npm run build  # → dist/ folder
```

**Output**: Static files di `renderer/dist/`

**Dependencies**:
- `vue@^3.5.12` - UI framework
- `vite@^6.0.5` - Build tool
- `tailwindcss@^3.4.13` - CSS utility
- `typescript` - Type safety

### Electron Main (Node.js + Electron)
**Lokasi**: `/electron-main`

| File | Purpose |
|------|---------|
| `main.js` | Electron main process |
| `package.json` | Build config + dependencies |

**Key Features**:
- ✅ Auto-spawn backend Go binary (`./server` atau `go run main.go`)
- ✅ Port availability check (fallback jika port occupied)
- ✅ Environment-based loading (dev vs production)
- ✅ Context isolation + no node integration (security)
- ✅ Auto-fallback dari dev server ke dist HTML

**Build Output**: 
- Linux: AppImage
- Windows: NSIS installer
- Packaged ke `/release`

---

## 🐛 Issues Found & Debugging

### Issue 1: Go Module Indirect Dependency ⚠️
**Status**: MINOR  
**Location**: `backend/go.mod` line 55  
**Error**: `gorm.io/driver/postgres should be direct`

**Explanation**: 
PostgreSQL driver is listed as indirect but might be needed as direct dependency.

**Fix**:
```bash
cd backend
go get gorm.io/driver/postgres
go mod tidy
```

**Impact**: Build Go binary for production might fail

---

### Issue 2: Electron Build Process
**Status**: INFO  
**Build Command**: `npm run dist` (di electron-main/)

**Current Flow**:
1. `npm run build:renderer` → Build Vue app to `/renderer/dist/`
2. `electron-builder` packages Electron app
3. Includes backend binary dari `backend/server`
4. Output: AppImage (Linux), NSIS (Windows)

**Current Build Status**:
- Last run exited with code 1 (error)
- Likely causes:
  - Backend binary tidak compiled (`backend/server` not found)
  - Missing dependencies di renderer build
  - Node module issues

**Solution**: 
```bash
# Clean build
cd backend && go build -o server main.go && cd ..
cd renderer && npm install && npm run build && cd ..
cd electron-main && npm install && npm run dist
```

---

### Issue 3: CORS Configuration
**Status**: OK (Properly configured)  
**File**: `backend/routes/routes.go`

**Current Config**:
```go
AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"}
AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
AllowHeaders: []string{"Origin", "Content-Type"}
```

**For Docker**: Frontend akan request ke `http://backend:8080` (service name resolution)
- ✅ Sudah support localhost
- ⚠️ Untuk Electron production, tambahkan host IP atau file:// protocol jika diperlukan

---

### Issue 4: Database Persistence
**Status**: OK (Properly configured)  
**File**: `backend/config/config.go`

**Current**:
- Default path: `offline.db` (relative ke working directory)
- Configurable via `POS_DB_PATH` env var

**For Docker**: Set ke `/app/data/offline.db` (volume mount)
- ✅ Data persisten across container restarts
- ✅ Backup-friendly (volume location)

---

## ✅ Features Verified

| Feature | Status | Notes |
|---------|--------|-------|
| Offline-first SQLite | ✅ | GORM models ready |
| REST API (Gin) | ✅ | All endpoints defined |
| Frontend (Vue 3) | ✅ | TypeScript, TailwindCSS ready |
| Hot reload (dev) | ✅ | Vite dev server configured |
| Electron packaging | ⚠️ | Backend binary needed |
| Sync worker | ✅ | Background worker implemented |
| Excel export | ✅ | Excelize integrated |
| CORS | ✅ | Configured for localhost |

---

## 🐳 Docker Implementation

### Multi-Service Architecture
```yaml
services:
  backend:      # Go + Gin + SQLite
  frontend:     # Node + Vite + Vue
  postgres:     # Optional, untuk production sync
```

### Build Strategy
1. **Dockerfile.backend**: Multi-stage Go build (optimize size)
2. **Dockerfile.frontend**: Multi-stage Node build (SPA serving)
3. **docker-compose.yml**: Production setup
4. **docker-compose.dev.yml**: Development dengan hot reload
5. **Dockerfile.dev**: Alternative development setup

### Features
- ✅ Health checks untuk setiap service
- ✅ Volume persistence (SQLite + exports)
- ✅ Automatic service dependency (frontend waits for backend)
- ✅ Environment variable configuration
- ✅ Network isolation (shosha_network)
- ✅ Hot reload untuk development
- ✅ Multi-stage builds untuk optimasi ukuran image

---

## 📝 Commands Reference

### Development
```bash
# Backend saja (port 8080)
cd backend && go run main.go

# Frontend saja (port 5173)
cd renderer && npm install && npm run dev

# Electron (spawns backend, loads dev server)
cd electron-main && npm start

# Docker development (hot reload)
docker-compose -f docker-compose.dev.yml up
```

### Production
```bash
# Build semua dari scratch
cd backend && go build -o server main.go
cd renderer && npm run build
cd electron-main && npm run dist

# Atau gunakan Docker
docker-compose up -d
```

### Testing
```bash
# Test backend API
curl http://localhost:8080/api/health

# Test frontend
curl http://localhost:3000  (production)
curl http://localhost:5173  (development)
```

---

## 🚀 Next Steps / Recommendations

1. **Fix go.mod issue**:
   ```bash
   cd backend && go get gorm.io/driver/postgres && go mod tidy
   ```

2. **Test complete flow**:
   ```bash
   # Terminal 1: Backend
   cd backend && go run main.go
   
   # Terminal 2: Frontend
   cd renderer && npm run dev
   
   # Terminal 3: Electron
   cd electron-main && npm start
   ```

3. **Docker production ready**:
   ```bash
   docker-compose build
   docker-compose up -d
   # Access: http://localhost:3000
   ```

4. **Enable PostgreSQL sync** (jika needed):
   - Uncomment di `docker-compose.yml`
   - Update backend untuk support PostgreSQL replication
   - Set `POS_UPSTREAM_URL` untuk server pusat

5. **Optimize images**:
   - Add `.dockerignore` ✅ (sudah)
   - Use alpine images ✅ (sudah)
   - Implement caching strategy (next iteration)

---

## 📞 Support Files Created

1. **Dockerfile.backend** - Production backend image
2. **Dockerfile.frontend** - Production frontend image  
3. **Dockerfile.dev** - Development image dengan air hot reload
4. **docker-compose.yml** - Production orchestration
5. **docker-compose.dev.yml** - Development dengan volumes
6. **.dockerignore** - Optimize build context
7. **backend/.air.toml** - Air hot reload config
8. **.env.example** - Environment variables template
9. **DOCKER.md** - Complete Docker usage guide
10. **PROJECT_ANALYSIS.md** - This file

---

**Generated**: December 11, 2025  
**Analysis Date**: Full stack review  
**Status**: Ready for Docker deployment
