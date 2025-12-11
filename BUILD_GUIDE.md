# Build Guide - Shosha Mart Electron App

Panduan lengkap untuk build aplikasi desktop untuk berbagai platform.

## 📋 Persyaratan

### Semua Platform
- **Node.js**: v20+
- **npm**: v10+
- **Go**: v1.23+

### Platform Spesifik
| OS | Requirements |
|---|---|
| **Linux** | gcc, sqlite-dev, (optional) wine untuk Windows cross-compile |
| **Windows** | Visual Studio Build Tools, atau MinGW |
| **macOS** | Xcode Command Line Tools |

## 🚀 Quick Build

### Satu Command Build (Recommended)
```bash
bash build.sh
```

Akan automatically:
1. Build backend Go binary
2. Build renderer Vue
3. Build Electron app untuk OS saat ini
4. Output ke `release/` folder

### Build per Platform

#### Linux (AppImage)
```bash
cd electron-main
npm run dist:linux
# Output: release/ShoshaMart\ POS-0.1.0.AppImage
```

#### Windows (NSIS Installer)
```bash
cd electron-main
npm run dist:win
# Output: release/ShoshaMart\ POS-0.1.0\ Setup.exe
```

#### macOS (DMG)
```bash
cd electron-main
npm run build:renderer && npx electron-builder --mac --publish never
# Output: release/ShoshaMart\ POS-0.1.0.dmg
```

---

## 🔧 Detailed Build Steps

Jika ingin build manual dengan kontrol penuh:

### 1. Build Backend Binary
```bash
cd backend

# Linux/macOS
go mod download
CGO_ENABLED=1 go build -o server main.go

# Windows
go mod download
go build -o server.exe main.go
```

### 2. Build Frontend (Renderer)
```bash
cd renderer
npm ci                    # Fresh install
npm run build             # Build untuk production
```

Output: `renderer/dist/` folder

### 3. Copy Backend ke Electron Resources
```bash
mkdir -p electron-main/resources/backend
cp backend/server electron-main/resources/backend/    # Linux/macOS
cp backend/server.exe electron-main/resources/backend/ # Windows
```

### 4. Build Electron App
```bash
cd electron-main
npm ci

# Linux
npm run dist:linux

# Windows
npm run dist:win

# macOS
npm run build:renderer && npx electron-builder --mac --publish never
```

---

## 🔀 Cross-Platform Build

### Build Windows dari Linux

**Option 1: Menggunakan Wine (Recommended untuk Linux)**
```bash
# Install Wine
sudo pacman -S wine wine-mono wine-gecko  # Manjaro
sudo apt install wine wine-mono wine-gecko # Ubuntu/Debian

# Build Windows binary
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o server.exe backend/main.go

# Build Electron Windows
cd electron-main
npm run dist:win
```

**Option 2: Menggunakan GitHub Actions**
Lihat `.github/workflows/build-electron.yml` - Build otomatis di cloud untuk semua platform.

### Build macOS dari Linux
Tidak support secara native. Gunakan GitHub Actions atau Mac machine.

---

## 🐳 Build Menggunakan Docker

### Build Linux AppImage dalam Docker
```bash
docker build -t shosha-builder -f Dockerfile.electron .
docker run -v $(pwd)/release:/app/release shosha-builder sh -c "cd electron-main && npm run dist:linux"
```

### Development dalam Docker
```bash
docker-compose -f docker-compose.dev.yml up electron-dev
# Hot reload untuk backend & frontend
```

---

## ⚙️ Environment Variables

### Backend Build
```bash
# Configurable path untuk SQLite database
export POS_DB_PATH="offline.db"

# Backend bind address
export POS_BIND_ADDR="127.0.0.1:8080"

# Export directory untuk reports
export POS_EXPORT_DIR="exports"
```

### Electron Build
```bash
# Skip Windows build jika di Linux
export SKIP_WIN=1

# Skip macOS build jika di Linux
export SKIP_MAC=1

# Custom output directory
export ELECTRON_OUT_DIR="./build"
```

