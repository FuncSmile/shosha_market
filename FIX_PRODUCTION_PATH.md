# Fix: Production AppImage Path Issue

## Problem
```
Not allowed to load local resource: file:///tmp/.mount_ShoshavA9a6H/resources/renderer/dist/index.html
```

Error terjadi karena:
1. `renderer/dist` tidak ter-bundle dengan benar ke dalam `app.asar`
2. Path loading di `main.js` salah untuk production build

## Root Cause

### Before Fix
**electron-main/package.json**:
```json
"files": [
  "**/*",
  "../renderer/dist/**/*"  // ❌ Path relatif tidak work untuk electron-builder
]
```

**electron-main/main.js**:
```javascript
const distPath = path.join(__dirname, '../renderer/dist/index.html')  // ❌ Hanya work untuk development
```

### Issue
- `renderer/dist` tidak masuk ke `app.asar`
- Path hanya work untuk development, bukan production
- Electron mencari file di `/tmp/.mount_XXX/resources/renderer/dist/` tapi tidak ada

## Solution

### 1. Fix package.json
```json
"files": [
  "main.js",
  "package.json",
  {
    "from": "../renderer/dist",
    "to": "renderer/dist",
    "filter": ["**/*"]
  }
]
```

**Penjelasan**:
- Explicitly copy `renderer/dist` ke dalam asar
- Use object notation dengan `from` dan `to`
- Semua file masuk ke `app.asar/renderer/dist/`

### 2. Fix main.js
```javascript
async function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    backgroundColor: '#0b1629',
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
    },
  })

  const devURL = process.env.VITE_DEV_SERVER_URL || 'http://localhost:5173'
  
  // ✅ Determine correct path based on packaging
  let distPath
  if (app.isPackaged) {
    // In packaged app, files are in app.asar at /renderer/dist/
    distPath = path.join(__dirname, 'renderer', 'dist', 'index.html')
  } else {
    // In development
    distPath = path.join(__dirname, '../renderer/dist/index.html')
  }

  let loaded = false

  // Try dev server first if in dev mode
  if (process.env.ELECTRON_DEV) {
    try {
      await win.loadURL(devURL)
      loaded = true
      console.log('✓ Loaded from dev server:', devURL)
    } catch (err) {
      console.error('✗ Dev server not reachable, fallback to dist:', err?.message || err)
    }
  }

  // Load from file if dev server failed or not in dev mode
  if (!loaded) {
    console.log('Loading from file:', distPath)
    console.log('  __dirname:', __dirname)
    console.log('  app.isPackaged:', app.isPackaged)
    
    try {
      await win.loadFile(distPath)
      console.log('✓ Successfully loaded from:', distPath)
    } catch (err) {
      console.error('✗ Failed to load file:', err)
      console.error('  Attempted path:', distPath)
      console.error('  File exists:', fs.existsSync(distPath))
    }
  }
}
```

**Key Changes**:
- Check `app.isPackaged` to determine environment
- Production: `__dirname/renderer/dist/index.html` (inside asar)
- Development: `__dirname/../renderer/dist/index.html` (file system)
- Added logging untuk debug

## Verification

### Check asar contents
```bash
npx @electron/asar extract release/linux-unpacked/resources/app.asar /tmp/test
find /tmp/test -name "index.html"
```

**Expected**:
```
/tmp/test/renderer/dist/index.html  ✅
```

### Test AppImage
```bash
./release/ShoshaMart\ POS-0.1.0.AppImage
```

**Expected Log**:
```
Loading from file: /tmp/.mount_ShoshalpHYpe/resources/app.asar/renderer/dist/index.html
  __dirname: /tmp/.mount_ShoshalpHYpe/resources/app.asar
  app.isPackaged: true
✓ Successfully loaded from: /tmp/.mount_ShoshalpHYpe/resources/app.asar/renderer/dist/index.html
```

## File Structure

### Development
```
electron-main/
├── main.js
└── (loads from) ../renderer/dist/index.html
```

### Production (Packaged)
```
app.asar/
├── main.js
├── package.json
└── renderer/
    └── dist/
        ├── index.html          ✅
        ├── assets/
        │   ├── index-*.css
        │   └── index-*.js
        └── ...
```

## Backend Path (Bonus Fix)

Also fixed backend binary path untuk production:

```javascript
async function spawnBackend() {
  const port = Number(process.env.BACKEND_PORT || 8080)

  const free = await isPortFree(port)
  if (!free) {
    console.log(`Backend already running on port ${port}, skip spawn.`)
    return
  }

  let backendCwd
  let backendBinary

  if (app.isPackaged) {
    // In production, backend is in resources/backend
    backendCwd = path.join(process.resourcesPath, 'backend')
    backendBinary = process.platform === 'win32' ? 'server.exe' : 'server'
    
    // Fallback ke resources/app/backend jika tidak ada
    if (!fs.existsSync(path.join(backendCwd, backendBinary))) {
      backendCwd = path.join(process.resourcesPath, 'app', 'backend')
    }
  } else {
    // In development
    backendCwd = path.join(__dirname, '../backend')
    backendBinary = process.platform === 'win32' ? 'server.exe' : 'server'
  }

  const backendPath = path.join(backendCwd, backendBinary)
  
  console.log('Backend path:', backendPath)
  console.log('Backend exists:', fs.existsSync(backendPath))

  const useGoRun =
    process.env.BACKEND_DEV === 'true' ||
    process.env.ELECTRON_DEV ||
    !fs.existsSync(backendPath)

  let cmd, args

  if (useGoRun) {
    cmd = 'go'
    args = ['run', '.']
    console.log('Using go run for development')
  } else {
    cmd = backendPath
    args = []
    console.log('Using compiled backend:', cmd)
  }

  goProcess = spawn(cmd, args, { cwd: backendCwd, env: process.env })
  // ... (rest of code)
}
```

## Testing Checklist

- [x] Development mode works (`npm start`)
- [x] Production AppImage loads correctly
- [x] Frontend loads from asar
- [x] Backend binary executes
- [x] No "Not allowed to load local resource" error
- [x] Console shows "✓ Successfully loaded from: ..."

## Files Changed

1. **electron-main/package.json** - Fixed `files` configuration
2. **electron-main/main.js** - Fixed path detection for production

## Result

✅ **FIXED**  
- AppImage size: 145 MB (includes renderer/dist)
- Frontend loads successfully from asar
- Backend spawns correctly
- No path errors

---

**Date**: December 11, 2025  
**Status**: Production ready ✅
