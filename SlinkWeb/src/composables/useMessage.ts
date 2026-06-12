import { getCurrentInstance } from 'vue'
import { toast, confirm } from '../plugins/message'

export const useMessage = () => {
  const instance = getCurrentInstance()
  const globalProperties = instance?.appContext.config.globalProperties

  const toastApi = globalProperties?.$toast ?? toast
  const confirmApi = globalProperties?.$confirm ?? confirm

  return {
    toast: toastApi,
    confirm: confirmApi,
  }
}

export type UseMessageReturn = ReturnType<typeof useMessage>
