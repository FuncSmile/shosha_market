const { app, BrowserWindow } = require('electron')
const { spawn } = require('child_process')
const path = require('path')
const fs = require('fs')
const net = require('net')
const { setupAutoUpdate } = require('./updater')

let goProcess

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

  // Setup auto-update
  setupAutoUpdate(win)

  const devURL = process.env.VITE_DEV_SERVER_URL || 'http://localhost:5173'
  
  // Determine correct path based on whether app is packaged
  let distPath
  if (app.isPackaged) {
    // In packaged app, files are in app.asar or app.asar directory
    // renderer/dist is bundled inside asar at /renderer/dist/index.html
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

function isPortFree(port) {
  return new Promise((resolve) => {
    const tester = net
      .createServer()
      .once('error', () => resolve(false))
      .once('listening', () => tester.close(() => resolve(true)))
      .listen(port, '127.0.0.1')
  })
}

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
    
    // Fallback ke resources/app/backend jika tidak ada di resources/backend
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
    if (process.platform === 'win32') {
      cmd = backendPath
      args = []
    } else {
      cmd = backendPath
      args = []
    }
    console.log('Using compiled backend:', cmd)
  }

  goProcess = spawn(cmd, args, { cwd: backendCwd, env: process.env })

  goProcess.stdout.on('data', (data) => console.log(`[go] ${data}`.trim()))
  goProcess.stderr.on('data', (data) => console.error(`[go err] ${data}`.trim()))
  goProcess.on('close', (code) => console.log(`go backend exited with code ${code}`))
}

app.whenReady().then(() => {
  spawnBackend()
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('before-quit', () => {
  if (goProcess) {
    goProcess.kill()
  }
})
