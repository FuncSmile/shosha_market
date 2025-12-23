/**
 * useToast - Contoh Penggunaan
 * 
 * Hook untuk menampilkan toast notifications dengan vue-sonner
 * Sudah terintegrasi dengan Toaster component di App.vue
 */

import { useToast } from './useToast'

export function exampleUsage() {
  const toast = useToast()

  // ✅ Success toast (hijau)
  toast.success('Data berhasil disimpan!')
  toast.success('Produk ditambahkan ke keranjang', 'Berhasil')

  // ❌ Error toast (merah)
  toast.error('Gagal menyimpan data')
  toast.error('Koneksi ke server terputus', 'Error Jaringan')

  // ⚠️ Warning toast (kuning/amber)
  toast.warning('Stok produk tinggal 3')
  toast.warning('Data akan dihapus permanen', 'Perhatian')

  // ℹ️ Info toast (biru)
  toast.info('Sinkronisasi akan dimulai dalam 5 detik')
  toast.info('10 produk baru ditambahkan', 'Update')

  // 📝 Default toast (putih/abu)
  toast.message('Operasi sedang diproses...')
  toast.message('Tunggu sebentar', 'Memproses')

  // 🔄 Promise toast (loading → success/error)
  const saveData = async () => {
    return new Promise((resolve) => setTimeout(resolve, 2000))
  }

  toast.promise(saveData(), {
    loading: 'Menyimpan data...',
    success: 'Data berhasil disimpan!',
    error: 'Gagal menyimpan data',
  })

  // Promise dengan custom message function
  const fetchData = async () => {
    const response = await fetch('/api/products')
    return response.json()
  }

  toast.promise(fetchData(), {
    loading: 'Memuat produk...',
    success: (data) => `${data.length} produk berhasil dimuat`,
    error: (err: unknown) => {
      const errorMsg = err instanceof Error ? err.message : String(err)
      return `Error: ${errorMsg}`
    },
  })

  // ❌ Dismiss toast
  toast.dismiss() // Dismiss all toasts
  
  // Advanced: Access vue-sonner directly
  const toastId = toast.toast('Custom toast', {
    description: 'Dengan action button',
    action: {
      label: 'Undo',
      onClick: () => console.log('Undo clicked'),
    },
  })
  
  // Dismiss specific toast
  setTimeout(() => toast.dismiss(toastId), 3000)
}

/**
 * Styling
 * 
 * Toast sudah memiliki style default berdasarkan variant:
 * - success: bg-emerald-50 border-emerald-200 text-emerald-900
 * - error: bg-red-50 border-red-200 text-red-900
 * - warning: bg-amber-50 border-amber-200 text-amber-900
 * - info: bg-blue-50 border-blue-200 text-blue-900
 * - default: bg-white border-slate-200 text-slate-900
 * 
 * Untuk customize lebih lanjut, gunakan className option di toast() directly:
 * 
 * toast.toast('Custom', {
 *   className: 'bg-purple-50 border-purple-200 text-purple-900',
 * })
 */

/**
 * Durasi Default
 * 
 * - success: 3 detik
 * - error: 4 detik
 * - warning: 3 detik
 * - info: 3 detik
 * - default: 3 detik
 * 
 * Untuk custom duration, gunakan showToast() atau toast() directly:
 * 
 * toast.showToast('Pesan', 'success', 'Judul', 5000) // 5 detik
 */
