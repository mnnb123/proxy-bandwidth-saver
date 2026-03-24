import { create } from 'zustand'
import { toast } from 'sonner'

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning'
  message: string
}

interface ToastState {
  toasts: Toast[]
  addToast: (type: Toast['type'], message: string) => void
  removeToast: (id: string) => void
}

export const useToastStore = create<ToastState>(() => ({
  toasts: [],
  addToast: (type, message) => {
    if (type === 'success') toast.success(message)
    else if (type === 'error') toast.error(message, { duration: 5000 })
    else if (type === 'warning') toast.warning(message)
  },
  removeToast: () => {},
}))
