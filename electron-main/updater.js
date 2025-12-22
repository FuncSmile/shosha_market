// updater.js
// Modul auto-update Electron
const { autoUpdater } = require('electron-updater')
const { dialog } = require('electron')

function setupAutoUpdate(mainWindow) {
  autoUpdater.checkForUpdatesAndNotify()

  autoUpdater.on('update-available', () => {
    dialog.showMessageBox(mainWindow, {
      type: 'info',
      title: 'Update Tersedia',
      message: 'Versi baru tersedia. Update akan diunduh otomatis.'
    })
  })

  autoUpdater.on('update-downloaded', () => {
    dialog.showMessageBox(mainWindow, {
      type: 'info',
      title: 'Update Siap',
      message: 'Update sudah siap. Aplikasi akan restart untuk install update.'
    }).then(() => {
      autoUpdater.quitAndInstall()
    })
  })
}

module.exports = { setupAutoUpdate }
