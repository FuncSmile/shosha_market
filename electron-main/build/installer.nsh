; installer.nsh - NSIS include to stop running app and backend before uninstall/install
; This script will be included in NSIS generated installer by electron-builder

!macro preuninstall
  ; Attempt to stop the running app and backend gracefully on Windows
  ; Use taskkill as a fallback to ensure files are unlocked

  ; Try to close the main application window by process name
  ExecWait 'taskkill /IM "ShoshaMart POS.exe" /T /F' ; ignore errors
  ExecWait 'taskkill /IM "server.exe" /T /F' ; backend binary
  ExecWait 'taskkill /IM "electron.exe" /T /F' ; generic electron

  ; Small pause to allow processes to terminate and release handles
  Sleep 800
!macroend

; Also add a hook to stop processes before install when upgrading
!macro preinstall
  ExecWait 'taskkill /IM "ShoshaMart POS.exe" /T /F'
  ExecWait 'taskkill /IM "server.exe" /T /F'
  ExecWait 'taskkill /IM "electron.exe" /T /F'
  Sleep 600
!macroend
