# Panduan Deployment ke Dockploy

## ✅ Status Kesiapan

Docker-compose Anda sudah siap untuk Dockploy dengan beberapa penyesuaian.

## 📋 Checklist Deployment

### 1. **File yang diperlukan:**
- ✅ `Dockerfile.backend` - Multi-stage Go build
- ✅ `Dockerfile.frontend` - Multi-stage Node.js build
- ✅ `docker-compose.dockploy.yml` - Optimized untuk Dockploy
- ✅ `.env.dockploy` - Environment variables template

### 2. **Langkah Deployment ke Dockploy:**

#### Step 1: Setup Repository
```bash
# Pastikan semua sudah di-commit
git add -A
git commit -m "chore: prepare for Dockploy deployment"
git push origin fadli/dev
```

#### Step 2: Di Dockploy Dashboard
1. Buka Dockploy dashboard
2. **Create New Project** atau gunakan existing
3. **Connect Repository** → pilih `shosha_market` dari GitHub
4. **Configure**:
   - Branch: `fadli/dev`
   - Build Type: Docker Compose
   - Docker Compose File: `docker-compose.dockploy.yml`

#### Step 3: Set Environment Variables
Di Dockploy dashboard, tambahkan env vars dari `.env.dockploy`:
```
POS_BRANCH_ID=docker-local (atau custom)
POS_UPSTREAM_URL=(optional, untuk sync ke PostgreSQL)
VITE_API_BASE=http://backend:8080/api (atau sesuaikan URL domain)
```

#### Step 4: Configure Domain (Optional)
- Jika ingin akses public:
  1. Ke **Settings** → **Domains**
  2. Tambah domain → pointing ke frontend container
  3. Dockploy akan handle SSL auto dengan Let's Encrypt

#### Step 5: Deploy
- Click **Deploy** atau setup auto-deployment on push
- Monitor logs di dashboard
- Tunggu health checks pass (backend & frontend)

## 🔧 Penyesuaian yang Dibuat

### docker-compose.dockploy.yml
```yaml
# Change 1: Backend ports → expose
# Dari: ports: ["8080:8080"]
# Ke:   expose: ["8080"]
# Alasan: Dockploy handle reverse proxy, tidak perlu direct port exposure

# Change 2: Frontend ports → expose
# Dari: ports: ["3000:3000"]
# Ke:   expose: ["3000"]
# Alasan: Same reason

# Change 3: Environment variables → dynamic
# Dari: POS_BRANCH_ID=docker-local (hardcoded)
# Ke:   POS_BRANCH_ID=${POS_BRANCH_ID:-docker-local} (overridable)

# Change 4: Frontend API URL → configurable
# Dari: VITE_BACKEND_URL=http://backend:8080
# Ke:   VITE_API_BASE=${VITE_API_BASE:-http://backend:8080/api}
# Alasan: Agar bisa set custom domain atau URL di Dockploy
```

## 📊 Akses Setelah Deploy

- **Frontend**: https://your-domain.com
- **Backend API**: http://backend:8080 (internal network)
- **Health Check**: https://your-domain.com/api/health

## 🆘 Troubleshooting

### Problem: Frontend error "API not found"
**Solusi**: Set `VITE_API_BASE` di Dockploy dengan URL backend yang benar
- Jika backend exposed: `https://api.your-domain.com/api`
- Jika internal: `http://backend:8080/api`

### Problem: Port 8080 already in use
**Solusi**: Sudah fixed! Gunakan `expose` instead of `ports`

### Problem: Database lost after redeploy
**Solusi**: Volume `backend_data` sudah persistent di Dockploy, data aman

## 📈 Scaling (Opsional)

Jika perlu production-grade:
1. Uncomment PostgreSQL di `docker-compose.dockploy.yml`
2. Set `POS_UPSTREAM_URL` ke PostgreSQL connection string
3. Sync data offline ke cloud

## 🔒 Security Tips

1. **Change default values**:
   - Set unique `POS_BRANCH_ID`
   - Add authentication jika diperlukan

2. **Enable HTTPS**:
   - Dockploy auto-handle dengan Let's Encrypt
   - Enable di domain settings

3. **Backup Database**:
   - Setup automated volume backup di Dockploy
   - Atau export regular dari `/app/exports`

## ✨ Summary

Docker-compose Anda **siap deploy ke Dockploy**!

**Yang perlu diingat:**
- ✅ Use `docker-compose.dockploy.yml` (bukan original)
- ✅ Set env vars di Dockploy dashboard
- ✅ Frontend port 3000, Backend port 8080 (internal)
- ✅ Domain & SSL auto-configured

Good luck! 🚀
