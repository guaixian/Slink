/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE?: string
}

import type { toast, confirm } from './src/plugins/message'

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $toast: typeof toast
    $confirm: typeof confirm
  }
}
