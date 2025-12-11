# Build Strategy & Solutions

## Status Build Saat Ini

✅ **Linux AppImage**: Berhasil  
❌ **Windows NSIS**: Gagal (wine required)

---

## 3 Solusi untuk Multi-Platform Build

### Solusi 1: Build Per-Platform (Easiest)
Build di OS masing-masing:

```bash
# Di Linux
npm run dist:linux
# Output: release/ShoshaMart\ POS-0.1.0.AppImage

# Di Windows
npm run dist:win
# Output: release/ShoshaMart\ POS-0.1.0\ Setup.exe

# Di macOS
npm run build:renderer && npx electron-builder --mac
# Output: release/ShoshaMart\ POS-0.1.0.dmg
```

**Pros**: Sederhana, native build  
**Cons**: Butuh multiple machines

---

### Solusi 2: GitHub Actions (Recommended)
Automated build di cloud untuk semua platform

**Sudah setup** di `.github/workflows/build-electron.yml`

**How to use**:
1. Push code ke GitHub
2. Tag untuk release: 
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. GitHub Actions akan otomatis build untuk Linux + Windows + macOS
4. Download dari GitHub Releases

**Pros**: Semua platform, fully automated, no local setup  
**Cons**: Butuh GitHub repo

---

### Solusi 3: Local Cross-Compile + Wine
Build Windows dari Linux menggunakan Wine

```bash
# Install Wine (Manjaro)
sudo pacman -S wine wine-mono wine-gecko

# Atau Ubuntu
sudo apt install wine wine-mono wine-gecko

# Build semua
npm run dist
# Akan build Linux AppImage + Windows NSIS
```

**Pros**: All-in-one command, local build  
**Cons**: Wine setup kompleks, berat

---

## Recommended Workflow

### Development
```bash
# Quick test
cd electron-main
npm start  # Run dengan dev server

# Or use Docker
docker-compose -f docker-compose.dev.yml up
```

### Pre-Release Testing
```bash
# Build untuk platform saat ini
bash build.sh

# Test executable
./release/ShoshaMart\ POS-0.1.0.AppImage  # Linux
release\ShoshaMart\ POS-0.1.0\ Setup.exe  # Windows (from Windows)
```

### Release ke Production
```bash
# Update version
# Edit electron-main/package.json -> version

# Commit & tag
git add .
git commit -m "Release v0.2.0"
git tag v0.2.0
git push origin main
git push origin v0.2.0

# GitHub Actions akan auto-build
# Download dari: https://github.com/FuncSmile/shosha_market/releases
```

---

## Scripts & Tools Tersedia

| File | Purpose |
|------|---------|
| `build.sh` | Build untuk platform saat ini |
| `build-windows.sh` | Cross-compile Windows dari Linux |
| `electron-main/scripts/prebuild.js` | Auto-build backend sebelum package |
| `.github/workflows/build-electron.yml` | GitHub Actions workflow |
| `BUILD_GUIDE.md` | Detailed build documentation |

---

## Next Steps

### Immediate (untuk test)
```bash
# Build Linux AppImage saja
npm run dist:linux
```

### Short-term
1. Setup GitHub Actions (sudah buat file, tinggal push ke GitHub)
2. Tag versi pertama untuk auto-build

### Long-term
- Monitor GitHub Actions builds
- Release production ke GitHub Releases
- Update download links di website

---

## File Structure untuk Output

```
release/
├── ShoshaMart POS-0.1.0.AppImage        # Linux executable
├── ShoshaMart POS-0.1.0 Setup.exe       # Windows installer (dari Windows/GitHub Actions)
├── ShoshaMart POS-0.1.0.dmg             # macOS installer (dari GitHub Actions)
└── builder-effective-config.yaml        # Electron-builder config
```

---

**Last Updated**: December 11, 2025  
**Recommended**: GitHub Actions untuk production
