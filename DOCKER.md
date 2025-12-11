# Docker Setup Guide - Shosha Mart POS

Project ini adalah aplikasi POS offline-first dengan arsitektur multi-service:
- **Backend**: Go (Gin) Sidecar Service pada port 8080
- **Frontend**: Vue 3 (Vite) pada port 3000
- **Database**: SQLite (lokal), optional PostgreSQL untuk production sync

## Quick Start

### 1. Production Build & Run

```bash
# Build dan jalankan dengan docker-compose
docker-compose up -d

# Backend akan accessible di: http://localhost:8080
# Frontend akan accessible di: http://localhost:3000

# Lihat logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Stop
docker-compose down
```

### 2. Development dengan Hot Reload

```bash
# Jalankan dengan docker-compose dev
docker-compose -f docker-compose.dev.yml up

# Backend (Go):
#   - Port: 8080
#   - Hot reload: enabled (air)
#   - Ubah file .go akan auto rebuild

# Frontend (Vue):
#   - Port: 5173
#   - Hot reload: enabled (Vite)
#   - Ubah file .vue/.ts akan auto reload di browser

# Akses: http://localhost:5173
```

### 3. Build Images Manually

```bash
# Build backend image
docker build -t shosha-mart-backend:latest -f Dockerfile.backend .

# Build frontend image
docker build -t shosha-mart-frontend:latest -f Dockerfile.frontend .

# Run backend
docker run -d \
  --name shosha-backend \
  -p 8080:8080 \
  -v shosha_data:/app/data \
  -e POS_BIND_ADDR=0.0.0.0:8080 \
  shosha-mart-backend:latest

# Run frontend
docker run -d \
  --name shosha-frontend \
  -p 3000:3000 \
  --link shosha-backend:backend \
  shosha-mart-frontend:latest
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│         Shosha Mart POS (Docker Compose)            │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌────────────────────┐   ┌──────────────────────┐ │
│  │   Frontend (Vue)   │   │  Backend (Go/Gin)    │ │
│  │   Port: 3000       │   │  Port: 8080          │ │
│  │                    │   │                      │ │
│  │  - Vite Dev Server │   │  - REST API Sidecar  │ │
│  │  - TailwindCSS     │   │  - SQLite Database   │ │
│  │  - Vue 3           │   │  - Sync Worker       │ │
│  └────────────────────┘   └──────────────────────┘ │
│          │                        │                 │
│          │────── HTTP/REST ───────│                 │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │   Persistent Volumes                         │  │
│  │   - backend_data (SQLite database)           │  │
│  │   - backend_exports (Report exports)         │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## Environment Variables

### Backend
- `POS_DB_PATH`: Path ke SQLite database (default: `/app/data/offline.db`)
- `POS_BIND_ADDR`: Address untuk bind server (default: `0.0.0.0:8080`)
- `POS_EXPORT_DIR`: Directory untuk export reports (default: `/app/exports`)
- `POS_BRANCH_ID`: Branch ID untuk sinkronisasi (default: `docker-local`)
- `POS_UPSTREAM_URL`: URL server pusat untuk sync (optional)

### Frontend
- `VITE_BACKEND_URL`: URL backend untuk API calls (default dalam compose: `http://backend:8080`)

## Debugging

### Lihat logs backend
```bash
docker-compose logs -f backend
```

### Lihat logs frontend
```bash
docker-compose logs -f frontend
```

### Enter container
```bash
# Backend shell
docker-compose exec backend sh

# Frontend shell
docker-compose exec frontend sh
```

### Test backend health
```bash
curl http://localhost:8080/api/health
```

### Test frontend
```bash
curl http://localhost:3000
```

## Troubleshooting

### Port sudah digunakan
```bash
# Ganti port di docker-compose.yml atau .env
# Contoh:
docker-compose -f docker-compose.yml up \
  --set BACKEND_PORT=8081 \
  --set FRONTEND_PORT=3001
```

### Database permission error
```bash
# Pastikan volume memiliki permission yang tepat
docker volume ls
docker volume inspect shosha_mart_backend_data
```

### Frontend tidak bisa connect ke backend
- Pastikan service `backend` healthy: `docker-compose ps`
- Check CORS settings di `backend/routes/routes.go`
- Frontend dalam container gunakan: `http://backend:8080`

### Go module issues
```bash
# Clear Go cache
docker-compose exec backend go clean -modcache

# Rebuild dependencies
docker-compose exec backend go mod download
docker-compose rebuild backend
```

## Production Deployment

Untuk production dengan PostgreSQL:

1. Uncomment PostgreSQL service di `docker-compose.yml`
2. Set environment variables:
   ```bash
   UPSTREAM_URL=https://your-server.com
   DB_USER=shosha
   DB_PASSWORD=secure-password
   ```
3. Deploy dengan:
   ```bash
   docker-compose -f docker-compose.yml up -d
   ```

## Notes

- Database SQLite disimpan di volume `backend_data` - persistent across container restarts
- Frontend dioptimasi dengan multi-stage build untuk ukuran image yang lebih kecil
- Health checks diatur untuk automatic recovery
- CORS dikonfigurasi untuk komunikasi frontend-backend lokal
- Environment variables dapat override via `.env` file atau command line
