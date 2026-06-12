/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 可选：API 根地址。默认同源（空字符串），适用于 Docker/反代同域。例：https://api.example.com */
  readonly VITE_API_BASE?: string
}

import type { toast, confirm } from './src/plugins/message'

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $toast: typeof toast
    $confirm: typeof confirm
  }
}
