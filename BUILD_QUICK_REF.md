# Build Quick Reference

## 🎯 Current Status
```
Linux ............ ✅ READY (AppImage built: 132 MB)
Windows .......... ⚠️ Needs Wine or GitHub Actions
macOS ............ ⚠️ Needs macOS or GitHub Actions
```

## 📦 Available Outputs
```
release/
├── ShoshaMart POS-0.1.0.AppImage ✅ Ready to distribute
```

---

## 🚀 Quick Build Commands

### Test Linux Build
```bash
cd electron-main
npm run dist:linux
```

### Build All (on current OS)
```bash
cd electron-main
npm run dist
```

### Build Specific OS
```bash
npm run dist:linux      # Linux AppImage
npm run dist:win        # Windows NSIS (requires Wine on Linux)
```

### Use Build Script
```bash
bash build.sh           # Auto-detect OS and build
bash build-windows.sh   # Build Windows binary + try electron-builder
```

---

## 🐳 Build in Docker

```bash
docker build -t shosha-builder -f Dockerfile.electron .
docker run -v $(pwd)/release:/app/release shosha-builder \
  sh -c "cd electron-main && npm run dist:linux"
```

---

## 🔧 Common Issues & Fixes

| Issue | Fix |
|-------|-----|
| Wine required | `sudo pacman -S wine wine-mono wine-gecko` |
| Port already in use | Change `BACKEND_PORT` env var |
| Build cache issues | `rm -rf node_modules package-lock.json && npm install` |
| Database locked | Delete `backend/offline.db` and rebuild |

---

## 📊 File Sizes
```
Linux AppImage .... 132 MB (includes Electron + Go backend + Vue app)
Windows NSIS ...... ~100 MB (similar structure)
macOS DMG ........ ~120 MB (similar structure)
```

---

## ✅ Tested & Verified
- [x] Linux AppImage builds successfully
- [x] Backend binary included in package
- [x] Frontend bundled with Vite
- [x] Scripts created for automation
- [x] GitHub Actions workflow ready
- [x] Docker support available

---

## 📖 Full Documentation
- `BUILD_GUIDE.md` - Detailed step-by-step
- `BUILD_SOLUTIONS.md` - Strategy comparison
- `BUILD_RESULTS.md` - Current build status
- `.github/workflows/build-electron.yml` - GitHub Actions

---

**Generated**: Dec 11, 2025  
**Next Step**: Push to GitHub → enable Actions → tag release
