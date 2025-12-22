// updater.js
// Modul auto-update Electron
const { autoUpdater } = require('electron-updater')
const { dialog } = require('electron')

/**
 * setupAutoUpdate(mainWindow, stopCallback)
 * - mainWindow: BrowserWindow instance used for dialogs
 * - stopCallback: optional async function that attempts to stop background processes (like backend server)
 */
function setupAutoUpdate(mainWindow, stopCallback) {
  autoUpdater.checkForUpdatesAndNotify()

  autoUpdater.on('update-available', () => {
    dialog.showMessageBox(mainWindow, {
      type: 'info',
      title: 'Update Tersedia',
      message: 'Versi baru tersedia. Update akan diunduh otomatis.'
    })
  })

  autoUpdater.on('update-downloaded', async () => {
    const answer = await dialog.showMessageBox(mainWindow, {
      type: 'info',
      title: 'Update Siap',
      message: 'Update sudah siap. Aplikasi akan menutup dan menginstall update. Pastikan semua pekerjaan disimpan.',
      buttons: ['Install sekarang', 'Batal'],
      defaultId: 0,
      cancelId: 1,
    })

    if (answer.response === 0) {
      // Try to stop backend or other background processes before quitting
      if (typeof stopCallback === 'function') {
        try {
          // allow stopCallback to be async and wait up to a few seconds
          await Promise.race([
            stopCallback(),
            new Promise((res) => setTimeout(res, 3000)),
          ])
        } catch (e) {
          console.error('stopCallback failed:', e)
        }
      }

      // small delay to ensure file handles are released
      setTimeout(() => {
        try {
          autoUpdater.quitAndInstall()
        } catch (e) {
          console.error('quitAndInstall error:', e)
        }
      }, 800)
    }
  })
}

module.exports = { setupAutoUpdate }
