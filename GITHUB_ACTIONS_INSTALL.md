# GitHub Actions Build & Installation Guide

## 📦 Cara Build via GitHub Actions

### Step 1: Push Code ke GitHub

```bash
# Pastikan di branch yang benar
git checkout fadli/dev

# Add semua perubahan
git add .
git commit -m "Ready for production build"

# Push ke GitHub
git push origin fadli/dev
```

### Step 2: Buat Release Tag

```bash
# Buat tag untuk versi (misal v0.1.0)
git tag v0.1.0

# Push tag ke GitHub
git push origin v0.1.0
```

### Step 3: GitHub Actions Otomatis Build

Setelah push tag, GitHub Actions akan **otomatis**:

1. ✅ Build **Linux AppImage** (di Ubuntu runner)
2. ✅ Build **Windows NSIS Installer** (di Windows runner - **CGO otomatis enabled!**)
3. ✅ Build **macOS DMG** (di macOS runner)

**Tidak perlu install apapun!** GitHub Actions sudah provide semua environment.

---

## 🔍 Cek Build Progress

### Via GitHub Website

1. Buka repository: `https://github.com/FuncSmile/shosha_market`
2. Klik tab **Actions**
3. Lihat workflow **"Build Electron App (Multi-Platform)"**
4. Klik pada workflow run yang sedang jalan
5. Lihat progress untuk setiap platform:
   - `build-linux` - Linux AppImage
   - `build-windows` - Windows NSIS ✅
   - `build-macos` - macOS DMG

### Build Time
- **Linux**: ~5-8 menit
- **Windows**: ~8-12 menit
- **macOS**: ~10-15 menit

**Total**: ~15-20 menit untuk semua platform

---

## 📥 Download Installer Windows

### Option 1: Dari GitHub Releases (Untuk User)

Setelah build selesai:

1. Buka: `https://github.com/FuncSmile/shosha_market/releases`
2. Klik pada release tag (misal: `v0.1.0`)
3. Scroll ke **Assets**
4. Download file:
   ```
   ShoshaMart-POS-0.1.0-Setup.exe  (~100 MB)
   ```

### Option 2: Dari Artifacts (Untuk Development)

Jika tidak buat release tag (hanya push biasa):

1. Buka: `https://github.com/FuncSmile/shosha_market/actions`
2. Klik workflow run terbaru
3. Scroll ke **Artifacts**
4. Download:
   ```
   windows-installer
   ```
5. Extract ZIP, di dalamnya ada:
   ```
   ShoshaMart-POS-0.1.0-Setup.exe
   ShoshaMart-POS-0.1.0-Setup.exe.blockmap
   ```

---

## 💻 Instalasi di Windows

### Step 1: Download Installer

Download `ShoshaMart-POS-0.1.0-Setup.exe` dari GitHub Releases

### Step 2: Run Installer

