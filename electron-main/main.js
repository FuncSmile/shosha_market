const { app, BrowserWindow } = require('electron')
const { spawn } = require('child_process')
const path = require('path')
const fs = require('fs')
const net = require('net')

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

  const devURL = process.env.VITE_DEV_SERVER_URL || 'http://localhost:5173'
  const distPath = path.join(__dirname, '../renderer/dist/index.html')
  let loaded = false

  if (process.env.ELECTRON_DEV) {
    try {
      await win.loadURL(devURL)
      loaded = true
    } catch (err) {
      console.error('Dev server not reachable, fallback to dist:', err?.message || err)
    }
  }

  if (!loaded) {
    await win.loadFile(distPath)
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
  const backendCwd = path.join(__dirname, '../backend')
  const compiled = process.platform === 'win32' ? 'server.exe' : './server'
  const compiledPath = path.join(backendCwd, compiled)
  const port = Number(process.env.BACKEND_PORT || 8080)

  const free = await isPortFree(port)
  if (!free) {
    console.log(`Backend already running on port ${port}, skip spawn.`)
    return
  }

  const useGoRun =
    process.env.BACKEND_DEV === 'true' ||
    process.env.ELECTRON_DEV ||
    !fs.existsSync(compiledPath)

  const cmd = useGoRun ? 'go' : compiled
  const args = useGoRun ? ['run', '.'] : []

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
