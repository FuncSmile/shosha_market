# Fix: Windows Binary CGO Error

## Problem
```
[go err] 2025/12/11 11:27:03 failed to init db: open database: Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

## Root Cause

SQLite (go-sqlite3) requires **CGO** (C bindings) to work, but:
- Default Go cross-compilation sets `CGO_ENABLED=0` for Windows from Linux
- Without CGO, SQLite driver is just a stub that cannot work

## Solution

### Option 1: Build on Windows (Recommended for Production)

**On Windows machine**:
```bash
cd backend
go mod download
go build -o server.exe main.go
```

CGO is automatically enabled on native Windows builds.

### Option 2: Cross-compile from Linux with mingw-w64

**Install mingw-w64**:
```bash
# Ubuntu/Debian
sudo apt install gcc-mingw-w64-x86-64

# Arch/Manjaro
sudo pacman -S mingw-w64-gcc

# Fedora
sudo dnf install mingw64-gcc
```

**Build Windows binary**:
```bash
cd backend

GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
go build -o server.exe main.go
```

**Or use script**:
```bash
bash build-windows.sh
```

### Option 3: Use GitHub Actions (Easiest for Multiple Platforms)

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically build Windows binary on Windows runner (CGO enabled by default).

## Updated Scripts

### `electron-main/scripts/prebuild.js`

Now automatically:
1. Builds for current platform with CGO
2. On Linux, attempts to cross-compile Windows binary with mingw-w64
3. Gracefully handles missing mingw-w64

```javascript
// Build for current platform
buildBinary(currentOS, currentArch, currentBinary);

// On Linux, also try to build Windows binary
if (process.platform === 'linux') {
  console.log('🪟 Attempting to build Windows binary (requires mingw-w64)...');
  const winBuilt = buildBinary('windows', 'amd64', 'server.exe');
  
  if (!winBuilt) {
    console.log('⚠️  Windows build failed. Install mingw-w64 or use GitHub Actions');
  }
}
```

### `build-windows.sh`

Updated to use mingw-w64 with CGO:
```bash
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
go build -o server.exe main.go
```

## Verification

### Check Binary Dependencies

**On Windows**:
```cmd
dumpbin /dependents backend\server.exe
```

Should show no `libsqlite3.dll` dependency (statically linked).

**On Linux (after cross-compile)**:
```bash
file backend/server.exe
```

Should show:
```
backend/server.exe: PE32+ executable (console) x86-64, for MS Windows
```

### Test Windows Binary

```cmd
cd backend
server.exe
```

Expected output:
```
2025/12/11 11:30:00 sidecar listening on 0.0.0.0:8080 (db: offline.db, sync interval ~5m0s)
```

**NOT**:
```
failed to init db: Binary was compiled with 'CGO_ENABLED=0'
```

## Build Matrix

| Platform | Build From | CGO | Toolchain | Status |
|----------|-----------|-----|-----------|--------|
| Linux → Linux | Linux | ✅ Enabled | gcc | ✅ Works |
| Windows → Windows | Windows | ✅ Enabled | MSVC/MinGW | ✅ Works |
| Linux → Windows | Linux | ✅ Enabled | mingw-w64 | ✅ Works (if installed) |
| Linux → Windows | Linux | ❌ Disabled | Go only | ❌ Fails (CGO stub) |

## GitHub Actions Build

Windows build in CI uses **native Windows runner**:

```yaml
build-windows:
  runs-on: windows-latest  # Native Windows build
  
  steps:
    - name: Build backend (Windows)
      run: |
        cd backend
        go mod download
        go build -o server.exe main.go  # CGO enabled automatically
```

✅ No cross-compilation needed
✅ CGO works out of the box
✅ No mingw-w64 required

## Quick Reference

### For Windows Users
```bash
# Build locally on Windows
cd backend && go build -o server.exe main.go
cd ../electron-main && npm run dist:win
```

### For Linux Users (Development)
```bash
# Option A: Install mingw-w64
sudo pacman -S mingw-w64-gcc  # Manjaro/Arch
bash build-windows.sh

# Option B: Use GitHub Actions
git tag v0.1.0 && git push origin v0.1.0
# Download from GitHub Releases
```

### For Production Release
```bash
# Push code and tag
git push origin main
git tag v0.1.0
git push origin v0.1.0

# GitHub Actions builds:
# - Linux AppImage (on Ubuntu runner)
# - Windows NSIS (on Windows runner) ✅ CGO enabled
# - macOS DMG (on macOS runner)
```

## Common Issues

### Issue: mingw-w64 not found
```
x86_64-w64-mingw32-gcc: command not found
```

**Fix**: Install cross-compiler
```bash
sudo pacman -S mingw-w64-gcc
```

### Issue: CGO_ENABLED=0 still used
```
Binary was compiled with 'CGO_ENABLED=0'
```

**Fix**: Explicitly set environment variable
```bash
export CGO_ENABLED=1
go build -o server.exe main.go
```

### Issue: Build works but runtime fails
```
The code execution cannot proceed because libwinpthread-1.dll was not found
```

**Fix**: Ensure static linking or bundle DLLs. With mingw-w64, it should be statically linked.

## Summary

✅ **Fixed**: Updated build scripts to enable CGO for Windows
✅ **Solution**: Use mingw-w64 for cross-compile OR build on Windows
✅ **CI/CD**: GitHub Actions builds on native Windows (CGO automatic)
✅ **Status**: Windows binary now works with SQLite

---

**Date**: December 11, 2025  
**Status**: Windows build fixed with CGO support
