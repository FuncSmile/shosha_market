Arsitektur Proyek POS Desktop Offline-First (Electron, Vue 3, Go) – Sidecar Pattern

Aplikasi POS (Point of Sale) desktop offline-first dirancang agar tetap berfungsi tanpa koneksi internet, dengan sinkronisasi data ke server saat online. Arsitektur yang diusulkan memisahkan Frontend (Electron + Vue 3) dan Backend (service sidecar lokal berbasis Go) dalam proses terisolasi. Pola Sidecar ini menjaga UI tetap responsif dan modular, sementara logika bisnis dan penyimpanan data berlangsung di service backend terpisah
reddit.com
. Berikut adalah struktur proyek, alur komunikasi, mekanisme sinkronisasi offline-online, kode boilerplate awal, dan implementasi fitur-fitur utama POS.
Struktur Folder Proyek

Struktur direktori dibagi sesuai peran: Main Process Electron, Renderer Process (Vue 3 + TailwindCSS & ShadCN Vue), dan Backend Sidecar (Go + Gin + GORM). Contoh struktur folder proyek:
```
pos-app/
├── backend/              # Service sidecar backend (Go + Gin + GORM)
│   ├── main.go           # Entrypoint Go (start Gin server, define routes)
│   ├── go.mod            # Go module definitions (incl. Gin, GORM, drivers)
│   ├── config/           # Config (e.g. DB connection settings)
│   ├── models/           # GORM models (Barang, Cabang, Penjualan, dsb.)
│   ├── controllers/      # Handler untuk setiap route API (CRUD logic)
│   ├── sync/             # Modul sinkronisasi (antrean sync, dsb.)
│   └── ... (files lainnya: service, util, dll.)
├── electron-main/        # Kode Main Process (Electron)
│   ├── main.js           # Skrip utama Electron (buat jendela, spawn backend)
│   ├── preload.js        # (Opsional) Preload script untuk konteks IPC
│   └── package.json      # Dependensi & konfigurasi Electron (Main Process)
├── renderer/             # Kode Renderer Process (Vue 3 + TailwindCSS + ShadCN)
│   ├── package.json      # Dependensi frontend (Vue 3, libraries UI)
│   ├── vite.config.js    # (Contoh) konfigurasi build Vite
│   ├── tailwind.config.js# konfigurasi TailwindCSS
│   └── src/              # Source code Vue (UI)
│       ├── main.js       # Entrypoint aplikasi Vue (mount App ke DOM)
│       ├── App.vue       # Root component Vue
│       ├── components/   # Komponen Vue (contoh: form Barang, dsb.)
│       ├── pages/        # Halaman Vue (contoh: halaman Penjualan, Laporan)
│       └── ipc.js        # (Opsional) wrapper IPC renderer untuk memanggil backend
└── README.md             # Dokumentasi atau instruksi proyek
```
Penjelasan: Folder backend/ berisi service Go (Gin) yang menjalankan web server REST lokal. Folder electron-main/ memuat kode main process Electron yang menginisiasi jendela aplikasi dan menjalankan service backend sebagai proses terpisah. Folder renderer/ memuat aplikasi Vue 3 (sebagai renderer process) lengkap dengan konfigurasi TailwindCSS dan komponen ShadCN Vue untuk antarmuka modern. Pemisahan ini memastikan concern terorganisir: UI dan pengalaman pengguna ditangani Vue/Electron, sedangkan logika domain, akses database lokal, dan sinkronisasi ditangani oleh service Go sidecar.
Alur Komunikasi Frontend–Backend Lokal

Frontend (Renderer Vue dalam Electron) berkomunikasi dengan backend sidecar Go yang berjalan lokal. Main process Electron akan menjalankan (spawn) service Go saat aplikasi mulai, dan mematikannya saat aplikasi ditutup
medium.com
medium.com
. Service Go (Gin) tersebut akan mendengarkan permintaan pada port lokal (misalnya 127.0.0.1:8080). Agar aman di desktop, bind alamat 127.0.0.1 sehingga hanya bisa diakses secara lokal (menghindari pop-up firewall Windows dan akses dari luar)
reddit.com
. Pastikan port default (misal 8080) tidak bentrok; main process dapat mengecek ketersediaan port atau menggunakan port konfigurasi.

Opsi Komunikasi: Komunikasi dapat dilakukan dengan dua pendekatan, IPC maupun HTTP/REST:

- Via REST API lokal: Pendekatan sederhana adalah menganggap backend Go sebagai server API lokal. Aplikasi Vue dapat langsung melakukan request HTTP (misalnya dengan fetch atau Axios) ke endpoint REST Gin (contoh: GET http://localhost:8080/api/barang)
    stackoverflow.com
-  Karena aplikasi berjalan di Electron (bukan browser murni), kita dapat mengizinkan request ke localhost tanpa terhalang CORS (misalnya dengan mengatur properti webSecurity atau menambahkan header CORS di Gin jika diperlukan). Stack Overflow merekomendasikan model ini: perlakukan Electron seperti aplikasi web biasa yang memanggil backend melalui API
    . Dengan cara ini, Vue (renderer) dapat langsung memanfaatkan API menggunakan Axios/Fetch, dan backend merespons dalam format JSON.

    Via IPC + Main Process: Alternatifnya, gunakan IPC (Inter-Process Communication) Electron. Renderer mengirim pesan/permintaan ke main process, lalu main process meneruskan ke backend. Contoh pola: main process mendaftarkan handler IPC (ipcMain.handle) untuk event seperti "get-barang", yang ketika dipanggil akan melakukan request HTTP (via Axios) ke localhost:8080/api/barang, lalu mengembalikan data ke renderer
    medium.com
    medium.com
    . Renderer (Vue) cukup memanggil window.api.invoke('get-barang') untuk mendapatkan data. Pola ini digunakan misalnya dalam aplikasi Electron + Go oleh 최우진
    medium.com
    medium.com
    . Keuntungan metode ini: main process dapat mengelola detail seperti manajemen session cookie atau error global, serta meningkatkan keamanan (karena renderer tidak memanggil HTTP langsung, mencegah manipulasi dari DevTools). Namun, pendekatan IPC lebih kompleks. Untuk kebutuhan POS offline-first yang tidak terlalu rumit soal autentikasi, opsi REST langsung sudah memadai dan lebih mudah.

Inisiasi Backend: Kode main process akan mengeksekusi binary Go saat startup. Contoh (pseudo-code) di electron-main/main.js:

const { app, BrowserWindow } = require('electron');
const { spawn } = require('child_process');
let goProcess;

app.whenReady().then(() => {
  // Jalankan service Go (misal terkompilasi sebagai binary `server` di folder backend)
  goProcess = spawn(process.platform === 'win32' ? 'server.exe' : './server', [], { cwd: __dirname + '/../backend' });
  goProcess.stdout.on('data', data => console.log(`Go: ${data}`));
  goProcess.stderr.on('data', data => console.error(`Go Err: ${data}`));

  // Buat jendela Electron dan muat UI Vue
  const win = new BrowserWindow({ width: 1200, height: 800, webPreferences: { contextIsolation: true } });
  win.loadURL('http://localhost:5173'); // URL dev server Vue (saat development)
  // win.loadFile(path.join(__dirname, '../renderer/dist/index.html')); // (contoh) file HTML produksi
});
app.on('before-quit', () => { if (goProcess) goProcess.kill(); });

Kode di atas akan menjalankan file binary Go server sebagai child process. Dengan ini, Go backend mulai otomatis saat aplikasi Electron dijalankan
medium.com
. Kita juga menangkap output log/error Go ke console Electron untuk kemudahan debug. BrowserWindow Electron memuat UI Vue – dalam mode development kita bisa arahkan ke server dev (localhost:5173 jika menggunakan Vite), sedangkan dalam produksi diarahkan ke file HTML hasil build.

Dengan arsitektur ini, alur komunikasi umumnya sebagai berikut:

    UI Action (Frontend) – Pengguna berinteraksi dengan UI (misal klik “Tambah Barang” atau “Checkout”). Vue akan memanggil fungsi yang melakukan request ke backend lokal (baik via HTTP fetch maupun via IPC ke main process).

    Local API Request – Permintaan dikirim ke Gin di service Go (contoh: endpoint /api/penjualan untuk proses checkout).

    Backend Processing – Backend Go menerima request, memproses logika (misal validasi, hitung total, update stok di SQLite), lalu mengembalikan response (misal data berhasil disimpan atau data yang diminta).

    Response to Frontend – Vue menerima respons (misal data barang terbaru setelah disimpan) dan memperbarui tampilan (misal menutup modal input barang atau menampilkan notifikasi sukses).

Seluruh komunikasi ini terjadi secara lokal (offline), sangat cepat dan tanpa ketergantungan internet. Ketika koneksi internet tersedia, backend sidecar secara terpisah akan mencoba sinkronisasi ke server pusat (lihat bagian sinkronisasi di bawah) – proses sync ini tidak mengganggu alur request utama UI karena berjalan asynchronous di background.
Sinkronisasi Data Offline–Online (SQLite ↔ PostgreSQL)

Sebagai aplikasi offline-first, semua operasi data dilakukan terlebih dahulu di database lokal (SQLite) agar aplikasi selalu responsif meskipun offline
sqliteforum.com
. Seluruh CRUD (Create, Read, Update, Delete) akan langsung diterapkan ke SQLite lokal. Kemudian, saat koneksi internet terdeteksi, perubahan lokal akan disinkronkan ke server pusat (PostgreSQL) di cloud. Berikut mekanisme sinkronisasi dan penanganan konflik:

    Database Ganda (SQLite & PostgreSQL): Struktur skema database antara lokal (SQLite) dan pusat (PostgreSQL) sebaiknya serupa agar proses sync lebih mudah. Misal terdapat tabel Barang, Cabang, Penjualan, PenjualanItem, Stok, StockOpname, dll dengan kolom-kolom identik. Setiap record memiliki primary key unik (misal UUID) agar dapat dikenali di kedua sisi, serta kolom timestamp (mis. updated_at) dan penanda synced (boolean atau versi)
    sqliteforum.com
    sqliteforum.com
    . Ketika offline, aplikasi akan terus menambah/mengubah data di SQLite.

    Queue Perubahan (Outbox): Backend lokal akan menandai setiap perubahan baru sebagai belum tersinkronisasi (misal synced = 0 atau mencatat ID record di tabel antrian sync). Contohnya, saat menambah penjualan baru offline, record penjualan dan item-itemnya ditulis ke SQLite dengan flag synced = false. Data ini dianggap sumber kebenaran sementara di lokal.

    Deteksi Koneksi & Trigger Sync: Komponen backend perlu memonitor status konektivitas. Begitu koneksi internet tersedia, backend akan trigger sinkronisasi secara otomatis
    sqliteforum.com
    . Cara mendeteksi koneksi bisa melalui ping endpoint server secara periodik, atau memanfaatkan event online/offline di Electron. Misalnya, tiap 5 menit backend mencoba hit endpoint ringan (seperti /ping) ke server pusat; jika berhasil, dianggap online dan proses sync dimulai. Auto-retry: Jika sinkronisasi gagal (karena koneksi putus), sistem akan menunggu dan mencoba lagi secara otomatis ketika koneksi pulih, sehingga akhirnya semua perubahan terkirim
    sqliteforum.com
    .

    Mengirim Perubahan ke Server (Upload): Saat sync, backend mengambil semua record lokal yang belum tersinkron (flag synced=0). Kemudian untuk tiap record tersebut, dilakukan request ke server (misal REST API di cloud) untuk mengunggah perubahan. Contoh pseudocode
    sqliteforum.com
    :

    if networkAvailable {
        unsynced := db.Where("synced = 0").Find(&records)
        for each record in unsynced {
            resp := POST to serverAPI(record)
            if resp.success {
                record.synced = 1
                db.Save(record)  // tandai sudah tersinkron
            }
        }
    }

    Semua perubahan lokal dikirim batch agar efisien (bisa kirim banyak sekaligus)
    sqliteforum.com
    . Setelah server mengonfirmasi terima data, backend lokal menandai item tersebut synced = 1. Dengan strategi incremental sync, kita hanya mengirim data yang berubah saja, menghemat bandwidth dan waktu
    sqliteforum.com
    .

    Menarik Data dari Server (Download): Sinkronisasi dua arah diperlukan jika data juga bisa berubah di server pusat (misal admin di cloud menambah/ubah data master barang atau cabang). Maka, backend lokal akan menarik update dari server. Caranya: simpan timestamp terakhir sync (misal last_sync_at), lalu saat online, minta server mengirim data yang diubah sejak timestamp tersebut. Server (PostgreSQL) dapat menyediakan endpoint seperti /api/sync/changes?since=2025-12-01T00:00:00Z yang mengembalikan daftar record baru/terubah (misal barang baru, perubahan harga, dsb). Backend lalu menggabungkan perubahan itu ke SQLite lokal:

        Insert baru jika record belum ada di lokal.

        Update record lokal jika ada perubahan di server dengan timestamp lebih baru dari lokal.

        Delete jika ada penghapusan di server (bisa tandai dengan flag deleted_at, lalu lokal ikut menghapus atau menandai).

    Setiap entry yang diupdate/dimainkan lokal melalui proses download juga bisa diberi synced=1 karena kini sudah selaras.

    Resolusi Konflik: Konflik terjadi jika record yang sama diubah di lokal dan di server sebelum sync (contoh: Barang X di-rename offline di cabang, tapi pusat juga mengganti nama Barang X). Untuk mencegah data saling timpa tanpa kontrol, terapkan strategi resolusi:

        Strategi Last-Write-Wins (LWW) – perubahan dengan timestamp terbaru akan diambil sebagai kebenaran
        shakilbd.medium.com
        . Misal, jika perubahan offline dilakukan pukul 10:00 dan perubahan server pukul 10:05, maka perubahan server dianggap final (offline update ditimpa) atau sebaliknya sesuai kebijakan bisnis. Pendekatan LWW ini sederhana dan otomatis
        sqliteforum.com
        .

        Strategi Server-Override – prioritaskan data server (misal untuk data master, selalu anggap pusat paling benar). Setiap konflik, versi server menimpa lokal, namun mungkin log perubahan lokal disimpan terpisah untuk audit.

        Strategi Client-Override – kebalikan di atas, cocok bila offline user harus diprioritaskan (misal dalam entri transaksi di cabang).

        Merge Field-Level – untuk data kompleks, bisa lakukan merge per kolom jika memungkinkan. Namun untuk kebanyakan kasus POS (yang datanya relatif sederhana), ini jarang diperlukan.

        Manual Review – tandai konflik dan minta admin menyelesaikan secara manual (misal melalui UI khusus konflik).

    Pendekatan yang dipilih harus konsisten. Skenario POS umumnya: data transaksi (penjualan, stok) diinput oleh cabang, jarang diubah oleh pusat, sehingga konflik minimal. Konflik lebih mungkin di data master (Barang, Harga, dll). Sebuah aturan praktis: “terapkan perubahan terbaru” bisa diterima dengan catatan ada log perubahan. Kita juga dapat menambahkan kolom updated_by (misal “branch1” atau “pusat”) sehingga saat konflik, app tahu mana sumber perubahan dan bisa memberi prioritas (misal pusat override cabang untuk data tertentu). Best practice menyarankan mendefinisikan aturan jelas, apakah last-write-wins, server-preferred, atau user-resolved
    sqliteforum.com
    , sejak awal pengembangan.

    Integrasi Server Pusat: Pada sisi server (online), bisa dibangun API dengan stack yang sama (Go + Gin + GORM) menghubungkan PostgreSQL. Namun, bisa juga server menggunakan teknologi berbeda (asal API kompatibel). Intinya, server harus menerima data dari cabang (inbound sync) dan menyediakan data ke cabang (outbound sync). Untuk kemudahan, kita dapat menggunakan format JSON konsisten dan endpoint terstruktur (misal /api/sync/upload untuk menerima batch perubahan dari cabang, dan /api/sync/download?branch={id} untuk memberikan data terkini ke cabang). Gunakan transaksi di server saat memproses batch sync untuk menjaga konsistensi (semua insert/update terapkan atomik).

    Keamanan & Identitas Cabang: Setiap sinkronisasi perlu mengidentifikasi cabang asal data. Saat online, sidecar bisa mengirim token/kunci autentikasi (misal API key per cabang atau JWT) supaya server tahu data tersebut milik cabang mana. Juga, data dari server yang ditarik harus difilter per cabang (kecuali data global seperti daftar barang yang sama untuk semua cabang). Sebaiknya, di database pusat, tabel seperti Penjualan memiliki kolom branch_id untuk menandai cabang sumber. Sehingga, saat cabang A sync, ia hanya menarik penjualan miliknya saja.

Dengan mekanisme di atas, aplikasi cabang tetap beroperasi lancar offline, dan sinkron akan terjadi begitu internet tersedia
reddit.com
. Data lokal selalu up-to-date dengan usaha sinkronisasi periodik. Auto-retry memastikan tidak ada data hilang – perubahan akan queued dan dikirim ketika jaringan kembali
sqliteforum.com
. Pendekatan offline-first ini menjamin operasional toko tidak terganggu masalah internet, namun tetap terkonsolidasi ke pusat secara eventual consistency.
Kode Boilerplate Awal (Menjalankan Proyek)

Untuk membantu memulai proyek, berikut contoh kode dasar (boilerplate) yang menggabungkan Vue, Electron, dan Go. Kode ini mencakup menjalankan aplikasi secara lokal:

1. Backend (Go + Gin + GORM) – file: backend/main.go
Kode Go berikut menginisialisasi database SQLite, mendefinisikan model, dan menjalankan server Gin dengan satu endpoint contoh.

package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Model contoh: Barang (Product)
type Product struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"size:100"`
    Stock int
    Price float64
}

func main() {
    // Koneksi ke SQLite lokal
    db, err := gorm.Open(sqlite.Open("offline.db"), &gorm.Config{})
    if err != nil {
        panic("Failed to connect database")
    }
    // Migrasi schema (buat tabel jika belum ada)
    db.AutoMigrate(&Product{})

    router := gin.Default()

    // Endpoint health-check (untuk cek koneksi)
    router.GET("/api/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Endpoint GET /api/products - ambil daftar barang
    router.GET("/api/products", func(c *gin.Context) {
        var products []Product
        db.Find(&products)
        c.JSON(http.StatusOK, products)
    })

    // (Endpoint POST, PUT, DELETE untuk CRUD barang bisa ditambahkan serupa)

    // Jalankan server di localhost:8080
    router.Run("127.0.0.1:8080")
}

Penjelasan: Kode di atas membuat database offline.db (SQLite) dan tabel products sesuai model Product. Endpoint /api/products mengembalikan daftar produk dalam bentuk JSON. Server hanya mendengarkan di 127.0.0.1:8080 (loopback) demi keamanan. Dalam skenario sebenarnya, kita akan menambahkan lebih banyak endpoint (misalnya /api/sales untuk checkout, dll.), namun kode ini cukup untuk verifikasi awal.

2. Main Process (Electron) – file: electron-main/main.js
Kode JavaScript berikut menjalankan backend Go (binary sudah terbangun sebagai server), lalu membuka jendela Electron yang memuat aplikasi Vue.

const { app, BrowserWindow } = require('electron');
const { spawn } = require('child_process');
const path = require('path');

let goProcess;
function createWindow() {
  const win = new BrowserWindow({
    width: 1200, height: 800,
    webPreferences: {
      contextIsolation: true,  // isolasi konteks (keamanan)
      nodeIntegration: false,  // tidak izinkan require di renderer
      // preload: path.join(__dirname, 'preload.js') // (jika ada file preload)
    }
  });
  // Muat URL dev atau file produksi
  if (process.env.ELECTRON_DEV) {
    win.loadURL('http://localhost:5173'); // URL dev server Vite (Vue)
  } else {
    win.loadFile(path.join(__dirname, '../renderer/dist/index.html'));
  }
}

// Event saat Electron siap
app.whenReady().then(() => {
  // Jalankan binary Go backend
  const binaryPath = process.platform === 'win32' ? 'server.exe' : './server';
  goProcess = spawn(binaryPath, [], { cwd: path.join(__dirname, '../backend') });

  goProcess.stdout.on('data', data => console.log(`Go output: ${data}`));
  goProcess.stderr.on('data', data => console.error(`Go error: ${data}`));

  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

// Pastikan mematikan proses Go saat aplikasi ditutup
app.on('before-quit', () => {
  if (goProcess) goProcess.kill();
});

Penjelasan: Electron akan menjalankan child process server (dari folder backend) menggunakan Node.js child_process.spawn. Opsi stdio diset default sehingga output Go dapat ditangkap. Setelah itu, fungsi createWindow() membuat jendela UI. Saat development, kita menggunakan environment variable (mis. ELECTRON_DEV) untuk menentukan apakah memuat URL dev server atau file hasil build. Hook before-quit memastikan sidecar Go dihentikan agar tidak ada proses zombie saat aplikasi keluar.

3. Renderer (Vue 3) – file: renderer/src/App.vue (cuplikan)
Kode Vue berikut adalah contoh sangat sederhana untuk menguji koneksi ke backend. Ini akan memanggil endpoint /api/health saat komponen dimount dan menampilkan status koneksi.

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold">Status Backend:</h1>
    <p class="mt-2 text-lg">{{ status }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const status = ref('Memeriksa koneksi...');

onMounted(async () => {
  try {
    const res = await fetch('http://localhost:8080/api/health');
    if (!res.ok) throw new Error(res.statusText);
    const data = await res.json();
    status.value = 'Backend ' + data.status;  // seharusnya "Backend ok"
  } catch (err) {
    status.value = 'Backend tidak tersedia';
  }
});
</script>

Penjelasan: Komponen di atas mencoba melakukan fetch ke http://localhost:8080/api/health. Jika backend berjalan dan merespons {"status":"ok"}, maka teks akan berubah menjadi “Backend ok”. Jika gagal (misal backend belum siap), akan muncul “Backend tidak tersedia”. Ini membantu memastikan integrasi Electron ↔ Go ↔ Vue berfungsi. Pada implementasi asli, tentu Vue akan memiliki routing halaman, state management (misal Pinia/Vuex) untuk keranjang, dsb., namun contoh di atas memberi kerangka minimal.

Menjalankan Proyek: Setelah menulis kode di atas, berikut langkah untuk menjalankannya:

    Menjalankan Backend: Pastikan Go terinstal, lalu di folder backend/ jalankan go build -o server (untuk build binary) atau cukup go run main.go (untuk menjalankan langsung). Pastikan tidak ada error dan server mendengarkan di port 8080.

    Menjalankan Frontend: Pastikan Node.js terinstal. Masuk ke folder renderer/, lalu install dependensi (npm install). Jalankan server dev Vue, misal npm run dev (untuk config Vite akan default di localhost:5173).

    Menjalankan Electron: Di folder electron-main/ (atau root proyek jika disatukan), jalankan npm install (untuk Electron) lalu npm run start (tergantung script yang diset, contoh di package.json main kita pakai "start": "electron ."). Electron akan spawn backend (jika belum jalan) dan membuka window memuat URL Vue dev. Anda akan melihat tampilan Vue dan status koneksi backend.

    Verifikasi: Jika semua benar, teks status pada UI akan menunjukkan “Backend ok”. Ini berarti front-end berhasil berkomunikasi dengan API backend lokal. Anda kini memiliki kerangka dasar untuk mulai membangun fitur.

    Tip: Selama development, Anda dapat menjalankan backend, frontend, dan Electron secara terpisah. Untuk kemudahan, pertimbangkan menggunakan tool seperti Concurrently untuk menjalankan ketiga proses dalam satu perintah (frontend dev server, Go backend, dan Electron). Misal, tambahkan script di root package.json:

    "scripts": {
      "dev:go": "go run ./backend/main.go",
      "dev:vue": "cd renderer && npm run dev",
      "dev:electron": "cross-env ELECTRON_DEV=1 electron .",
      "dev": "concurrently -k -n \"GO,WEB,ELC\" -c \"blue,green,yellow\" \"npm:dev:go\" \"npm:dev:vue\" \"npm:dev:electron\""
    }

    Perintah npm run dev lalu akan mengeksekusi backend Go, server Vue, dan Electron secara bersamaan (dengan output berwarna). Ini mempercepat siklus pengembangan. 

Dengan boilerplate di atas, struktur dasar aplikasi telah siap. Selanjutnya, implementasi fitur-fitur POS dapat dibangun di atas kerangka ini.
Implementasi Fitur-Fitur Utama

Berikut penjabaran rancangan untuk fitur-fitur utama yang diminta. Tiap fitur dijelaskan dari sisi front-end (komponen UI, alur pengguna) dan back-end (endpoint API lokal, logika, serta pertimbangan sinkronisasi). Bahasa yang digunakan mengacu pada domain POS: Barang (produk), Keranjang & Checkout (transaksi penjualan), Laporan Penjualan, Cabang, dan Stock Opname.
Fitur 1: CRUD Barang (Produk)

Deskripsi: Fitur ini mengelola data barang/produk yang dijual toko. Termasuk membuat produk baru, membaca/menampilkan daftar produk, mengedit detail produk, dan menghapus produk. Data barang mencakup misalnya kode/ID, nama barang, kategori, harga, dan stok. Karena aplikasi mendukung multi-cabang, setiap barang dapat memiliki stok terpisah per cabang.

Frontend: Akan disediakan halaman Master Barang dalam aplikasi Vue. Halaman ini menampilkan tabel daftar barang. Di atasnya ada tombol “Tambah Barang” untuk membuka modal/form input barang baru. Tiap baris barang juga memiliki aksi “Edit” dan “Hapus”. Implementasi dengan ShadCN Vue dan Tailwind dapat menghasilkan tampilan modern – misal menggunakan komponen tabel, modal, dan form elemen bergaya konsisten.

    List Barang: Menggunakan komponen tabel responsif. Data diambil dari API lokal /api/barang (GET) dan ditampilkan dalam tabel. Bisa ditambahkan fitur pencarian atau filter.

    Tambah/Edit Barang: Modal berisi form (input nama, harga, dll). Saat submit, Vue akan memanggil method yang melakukan request ke API:

        POST /api/barang untuk tambah baru.

        PUT /api/barang/:id untuk update barang tertentu.
        Setelah sukses, modal ditutup dan daftar barang di-refresh. UI sebaiknya menampilkan notifikasi (toast) sukses menggunakan komponen ShadCN (misal komponen Alert).

    Hapus Barang: Tersedia button/trash icon di tabel. Ketika diklik, muncul konfirmasi (dialog). Jika user konfirmasi, panggil API DELETE /api/barang/{id}. Setelah sukses, hapus baris dari list atau refresh data.

Seluruh operasi di atas tidak membutuhkan internet karena API terpenuhi oleh backend lokal. Data disimpan ke SQLite segera. Misalnya, user menambah barang “Pulpen A”, backend akan insert ke SQLite offline dan mengembalikan data barang baru (dengan ID). Vue lalu menambahkannya ke tabel tampilan.

Backend: Menyediakan endpoints RESTful untuk barang, misalnya:

    GET /api/barang – Mengambil list barang (termasuk stok saat ini, bisa agregat atau khusus cabang tergantung desain tabel stok).

    GET /api/barang/{id} – (Opsional) Mengambil detail barang tertentu.

    POST /api/barang – Menambah barang baru. Backend akan membuat record baru di SQLite (GORM: db.Create(&Barang)).

    PUT /api/barang/{id} – Mengubah data barang. Backend update kolom-kolom diperbolehkan (nama, harga, dll).

    DELETE /api/barang/{id} – Menghapus barang (bisa soft delete dengan flag, atau hard delete).

Karena stok bisa berbeda per cabang, desain database dapat:

    Tabel barang (id, nama, harga, dll) tanpa kolom stok.

    Tabel stok_cabang terpisah (barang_id, cabang_id, qty). Setiap kali barang dibuat, kita inisialisasi stok 0 untuk setiap cabang (atau khusus cabang terkait). Atau, jika barang berlaku di semua cabang, bisa ditambahkan entri stok per cabang saat sinkronisasi.

    Alternatif: menyimpan stok dalam tabel barang tapi di-scope per cabang (misal ada kolom branch_id yang jika null berarti global barang, jika terisi berarti entri barang khusus cabang – namun ini kurang rapi).

Dalam konteks cabang offline: Cabang A menambah barang baru -> data masuk SQLite lokal dengan branch_id = A (atau stok entri terpisah). Saat online, data barang disinkron ke server pusat. Server bisa memutuskan apakah barang ini global (terlihat di semua cabang) atau hanya cabang A. Misal, untuk POS retail biasanya daftar barang disamakan di semua cabang, maka mungkin penambahan barang hanya boleh oleh admin pusat. Namun jika cabang boleh menambah barang sendiri (mungkin di waralaba kecil), perlu disinkron ke pusat dan diberi flag cabang.

Untuk simplifikasi, kita anggap data barang global (semua cabang punya list barang yang sama). Jadi skenario: barang ditambah di cabang offline -> sync -> server simpan barang baru -> server distribusikan barang itu ke cabang lain saat sync (jika relevan). Pastikan ID unik (gunakan UUID agar tidak tabrakan).

Sinkronisasi Barang:

    Upload: Perubahan CRUD barang di cabang (tambah/edit/hapus) ditandai unsynced. Saat online, kirim ke server.

    Download: Jika pusat menambah/mengubah barang (misal update harga) saat cabang offline, cabang akan mendapat update itu saat sync. Contoh konflik: pusat ubah harga barang X, cabang offline juga ubah harga barang X. Konflik diselesaikan sesuai aturan (misal pusat override).

    Barang yang dihapus di pusat harus membuat cabang menghapus juga (bisa pakai soft delete dengan timestamp deleted_at agar terpropagasi).

Saat melakukan DELETE, sebaiknya hapus logis (soft delete) agar data histori penjualan tidak hilang referensi. Soft delete berarti kolom deleted_at atau flag is_deleted. Cabang menerima informasi ini lalu dapat menyembunyikan barang tsb dari list.

UI/UX: Pastikan ketika offline tanpa server, fitur CRUD barang tetap berjalan normal. User dapat menambah/edit berkali-kali. Begitu konek, sync mungkin memakan waktu sejenak di latar belakang. Kita bisa menambahkan indikator sinkronisasi (misal ikon status di footer: “Sinkronisasi...” atau warna koneksi). Namun idealnya, pengguna tidak perlu terlalu khawatir karena data akan otomatis sync tanpa intervensi
thisisglance.com
.
Fitur 2 & 3: Keranjang & Checkout (Transaksi Penjualan)

Deskripsi: Fitur ini mencakup proses penjualan barang. Keranjang merepresentasikan daftar barang yang akan dibeli customer (beserta kuantitas tiap item), dan Checkout memfinalisasi transaksi (mengurangi stok, mencatat penjualan, dan mengosongkan keranjang). Output dari checkout biasanya termasuk struk/nota penjualan.

Frontend (Keranjang): Dalam aplikasi POS, UI keranjang biasanya selalu terlihat di layar penjualan. Kita siapkan halaman Kasir/Penjualan di Vue. Bagian kiri daftar barang (bisa cari & tambah ke keranjang), bagian kanan adalah keranjang current. Implementasi:

    Daftar barang (kiri): dapat berupa tabel atau grid item. User bisa memilih barang (misal klik baris atau tombol “Add to cart”). Bisa juga ada field pencarian cepat.

    Saat barang ditambahkan, keranjang (kanan) menampilkan item tersebut dengan kolom nama, qty, harga satuan, subtotal. User bisa mengubah quantity di keranjang (misal input number atau +/– button).

    Total harga dihitung otomatis di UI (sum of subtotal).

    Tombol “Checkout” akan mengonfirmasi transaksi.

Gunakan komponen UI ShadCN Vue untuk elemen input dan list. Misal, keranjang bisa dibuat scrollable list, tiap item keranjang barisnya pakai komponen Card/List. Pastikan styling Tailwind konsisten.

Frontend (Checkout): Ketika user klik “Checkout”, berikan dialog konfirmasi (muncul modal “Yakin proses transaksi?” total X). Jika OK, proses sebagai berikut:

    Kirim data transaksi ke backend. Data transaksi mencakup: daftar item (ID barang dan qty, harga per item), total, waktu, ID kasir (jika multi user login), dan ID cabang.

    Endpoint yang dipanggil: POST /api/penjualan (misal) dengan payload JSON berisi detail transaksi.

    Sambil menunggu respon, UI bisa menampilkan loading indicator. Setelah berhasil:

        Tampilkan notifikasi “Transaksi Berhasil” (mungkin juga nomor nota).

        Reset state keranjang (kosongkan list).

        (Opsional) Tampilkan/print struk. Jika ingin mencetak struk, kita bisa generate struk di backend (misal PDF) atau langsung print dari frontend. Electron memungkinkan akses printer lokal, tapi bisa komplek; dapat disederhanakan dengan membuka window struk untuk cetak via browser dialog.

Backend (Transaksi): Endpoint POST /api/penjualan akan melakukan:

    Mencatat Penjualan: Buat record baru di tabel penjualan (fields: id, tanggal/waktu, total, branch_id, user/kasir, dll). Gunakan transaksi database: GORM db.Transaction(...).

    Mencatat Item Penjualan: Loop daftar items dalam request, untuk masing-masing:

        Kurangi stok barang terkait di tabel stok_cabang (stok = stok - qty terjual).

        Buat record di tabel penjualan_item (fields: id, penjualan_id (FK), barang_id, qty, harga_saat_ini).

    Commit transaksi: Jika seluruh insert/update berhasil, commit. Jika ada gagal (misal stok tidak cukup, walau seharusnya dicek di UI), lakukan rollback dan kembalikan error.

Backend mengembalikan respons sukses (bisa berisi data penjualan baru atau sekadar message). Karena bekerja offline, semua ini ditulis ke SQLite. Penjualan dan item diberi flag synced = 0 untuk sync nanti.

Sinkronisasi Penjualan: Data penjualan biasanya critical untuk terpusat. Mekanisme sync:

    Upload: Record penjualan dan item dikirim ke server pusat saat online. Karena menggunakan ID (UUID) sendiri di cabang, server perlu memasukkan dengan ID tersebut atau map ke ID baru. Lebih mudah gunakan UUID agar id tidak bentrok. Data stok pengurangan juga perlu direfleksikan di pusat: server bisa mengurangi stok global atau stok cabang. Kemungkinan di pusat, stok dicatat per cabang juga, jadi server cukup mengurangi stok cabang A sesuai penjualan.

    Download: Penjualan biasanya tidak diubah di pusat setelah tercipta, jadi konflik jarang. Namun, kalau ada mekanisme edit transaksi di pusat (misal koreksi), perlu ditarik. Dalam desain sederhana, anggap tidak ada edit penjualan di pusat, jadi cabang yang buat penjualan, pusat hanya menerima.

    Konfirmasi: Bisa tambahkan mekanisme agar setelah sync, server mengirim balik konfirmasi (misal cap ‘synced’ atau nomor nota pusat). Namun untuk offline-first, tidak wajib real-time.

Stok: Pengurangan stok sudah dilakukan offline saat checkout. Namun, ada skenario race-condition jika ada beberapa POS di cabang sama? Jika aplikasi digunakan multi-terminal offline tanpa koneksi satu sama lain, stok bisa minus tanpa sadar. Asumsi kita: satu cabang satu database (satu sidecar), jadi tidak ada duplikasi di satu cabang. Antar cabang, stok terpisah. Pusat bisa menjumlah total stok semua cabang jika perlu.

Contoh: Kasir di Cabang A men-scan barang, 5 item, checkout. SQLite Cabang A: tabel penjualan bertambah 1 record, stok_cabang barang tersebut berkurang 5. Flag synced=0. Dua jam kemudian internet tersambung, data penjualan dikirim ke server. Server insert ke Postgres tabel penjualan (branch A), kurangi stok branch A di Postgres. Jika sukses, server respon OK; sidecar tandai penjualan synced=1. Jika gagal (misal server down), sidecar akan retry belakangan.

UI/UX: Pastikan saat checkout offline, pengguna tidak merasakan perbedaan (transaksi tetap tercatat). Jika internet ada, bisa kirim parallel ke server tanpa tunggu (fire-and-forget) untuk percepat sync. Namun, hal ini tidak boleh menghambat UI. UI harus anggap transaksi selesai begitu SQLite update, tidak perlu tunggu server. Keunggulan offline-first: transaksi selalu lancar – user bisa lanjut melayani customer berikutnya tanpa menunggu koneksi
thisisglance.com
thisisglance.com
.

Nota/Struk: Fitur cetak struk bisa ditangani dengan integrasi printer. Bisa generate HTML struk dan menggunakan window.print() dari renderer. Atau generate PDF di backend (misal dengan library PDF) dan buka file tersebut. Ini implementasi tambahan; inti disini adalah data penjualan tersedia untuk dicetak.
Fitur 4: Export Laporan Penjualan (Excel, dengan Header Toko)

Deskripsi: Fitur ini menghasilkan laporan penjualan dalam format Excel (.xlsx). Laporan berisi daftar transaksi penjualan dalam periode tertentu, metrik total, dsb., dan harus dapat mencantumkan nama & logo toko yang bisa diubah (editable).

Frontend: Sediakan halaman Laporan Penjualan. User dapat memilih filter (misal rentang tanggal “dari - sampai”, atau pilih bulan tertentu). Klik tombol “Export Excel” akan memicu permintaan pembuatan laporan. UI bisa menampilkan loader selama pembuatan file. Setelah file siap, user diberi opsi untuk membuka file atau memilih lokasi penyimpanan.

Karena aplikasi desktop, kita bisa langsung menyimpan file ke disk. Electron dapat memanfaatkan modul dialog dialog.showSaveDialog untuk meminta user menentukan lokasi simpan file. Skenario:

    User klik Export Excel.

    Vue memanggil backend API misal GET /api/laporan/penjualan?start=YYYY-MM-DD&end=YYYY-MM-DD.

    Backend membuat file Excel dan mengembalikan file (biner) atau path.

    Jika API mengembalikan file biner, front-end bisa memicu download. Namun dengan Electron, lebih baik backend simpan file di folder sementara lalu mengembalikan path, dan kemudian Electron (main process) menggunakan dialog save to move file tersebut ke lokasi pilihan user.

Alternatif implementasi: langsung generate di front-end? Tapi pembuatan Excel lebih cocok di backend (dengan akses DB). Jadi kita fokus backend.

Backend: Endpoint contoh GET /api/laporan/penjualan yang menerima parameter tanggal awal/akhir. Langkah:

    Query SQLite untuk data penjualan dalam range tersebut, termasuk join ke item dan mungkin agregat per hari atau per produk, tergantung kebutuhan laporan.

    Gunakan library Go seperti Excelize
    github.com
    untuk membuat file Excel. Excelize memungkinkan pembuatan workbook, penulisan cell, format, hingga menyisipkan gambar (logo).

    Layout laporan: Pada sheet, di bagian header atas, tulis nama toko dan letakkan logo. Nama toko bisa diambil dari konfigurasi (misal tabel toko atau file config yang bisa diedit pengguna). Logo toko mungkin disimpan sebagai file (PNG/JPG) di folder aset aplikasi; path/logo bisa diambil dari config juga. Excelize menyediakan fungsi AddPicture untuk menyisipkan gambar ke sel tertentu
    dev.to
    .

    Tulis judul laporan (misal “Laporan Penjualan Tanggal X s.d Y”), kemudian tabel berisi kolom: No, Tanggal, No. Nota, Total Penjualan, dsb. Jika butuh detail per item, bisa multi-sheet atau multiple sections.

    Setelah data ditulis, simpan workbook ke file (misal LaporanPenjualan_2025-12-01_2025-12-31.xlsx di folder sementara).

    Kembalikan file tersebut ke user. Dengan Gin, untuk mengirim file bisa pakai c.File(filepath) sehingga Electron akan mendownload otomatis. Namun kadang untuk Electron, lebih baik kembalikan path. Tergantung pendekatan:

        Option A (download via Axios): Kirim file melalui response. Perlu set header application/vnd.openxmlformats-officedocument.spreadsheetml.sheet dan Content-Disposition: attachment; filename=.... Renderer (axios) harus set responseType: 'blob' dan lalu membuat blob URL untuk trigger download. Ini mirip web biasa.

        Option B (save then prompt): Backend menyimpan file di disk (misal di folder exports/), lalu merespons JSON berisi path: {"path": "C:/app/exports/report.xlsx"}. Renderer kemudian memanggil ipcRenderer.invoke('save-dialog', filePath) yang memunculkan dialog save. Main process kopi file ke lokasi user pilih.

    Opsi B lebih “native” dan memberi user kendali lokasi. Kita dapat memanfaatkan modul fs untuk copy file.

Dalam hal ini, langkah mudah: kembalikan langsung file untuk simplicity, anggap user mengambil dari folder default (misal Download). Namun karena user ingin nama/logo editable, kemungkinan mereka akan buka file dan bisa mengedit di Excel manual (nama toko atau logo). Itu oke, tapi juga bisa diartikan user dapat mengatur nama toko/logo di aplikasi sebelum export.

Konfigurasi Nama/Logo Toko: Tambahkan di aplikasi suatu tempat untuk mengatur profil toko (nama, alamat, logo). Nilai ini disimpan lokal (dan mungkin di server juga). Saat export, backend membaca config tsb. Logo bisa disimpan di disk; simpan path di config DB. Excelize: gunakan AddPicture dengan path logo file
dev.to
. Misal meletakkan di cell A1 with some offset/size.

Sinkronisasi: Laporan penjualan sebenarnya dihasilkan dari data penjualan lokal. Karena model offline-first, laporan ini bisa dibuat kapan saja bahkan tanpa internet, tapi datanya hanya akan selengkap data yang dimiliki lokal sampai saat itu. Jika ingin laporan full sampai hari ini, tapi cabang baru online sekarang, data penjualan hari ini ada lokal, data hari sebelumnya juga (sinkronisasi sudah menarik semua histori penjualan cabang? Biasanya penjualan cabang lain tidak perlu ditarik ke cabang ini, kecuali laporan tingkat pusat). Diasumsikan laporan ini untuk cabang lokal saja (menampilkan transaksi cabang tersebut).

Jika ingin laporan gabungan seluruh cabang, mungkin fitur ini hanya digunakan di mode online atau di aplikasi pusat. Dalam konteks offline-first di cabang, laporan berarti laporan cabang itu. Jadi tidak perlu data dari cabang lain.

Contoh: User (manager cabang) memilih export laporan bulan Desember 2025. Aplikasi memanggil GET /api/laporan/penjualan?start=2025-12-01&end=2025-12-31. Backend query semua penjualan di SQLite antara tanggal tersebut. Dibuat Excel dengan nama toko “Toko ABC Cabang X” di header dan logo perusahaan. File disimpan, lalu dialog meminta user simpan sebagai “LaporanPenjualan_Des2025_CabangX.xlsx”. User buka file di Excel, bisa mengganti logo atau nama toko jika perlu (misal logo baru) – namun jika ingin mengubah secara permanen, lebih baik update di aplikasi config dan export ulang.

Implementasi Teknis: Gunakan Excelize untuk kemudahan. Contoh pembuatan file:

f := excelize.NewFile()
sheet := f.GetActiveSheet()
f.SetCellValue(sheet, "A1", "Laporan Penjualan")
f.SetCellValue(sheet, "A2", fmt.Sprintf("Periode: %s s.d %s", start, end))
// Merge cells for title, apply style bold, etc.
// Add table header
headers := []string{"No", "Tanggal", "No Nota", "Total", "Kasir"}
for i, h := range headers {
    cell := fmt.Sprintf("%c4", 'A'+i)  // header starting at row 4
    f.SetCellValue(sheet, cell, h)
    // (apply header styling)
}
// Fill data rows
row := 5
for idx, sale := range sales {
    f.SetCellValue(sheet, fmt.Sprintf("A%d", row), idx+1)
    f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sale.Date.Format("02-01-2006"))
    f.SetCellValue(sheet, fmt.Sprintf("C%d", row), sale.ReceiptNo)
    f.SetCellValue(sheet, fmt.Sprintf("D%d", row), sale.Total)
    f.SetCellValue(sheet, fmt.Sprintf("E%d", row), sale.CashierName)
    row++
}
// Add logo image
f.AddPicture(sheet, "F1", "/path/to/logo.png", `{"x_scale": 0.5, "y_scale": 0.5}`)
// Save to file
filePath := fmt.Sprintf("LaporanPenjualan_%s_%s.xlsx", start, end)
if err := f.SaveAs(filePath); err != nil {
    // handle error
}

Di atas, kita menulis judul di A1, header tabel di baris 4, data mulai baris 5. Penomoran kolom kita hitung manual dengan char. Excelize juga mendukung style (font bold, border) yang bisa digunakan untuk mempercantik. Logo ditaruh di cell F1 dengan skala 50%. Terakhir simpan ke file. (Untuk integrasi lebih rapi, bisa buat template Excel lalu diisi, tapi langsung generate juga oke).

Catatan: Periksa kompatibilitas Excel hasilnya (Excelize umumnya kompatibel Excel 2007+). Ukuran logo mungkin perlu diatur. Dan jika nama toko panjang, mungkin merge cell untuk nama toko.

UX: Setelah export, user kemungkinan buka filenya. Tidak perlu menjelaskan banyak, cukup pastikan file terbuka dengan benar. Agar proses ini cepat, bisa berikan notifikasi “File Laporan berhasil dibuat”. Jika file langsung diunduh ke lokasi default, beri tahu lokasinya.
Fitur 5: CRUD Cabang

Deskripsi: Fitur ini mengelola data Cabang toko. Termasuk penambahan cabang baru, edit informasi cabang, dan (mungkin) penghapusan cabang. Informasi cabang minimal: ID/Code, nama cabang, alamat, kontak, dll. Fitur ini umumnya diakses oleh admin pusat, bukan tiap cabang. Namun, karena diminta dalam konteks aplikasi desktop, kemungkinan aplikasi dapat dijalankan dalam mode “pusat” juga, atau cabang itu sendiri dapat mengedit data cabangnya (misal memperbarui alamat). Interpretasi lain: jika toko tunggal dengan beberapa cabang, admin di salah satu cabang bisa menambah data cabang lain (kurang umum).

Kemungkinan besar, CRUD Cabang terutama berfungsi saat online, karena menambah cabang baru harus disinkronkan ke semua node (semua cabang harus tahu cabang baru?). Mungkin implementasi ini lebih condong ke pusat. Namun, kita tetap jelaskan arsitektur untuk completeness.

Frontend: Halaman Data Cabang menampilkan list cabang (Nama, Kode, Alamat, dsb). Tersedia tombol “Tambah Cabang” (modal form), aksi edit pada setiap baris, dan hapus. Mirip CRUD Barang dari sisi UI (tabel + modal form).

Jika aplikasi dijalankan di mode cabang, mungkin fitur ini dibatasi (hanya tampil data cabangnya saja, tidak boleh tambah baru). Jika di mode pusat (misal ada config mode), fitur full access.

Backend: Endpoint CRUD cabang, misalnya:

    GET /api/cabang – list semua cabang

    POST /api/cabang – tambah baru

    PUT /api/cabang/{id} – edit info

    DELETE /api/cabang/{id} – hapus cabang (hati-hati, hapus cabang berarti menonaktifkan satu unit toko, mungkin soft delete)

Data disimpan di tabel cabang di SQLite dan PostgreSQL. Struktur sederhana: id (UUID atau kode unik), nama, alamat, dsb. Jika offline cabang menambah cabang baru – ini scenario aneh, tapi asumsikan mungkin admin pusat offline menambah cabang lalu konek internet. Atau cabang bisa berdiri sendiri dulu lalu di-merge ke pusat.

Sinkronisasi:

    Upload: Cabang baru yang ditambah offline akan sync ke server sehingga semua cabang lain (atau setidaknya pusat) tahu. Misal Cabang A offline menambah "Cabang Z" (mungkin cabang sub-branch?), ketika online server tambahkan "Cabang Z" di pusat. Pusat mungkin harus menyetujui data ini, tergantung kebijakan.

    Download: Jika pusat menambah cabang baru (umumnya pusat yang lakukan), cabang-cabang eksisting harus mendapatkan data cabang baru itu saat sync, untuk update referensi (misal untuk pilihan transaksi antar cabang, dsb). Jadi sidecar cabang ketika sync akan memasukkan record cabang baru ke SQLite-nya.

    Konflik kecil kemungkinan (dua tempat menambah cabang dengan nama sama?) – bisa diabaikan atau ditangani manual (ID distinct).

    Hapus cabang: jika pusat hapus, cabang yang dihapus tidak akan sync lagi (mungkin device cabang tsb tidak digunakan lagi). Cabang lain bisa mendapati data cabang itu ditandai nonaktif.

Penggunaan: Data cabang dipakai di fitur lain, misal:

    Saat input barang, mungkin pilih cabang jika barang khusus cabang (jarang, biasanya barang global).

    Laporan pusat per cabang (di pusat app).

    Stock opname per cabang.
    Karena aplikasi kita per cabang, mungkin data cabang hanya untuk referensi ID cabangnya sendiri. Kecuali aplikasi pusat dijalankan dengan modul sama.

Agar sesuai soal, kita jelaskan CRUD-nya saja. Intinya, implementasi mirip CRUD Barang:
form input (nama, alamat, dsb), list table, API calls, offline-first (tulis SQLite, sync belakangan).

Contoh: Admin membuka Data Cabang di mode pusat: ada 5 cabang terdaftar. Admin klik “Tambah Cabang”, isi nama “Cabang Baru XYZ”. Submit: panggil POST /api/cabang, backend Go (mungkin berjalan di mode pusat dengan Postgres?) atau kalau di cabang offline akan ke SQLite. Karena ini rancu, mungkin lebih masuk akal fitur ini dijalankan ketika app terkoneksi ke pusat.

Namun, sesuai permintaan, kita tetap bisa menaruh modul ini di app offline-first: cabang baru bisa ditambahkan offline (walau tidak terlalu realistis di dunia nyata operasi). Jadi kita dukung offline: data masuk lokal, sync ke pusat nanti.
Fitur 6: Stock Opname & Export Laporan (.xlsx)

Deskripsi: Stock Opname adalah proses pencatatan stok fisik barang di toko lalu mencocokkannya dengan stok di sistem, biasanya dilakukan secara periodik (misal bulanan). Fitur ini meliputi: mengambil data stok sistem saat ini, memasukkan hasil hitung fisik per barang, menghitung selisih, memperbarui stok sistem sesuai hasil hitung, dan kemudian mengekspor laporan hasil stock opname dalam Excel.

Frontend: Sediakan halaman Stock Opname. Alur penggunaan:

    User memilih aksi "Mulai Stock Opname". Aplikasi akan mengambil daftar barang beserta stok menurut sistem (SQLite) saat ini. Tampilkan dalam tabel: Kolom Barang, Stok Sistem, input Stok Fisik (editable), dan Selisih (auto kalkulasi = fisik - sistem).

    User melakukan pengisian stok fisik untuk setiap item (bisa real-time hitung selisih).

    Setelah selesai, user klik “Simpan/Finalisasi Opname”. Muncul konfirmasi, karena tindakan ini akan mengubah stok sistem.

    Setelah konfirmasi, data perbedaan stok akan dikirim ke backend untuk diproses.

UI dapat ditingkatkan: misal bisa import dari scanner atau CSV, tapi secara default manual input. Pastikan UI mendukung banyak item (mungkin ratusan) – bisa buat scrollable area. ShadCN Vue dapat dipakai untuk input number inline dalam tabel. Mungkin tambahkan fitur “Set semua stok fisik = stok sistem” sebagai default to ease input (user hanya ubah yang beda).

Backend: Kapan user finalisasi, kirim ke endpoint POST /api/stock-opname. Payload berisi daftar hasil count: {barang_id, qty_sistem, qty_fisik}. Langkah backend:

    Buat record baru di tabel stock_opname (fields: id, tanggal, user, branch_id).

    Loop tiap item dari payload, hitung selisih = qty_fisik - qty_sistem.

    Update tabel stok: set stok = qty_fisik untuk barang tersebut (di cabang terkait). Ini menyesuaikan sistem dengan real count.

    Simpan record detail ke tabel stock_opname_item dengan kolom: opname_id, barang_id, qty_sistem, qty_fisik, selisih. Tujuannya untuk pelaporan historis.

    Selesai, return sukses.

Catatan: Proses ini mungkin ingin atomic, jadi bungkus dalam transaksi. Juga, penyesuaian stok akibat opname harus diperhitungkan jika ada transaksi berjalan atau sudah terjadi setelah data diambil. Dalam offline mode single-user, kita anggap toko ditutup sementara saat opname, jadi tidak ada penjualan yang mengubah stok di tengah proses.

Export Laporan Opname: Mirip dengan laporan penjualan, setelah finalisasi, user kemungkinan akan ingin export hasil stock opname sebagai Excel. Bisa ada tombol “Export Laporan Opname” di halaman tersebut (atau langsung tanya “Opname berhasil, unduh laporan?”).

Laporan ini berisi: Tanggal opname, siapa petugas, lalu tabel item dengan kolom Barang, Stok Sistem sebelum, Stok Fisik, Selisih, mungkin kolom nilai (kalau menghitung kerugian selisih). Format Excel juga bisa memuat nama & logo toko seperti laporan penjualan.

Implementasi export mirip fitur 4:

    Endpoint GET /api/laporan/stock-opname/{opname_id} yang menghasilkan Excel. Atau mungkin tanpa param, kirim opname terbaru.

    Backend query stock_opname_item berdasarkan ID atau tanggal, kemudian pakai Excelize membuat sheet serupa: header info (nama toko, judul "Laporan Stock Opname [Tanggal]"), tabel per barang dengan kolom-kolom di atas.

    Sertakan total ringkasan: misal hitung berapa item yang selisih plus, minus, total selisih.

    Simpan atau kirim file ke front-end. Proses sama seperti laporan penjualan.

Sinkronisasi:

    Stock opname memperbarui stok barang di SQLite. Ini perubahan data barang (stok) yang perlu sync ke pusat:

        Ketika online, sidecar kirim update stok ke server (server akan update stok di Postgres). Karena mungkin banyak barang, bisa kirim ringkas: misal endpoint /api/sync/stock dengan payload {branch: X, [ {barang_id, new_stock}, ...] }. Atau rely on normal per-item update sync.

        Juga kirim data stock_opname dan detail ke server untuk arsip pusat. Pusat bisa simpan record ini juga, agar pusat punya catatan kapan cabang lakukan opname dan hasilnya (ini penting buat audit).

    Download: Jika pusat atau cabang lain mengubah stok barang yang sama di waktu berdekatan (tidak umum, karena stok seharusnya diubah hanya via transaksi atau opname cabang sendiri), conflict kecil. Asal cabang hanya mengurus stoknya, tidak perlu conflict dengan cabang lain.

    Jika offline saat opname (likely yes, opname bisa dilakukan offline), tak masalah. Data akan sync nanti.

UX: Stock opname biasanya dilakukan secara tak rutin, tapi ketika dilakukan, aplikasi harus mendukung offline entry. Pengguna mungkin berjalan di toko menghitung, kembali input. Pastikan bila perlu aplikasi bisa di-bawa (laptop) offline di gudang. Aplikasi offline-first ideal untuk ini.

Setelah finalisasi, user mendapatkan laporan yang bisa di-print atau dikirim. Pastikan perubahan stok langsung terlihat di modul lain (misal di layar barang stok sudah update sesuai opname). Kemungkinan perlu refresh data barang setelah opname.

Kesimpulan: Arsitektur dan implementasi di atas memberikan pondasi untuk aplikasi POS desktop modern yang cepat, ringan, dan mendukung offline-online. Dengan Electron dan Vue 3 di sisi UI, kita mendapatkan tampilan interaktif dan kemudahan pengembangan front-end. Sidecar backend Go menawarkan kinerja tinggi dan akses langsung ke sistem file/db lokal, serta kemudahan sinkronisasi ke server pusat (PostgreSQL) melalui REST. Pola offline-first memastikan operasional toko tidak terhenti meskipun internet down, dan data akan konsisten tersinkron ke pusat begitu koneksi pulih. Setiap fitur POS utama telah didesain mengikuti prinsip ini:

    CRUD master data (Barang, Cabang) dilakukan lokal dengan sync ke pusat.

    Transaksi penjualan berjalan lancar offline, dan tersinkron untuk laporan pusat.

    Laporan-laporan dapat diekspor langsung dari data lokal.

    Stock opname dapat dilakukan offline dan memastikan data stok akurat.

Dengan dokumentasi dan kerangka proyek ini, pengembang dapat langsung memulai membangun aplikasi POS desktop tersebut dan melakukan penyesuaian sesuai kebutuhan spesifik bisnis. Semoga bermanfaat dan selamat membangun aplikasi POS Anda!