---

## 📦 Output Artifacts

### Linux
```
release/
├── ShoshaMart POS-0.1.0.AppImage      # Executable AppImage
├── ShoshaMart POS-0.1.0.AppImage.blockmap  # Signature
└── linux-unpacked/                    # Unpacked files
    └── resources/
        ├── app.asar                   # Electron app bundle
        └── backend/
            └── server                 # Go backend binary
```

### Windows
```
release/
├── ShoshaMart POS-0.1.0 Setup.exe     # NSIS Installer
├── ShoshaMart POS-0.1.0.exe.blockmap  # Signature
└── win-unpacked/                      # Unpacked files
    └── resources/
        ├── app.asar                   # Electron app bundle
        └── backend/
            └── server.exe             # Go backend binary
```

### macOS
```
release/
├── ShoshaMart POS-0.1.0.dmg           # DMG installer
├── ShoshaMart POS-0.1.0.dmg.blockmap  # Signature
└── mac-unpacked/                      # Unpacked files
    └── ShoshaMart POS.app/
        └── Contents/
            └── Resources/
                ├── app.asar           # Electron app bundle
                └── backend/
                    └── server         # Go backend binary
```

---

## 🐛 Troubleshooting

### "wine is required" Error (Linux)
```bash
# Install Wine
sudo pacman -S wine wine-mono wine-gecko

# Or skip Windows build
npm run dist:linux  # Hanya Linux
```

### "Build target is not set" Error
```bash
# Pastikan electron-main/resources/backend/server(exe) exists
ls -la electron-main/resources/backend/server*

# If not, rebuild backend:
cd backend && go build -o server main.go
```

### "Module not found" Error
```bash
# Clean install dependencies
cd renderer && rm -rf node_modules package-lock.json && npm install
cd ../electron-main && rm -rf node_modules package-lock.json && npm install
```

### macOS Code Signing Error
```bash
# Build tanpa signing (development)
npx electron-builder --mac --publish never --no-sign
```

### Windows Icon Not Found
Tambahkan icon ke `electron-main/package.json`:
```json
"build": {
  "icon": "./icon.png",  // 512x512
  "win": {
    "icon": "./icon.ico"  // Buat dari PNG dengan imagemagick
  }
}
```

---

## 🔄 CI/CD Pipeline (GitHub Actions)

Workflow otomatis akan trigger saat:
- Push ke `main`, `develop`, `fadli/dev`
- Buat tag (release)

**Artifacts tersedia di**:
- GitHub Actions > Artifacts (setelah build)
- GitHub Releases (saat ada tag)

**Setup**:
1. Push code ke GitHub
2. Tag untuk release: `git tag v0.1.0 && git push origin v0.1.0`
3. Workflow auto-build semua platform
4. Download dari GitHub Releases

---

## 📝 Package.json Scripts

```json
{
  "scripts": {
    "dev": "ELECTRON_DEV=1 electron .",           // Run dev
    "start": "electron .",                        // Run production build
    "build:renderer": "npm run build --prefix ../renderer",  // Build Vue
    "dist": "npm run build:renderer && electron-builder --linux AppImage --win nsis",  // Build semua
    "dist:linux": "npm run build:renderer && electron-builder --linux AppImage",       // Hanya Linux
    "dist:win": "npm run build:renderer && electron-builder --win nsis"                // Hanya Windows
  }
}
```

---

## 🚀 Release Checklist

Sebelum release ke production:

- [ ] Test di semua platform
- [ ] Update version di `package.json`
- [ ] Build dengan `bash build.sh`
- [ ] Test executable dari `release/`
- [ ] Verify no console errors
- [ ] Create git tag: `git tag v0.2.0`
- [ ] Push tag: `git push origin v0.2.0`
- [ ] GitHub Actions akan auto-build dan release

---

**Last Updated**: December 11, 2025  
**Status**: Multi-platform build ready
