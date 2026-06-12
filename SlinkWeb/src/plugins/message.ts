import { createApp, type App } from 'vue'
import Toast from '../components/common/Toast.vue'
import ConfirmDialog from '../components/common/ConfirmDialog.vue'

// Toast 消息提示
interface ToastOptions {
  message: string
  type?: 'success' | 'error' | 'warning' | 'info'
  duration?: number
}

let toastContainer: HTMLDivElement | null = null
type GenericApp = App<Element>

let toastQueue: Array<{ app: GenericApp; container: HTMLDivElement }> = []

function getToastContainer() {
  if (!toastContainer) {
    toastContainer = document.createElement('div')
    toastContainer.id = 'toast-container'
    document.body.appendChild(toastContainer)
  }
  return toastContainer
}

function showToast(options: ToastOptions) {
  const container = document.createElement('div')
  getToastContainer().appendChild(container)

  const app = createApp(Toast, {
    ...options,
    onClose: () => {
      app.unmount()
      container.remove()
      toastQueue = toastQueue.filter(item => item.app !== app)
    }
  })

  app.mount(container)
  toastQueue.push({ app, container })
}

export const toast = {
  success(message: string, duration?: number) {
    showToast({ message, type: 'success', duration })
  },
  error(message: string, duration?: number) {
    showToast({ message, type: 'error', duration })
  },
  warning(message: string, duration?: number) {
    showToast({ message, type: 'warning', duration })
  },
  info(message: string, duration?: number) {
    showToast({ message, type: 'info', duration })
  }
}

// Confirm 确认框
interface ConfirmOptions {
  title?: string
  message: string
  type?: 'warning' | 'danger' | 'info'
  confirmText?: string
  cancelText?: string
}

function showConfirm(options: ConfirmOptions): Promise<boolean> {
  return new Promise((resolve) => {
    const container = document.createElement('div')
    document.body.appendChild(container)

    const app = createApp(ConfirmDialog, {
      ...options,
      onConfirm: () => {
        resolve(true)
        setTimeout(() => {
          app.unmount()
          container.remove()
        }, 300)
      },
      onCancel: () => {
        resolve(false)
        setTimeout(() => {
          app.unmount()
          container.remove()
        }, 300)
      }
    })

    app.mount(container)
  })
}

export const confirm = {
  show(options: ConfirmOptions): Promise<boolean> {
    return showConfirm(options)
  },
  warning(message: string, title?: string): Promise<boolean> {
    return showConfirm({ message, title, type: 'warning' })
  },
  danger(message: string, title?: string): Promise<boolean> {
    return showConfirm({ message, title: title || '确认删除', type: 'danger', confirmText: '删除' })
  },
  info(message: string, title?: string): Promise<boolean> {
    return showConfirm({ message, title, type: 'info' })
  }
}

// Vue 插件
export default {
  install(app: App) {
    app.config.globalProperties.$toast = toast
    app.config.globalProperties.$confirm = confirm
  }
}