1. **Double-click** `ShoshaMart-POS-0.1.0-Setup.exe`
2. Windows mungkin tanya **"Do you want to allow this app?"** → Klik **Yes**
3. Installer akan muncul:
   - Pilih direktori install (default: `C:\Users\<user>\AppData\Local\Programs\shosha-mart\`)
   - Klik **Install**
4. Tunggu instalasi selesai (~30 detik)
5. Klik **Finish**

### Step 3: Jalankan Aplikasi

**Option A - Via Desktop Shortcut**:
- Double-click icon **ShoshaMart POS** di Desktop

**Option B - Via Start Menu**:
- Tekan `Win` key
- Ketik "ShoshaMart"
- Klik **ShoshaMart POS**

**Option C - Via Executable**:
```
C:\Users\<username>\AppData\Local\Programs\shosha-mart\ShoshaMart POS.exe
```

---

## 🔧 What's Inside the Installer

Windows NSIS installer includes:

```
ShoshaMart POS/
├── ShoshaMart POS.exe        # Main Electron app
├── resources/
│   ├── app.asar              # Frontend (Vue app)
│   └── backend/
│       ├── server.exe        # Go backend (dengan CGO!) ✅
│       ├── .env              # Environment variables
│       └── offline.db        # SQLite database
└── (Electron runtime files)
```

---

## ✅ Verification

### Check App Berjalan

Setelah install dan jalankan:

1. **Window terbuka** → Frontend Vue loaded ✅
2. **Backend otomatis start** → Port 8080 listening ✅
3. **Database terkoneksi** → SQLite berfungsi ✅

### Check Console (Development)

Jika ada masalah, cek console:

1. Buka app
2. Tekan `Ctrl + Shift + I` (Developer Tools)
3. Lihat tab **Console**

**Expected log**:
```
Loading from file: C:\Users\...\resources\app.asar\renderer\dist\index.html
Backend path: C:\Users\...\resources\backend\server.exe
Backend exists: true
✓ Successfully loaded from: ...
[go] 2025/12/11 ... sidecar listening on 0.0.0.0:8080
```

**NOT**:
```
Binary was compiled with 'CGO_ENABLED=0'  ❌
```

---

## 🔄 Update Aplikasi

### Untuk Release Baru

1. User **uninstall** versi lama:
   - Settings → Apps → ShoshaMart POS → Uninstall

2. Download installer versi baru dari GitHub Releases

3. Install seperti biasa

### Auto-Update (Future Enhancement)

Bisa implement `electron-updater` untuk auto-update tanpa uninstall manual.

---

## 📊 GitHub Actions Workflow Details

### Build Process

```yaml
build-windows:
  runs-on: windows-latest  # ✅ Native Windows build

  steps:
    - Setup Node.js 20
    - Setup Go 1.23
    - Install npm dependencies (renderer + electron-main)
    - Build backend:
        go build -o server.exe main.go  # ✅ CGO enabled by default
    - Build renderer:
        npm run build (Vue + Vite)
    - Package Electron:
        electron-builder --win nsis
    - Upload artifact
    - Create release (if tag)
```

### Advantages

✅ **CGO otomatis enabled** (native Windows build)
✅ **Tidak perlu mingw-w64**
✅ **Clean build environment**
✅ **Reproducible builds**
✅ **Multi-platform support**

---

## 🎯 Quick Reference

### For Developer (You)

```bash
# 1. Push code
git push origin fadli/dev

# 2. Create release tag
git tag v0.1.0
git push origin v0.1.0

# 3. Wait for GitHub Actions (~15 min)
# 4. Download dari GitHub Releases
```

### For End User (Windows)

```
1. Buka: https://github.com/FuncSmile/shosha_market/releases
2. Download: ShoshaMart-POS-0.1.0-Setup.exe
3. Run installer
4. Done! App siap pakai
```

---

## 🐛 Troubleshooting

### Issue: GitHub Actions Gagal Build

**Check logs**:
1. Actions tab → Click failed workflow
2. Klik `build-windows` job
3. Expand step yang error
4. Fix code, commit, push ulang

### Issue: Installer Tidak Bisa Dijalankan

**Windows SmartScreen**:
1. Klik **"More info"**
2. Klik **"Run anyway"**

(Untuk production, butuh **code signing certificate** agar tidak muncul warning)

### Issue: App Tidak Bisa Start

**Check antivirus**:
- Beberapa antivirus block `.exe` yang tidak di-sign
- Add exception untuk `ShoshaMart POS.exe`

---

## 📝 Example Workflow

### Scenario: Release v0.1.0

```bash
# Terminal (Linux/macOS)
cd /home/fad/Documents/myProject/shosha/shosha_mart

# Commit perubahan terakhir
git add .
git commit -m "Release v0.1.0 - Production ready"
git push origin fadli/dev

# Buat dan push tag
git tag v0.1.0
git push origin v0.1.0

# GitHub Actions starts building...
# Wait ~15 minutes

# Check progress:
# https://github.com/FuncSmile/shosha_market/actions

# After success:
# https://github.com/FuncSmile/shosha_market/releases/tag/v0.1.0
```

### User Downloads

```
1. User buka GitHub Releases
2. Download ShoshaMart-POS-0.1.0-Setup.exe
3. Run installer
4. App installed di C:\Users\<user>\AppData\Local\Programs\shosha-mart\
5. Shortcut created di Desktop
6. App ready to use!
```

---

## 🚀 Production Checklist

Sebelum release ke user:

- [ ] Test di Linux (AppImage)
- [ ] Test di Windows (NSIS installer via GitHub Actions)
- [ ] Verify backend CGO works (no stub error)
- [ ] Verify database persists
- [ ] Verify .env loaded correctly
- [ ] Update version di package.json
- [ ] Create git tag
- [ ] Push to GitHub
- [ ] Wait for GitHub Actions
- [ ] Download & test installers
- [ ] Create GitHub Release with notes
- [ ] Share download link dengan users

---

**Generated**: December 11, 2025  
**Method**: GitHub Actions (Recommended for Production)  
**No local tools needed**: Everything di-build di cloud! ☁️
