import { defineStore } from 'pinia'
import { ref } from 'vue'

/** 首页选择文件后带入后台上传页的待上传队列（内存，刷新即清空） */
export const usePendingUploadStore = defineStore('pendingUpload', () => {
  const files = ref<File[]>([])

  function setFiles(newFiles: File[] | FileList) {
    const arr = Array.from(newFiles).filter((f) => f.type.startsWith('image/'))
    files.value = arr
  }

  /** 取出并清空，供 UploadPage 消费 */
  function takeFiles(): File[] {
    const out = [...files.value]
    files.value = []
    return out
  }

  return { files, setFiles, takeFiles }
})
