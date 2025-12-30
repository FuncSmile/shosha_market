# Template Import Barang (Excel)

## Format File Excel

File Excel harus memiliki **5 kolom** dengan urutan sebagai berikut:

| Nama Barang | Satuan | Stok | Harga Investor | Harga SHOSHA |
|-------------|--------|------|----------------|--------------|
| Sabun       | pcs    | 10   | 4500           | 4000         |
| Beras       | kg     | 20   | 58000          | 55000        |
| Minyak      | liter  | 15   | 18000          | 16000        |

## Aturan Validasi

1. **Nama Barang** (wajib): Nama produk
2. **Satuan** (wajib): Unit produk (pcs, kg, liter, dll)
3. **Stok** (opsional): Jumlah stok awal (default: 0)
4. **Harga Investor** (wajib*): Harga untuk cabang Investor
5. **Harga SHOSHA** (wajib*): Harga untuk cabang SHOSHA

\* **Minimal salah satu harga (Investor atau SHOSHA) harus diisi dan > 0**

## Cara Import

1. Buka halaman **Master Barang**
2. Klik tombol **Import Excel**
3. Pilih file Excel (.xlsx atau .xls)
4. Sistem akan otomatis membaca dan menambahkan ke form input
5. Klik **Simpan Semua** untuk menyimpan ke database

## Contoh File Excel

Anda bisa membuat file Excel dengan format berikut:

```
Nama Barang,Satuan,Stok,Harga Investor,Harga SHOSHA
Sabun Mandi,pcs,10,4500,4000
Beras Premium,kg,20,58000,55000
Minyak Goreng,liter,15,18000,16000
Gula Pasir,kg,25,14000,12500
```

Simpan file dengan ekstensi **.xlsx** atau **.xls**

## Catatan

- File CSV sudah tidak didukung, gunakan Excel saja
- Harga Normal sudah dihapus, hanya gunakan Harga Investor dan Harga SHOSHA
- Sistem akan otomatis set harga normal dari salah satu harga yang diisi
- Duplikat nama barang akan ditolak oleh sistem
