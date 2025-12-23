import { toast } from 'vue-sonner'

export type ToastVariant = 'default' | 'success' | 'info' | 'warning' | 'error'

export interface ToastOptions {
  title?: string
  description?: string
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
}

/**
 * useToast - Vue composable for displaying toast notifications
 * Uses vue-sonner (shadcn-vue Toaster component)
 * 
 * @example
 * const toast = useToast()
 * toast.success('Data berhasil disimpan')
 * toast.error('Gagal menyimpan data', 'Error')
 * toast.promise(apiCall, {
 *   loading: 'Memproses...',
 *   success: 'Selesai!',
 *   error: 'Gagal!'
 * })
 */
export function useToast() {
  const showToast = (
    message: string,
    variant: ToastVariant = 'default',
    title?: string,
    duration: number = 3000
  ) => {
    const options: ToastOptions = {
      description: message,
      duration,
    }

    if (title) {
      options.title = title
    }

    // Map to vue-sonner API with proper styling
    switch (variant) {
      case 'success':
        toast.success(title || 'Success', {
          description: message,
          duration,
          class: 'border-emerald-200 bg-emerald-50 text-emerald-900',
        })
        break

      case 'error':
        toast.error(title || 'Error', {
          description: message,
          duration: duration || 4000,
          class: 'border-red-200 bg-red-50 text-red-900',
        })
        break

      case 'warning':
        toast.warning(title || 'Warning', {
          description: message,
          duration,
          class: 'border-amber-200 bg-amber-50 text-amber-900',
        })
        break

      case 'info':
        toast.info(title || 'Info', {
          description: message,
          duration,
          class: 'border-blue-200 bg-blue-50 text-blue-900',
        })
        break

      default:
        toast(title || message, {
          description: title ? message : undefined,
          duration,
          class: 'border-slate-200 bg-white text-slate-900',
        })
    }
  }

  const success = (message: string, title?: string) =>
    showToast(message, 'success', title || '✓ Berhasil', 3000)

  const error = (message: string, title?: string) =>
    showToast(message, 'error', title || '✕ Error', 4000)

  const warning = (message: string, title?: string) =>
    showToast(message, 'warning', title || '⚠ Perhatian', 3000)

  const info = (message: string, title?: string) =>
    showToast(message, 'info', title || 'ℹ Info', 3000)

  const message = (msg: string, title?: string) =>
    showToast(msg, 'default', title, 3000)

  /**
   * Display toast for promise lifecycle
   * Shows loading toast, then success/error based on promise result
   */
  const promiseToast = <T,>(
    promiseOrFn: Promise<T> | (() => Promise<T>),
    messages: {
      loading: string
      success: string | ((data: T) => string)
      error?: string | ((error: unknown) => string)
    }
  ) => {
    const promise = typeof promiseOrFn === 'function' ? promiseOrFn() : promiseOrFn

    toast.promise(promise, {
      loading: messages.loading,
      success: (data: T) => {
        if (typeof messages.success === 'function') {
          return messages.success(data)
        }
        return messages.success
      },
      error: (error: unknown) => {
        if (messages.error && typeof messages.error === 'function') {
          return messages.error(error)
        }
        const errorMsg = error instanceof Error ? error.message : String(error)
        return errorMsg || messages.error || 'An error occurred'
      },
      duration: 3000,
    })

    return promise
  }

  /**
   * Dismiss all visible toasts
   */
  const dismiss = (toastId?: string | number) => {
    toast.dismiss(toastId)
  }

  return {
    success,
    error,
    warning,
    info,
    message,
    promise: promiseToast,
    showToast,
    dismiss,
    // Direct access to vue-sonner toast for advanced usage
    toast,
  }
}

