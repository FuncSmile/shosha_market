# Setup GitHub Release Token untuk Auto-Publish

## Masalah
Workflow build electron berhasil (`conclusion: success`) tetapi release tidak ter-create di GitHub. Ini terjadi karena `electron-builder --publish=onTag` memerlukan token dengan permission `repo` untuk membuat/update release dan upload assets (termasuk `latest.yml` untuk auto-updater).

## Root Cause
- Workflow menggunakan `${{ secrets.GH_RELEASE_TOKEN || secrets.GITHUB_TOKEN }}`
- Jika `GH_RELEASE_TOKEN` tidak ada, workflow fallback ke `GITHUB_TOKEN` (token default GitHub Actions)
- `GITHUB_TOKEN` default **tidak punya write permission untuk releases** pada workflow yang di-trigger oleh tag push
- electron-builder gagal publish secara silent (tidak error keluar, build tetap sukses)

## Solusi: Buat Personal Access Token (PAT)

### Langkah 1: Buat PAT di GitHub
1. Buka https://github.com/settings/tokens/new
2. **Note**: `ShoshaMart Release Publisher` (atau nama lain)
3. **Expiration**: Pilih durasi (rekomendasi: 90 days atau No expiration jika repo private)
4. **Scopes**: Centang **`repo`** (full control of private repositories)
   - Ini memberikan permission: `repo:status`, `repo_deployment`, `public_repo`, `repo:invite`, `security_events`
   - Yang diperlukan electron-builder: write access untuk releases dan assets
5. Klik **Generate token**
6. **PENTING**: Copy token yang muncul (hanya ditampilkan sekali)

### Langkah 2: Tambahkan Secret ke Repository
1. Buka repository di GitHub: https://github.com/FuncSmile/shosha_market
2. Klik **Settings** → **Secrets and variables** → **Actions**
3. Klik **New repository secret**
4. **Name**: `GH_RELEASE_TOKEN` (harus exact nama ini)
5. **Value**: Paste token yang di-copy tadi
6. Klik **Add secret**

### Langkah 3: Trigger Build Ulang
Setelah secret ditambahkan, push tag baru untuk trigger workflow:

```bash
cd /home/fad/Documents/myProject/shosha/shosha_mart
git tag -a v1.0.9 -m "Release v1.0.9 - with proper publish token"
git push origin v1.0.9
```

### Langkah 4: Verifikasi Publish Berhasil
Setelah workflow selesai (~5-10 menit), cek:

1. **Workflow log**: Buka Actions → pilih run untuk v1.0.9 → pilih job (Linux/Windows/macOS) → cek step "Build Electron App" → harus ada output "Publishing to GitHub Releases"

2. **Release page**: https://github.com/FuncSmile/shosha_market/releases/tag/v1.0.9
   - Harus ada release ter-create
   - Assets harus include: `.AppImage`, `.exe`, `.dmg`, `.blockmap`, **dan `latest.yml`** + `latest-linux.yml`

3. **Metadata availability**:
```bash
curl -I https://github.com/FuncSmile/shosha_market/releases/download/v1.0.9/latest.yml
# Expected: HTTP/2 200
curl -I https://github.com/FuncSmile/shosha_market/releases/download/v1.0.9/latest-linux.yml
# Expected: HTTP/2 200
```

4. **Test updater**: Jalankan aplikasi dari release sebelumnya (v1.0.1 atau v1.0.4) → updater harus detect v1.0.9 available dan offer to download.

## Troubleshooting

### Token masih tidak bekerja
- Pastikan nama secret **exact**: `GH_RELEASE_TOKEN` (case-sensitive)
- Pastikan scope token include `repo` (bukan hanya `public_repo`)
- Cek workflow log: harus ada "GH_TOKEN is set (length: XXX)" di output

### Publish gagal dengan 403/401
- Token expired atau revoked → regenerate dan update secret
- Token tidak punya access ke repository → pastikan token dari account yang punya write access ke repo

### Release ter-create tapi latest.yml missing
- Cek log electron-builder untuk error saat upload metadata
- Pastikan workflow tidak punya step lain yang overwrite release (sudah dihapus di commit sebelumnya)

## Alternative: Gunakan GITHUB_TOKEN dengan Permission Update

Jika tidak mau pakai PAT, update workflow file (`.github/workflows/build-electron.yml`) untuk memberikan write permission:

```yaml
name: Build Electron App (Multi-Platform)

on:
  push:
    branches: [ main, develop, fadli/dev ]
    tags:
      - 'v*'
  pull_request:
    branches: [ main, develop ]
  workflow_dispatch:

permissions:
  contents: write  # Allow creating releases and uploading assets

jobs:
  build-linux:
    # ... rest of workflow
```

Kemudian gunakan `GITHUB_TOKEN` di env (tanpa perlu PAT):
```yaml
- name: Build Electron App (Linux)
  env:
    GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    cd electron-main
    npx electron-builder --linux AppImage --publish=onTag
```

**Catatan**: Approach ini lebih aman (token auto-expire per run) tapi butuh update workflow file.

## Hasil Akhir yang Diharapkan

Setelah setup token dan push tag baru:
- ✅ Workflow build sukses
- ✅ Release ter-create di GitHub dengan semua assets
- ✅ `latest.yml` dan platform metadata available via HTTP 200
- ✅ Electron app di production dapat fetch update dan download/install otomatis
- ✅ Tidak ada lagi error 404 untuk `latest-linux.yml` di runtime
