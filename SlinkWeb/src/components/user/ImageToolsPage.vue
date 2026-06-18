<template>
  <div class="tools-page">
    <div class="tool-card">
      <div class="tool-header">
        <div>
          <h2 class="tool-title">
            <i class="fa fa-magic mr-2 text-primary"></i>
            图片处理
          </h2>
          <p class="tool-desc">
            纯本地处理，无需任何外部服务：压缩、格式转换、高质量缩放/放大（高清）、生成缩略图。
            处理在服务端内存中完成，结果可直接预览与下载。
          </p>
        </div>
      </div>

      <div class="tool-body">
        <!-- 左侧：输入与参数 -->
        <div class="tool-col">
          <label class="field-label">选择图片</label>
          <div
            class="drop-zone"
            :class="{ active: dragOver }"
            @dragenter.prevent="dragOver = true"
            @dragover.prevent="dragOver = true"
            @dragleave.prevent="dragOver = false"
            @drop.prevent="onDrop"
            @click="fileRef?.click()"
          >
            <i class="fa fa-image text-3xl text-neutral-300 mb-2"></i>
            <span v-if="!file">点击或拖拽图片到此处</span>
            <span v-else class="text-sm">{{ file.name }}</span>
          </div>
          <input ref="fileRef" type="file" accept="image/*" class="hidden" @change="onPick" />

          <label class="field-label mt-4">操作</label>
          <select v-model="operation" class="tool-select">
            <option value="compress">压缩（按质量重新编码）</option>
            <option value="convert">格式转换</option>
            <option value="resize">缩放 / 高清放大</option>
            <option value="thumbnail">生成缩略图</option>
          </select>

          <!-- 输出格式：压缩/转换/缩放可选 -->
          <template v-if="operation !== 'thumbnail'">
            <label class="field-label mt-4">输出格式</label>
            <select v-model="format" class="tool-select">
              <option value="">保持原格式</option>
              <option v-for="f in outputFormats" :key="f" :value="f">{{ f.toUpperCase() }}</option>
            </select>
          </template>

          <!-- 质量：输出 jpeg 时有效 -->
          <template v-if="showQuality">
            <label class="field-label mt-4">JPEG 质量：{{ quality }}</label>
            <input v-model.number="quality" type="range" min="1" max="100" class="range-input" />
          </template>

          <!-- 缩放参数 -->
          <template v-if="operation === 'resize'">
            <label class="field-label mt-4">缩放方式</label>
            <select v-model="resizeMode" class="tool-select">
              <option value="scale">按倍数</option>
              <option value="dim">按宽高</option>
            </select>
            <template v-if="resizeMode === 'scale'">
              <label class="field-label mt-4">倍数（&gt;1 放大，&lt;1 缩小）：{{ scale }}x</label>
              <input v-model.number="scale" type="range" min="0.1" max="4" step="0.1" class="range-input" />
            </template>
            <template v-else>
              <div class="dim-row">
                <div class="dim-field">
                  <label class="field-label">宽度(px)</label>
                  <input v-model.number="width" type="number" min="0" class="text-input" placeholder="留空按比例" />
                </div>
                <div class="dim-field">
                  <label class="field-label">高度(px)</label>
                  <input v-model.number="height" type="number" min="0" class="text-input" placeholder="留空按比例" />
                </div>
              </div>
            </template>
          </template>

          <!-- 缩略图参数 -->
          <template v-if="operation === 'thumbnail'">
            <label class="field-label mt-4">边界框最长边(px)：{{ maxSize }}</label>
            <input v-model.number="maxSize" type="range" min="32" max="1024" step="16" class="range-input" />
          </template>

          <button type="button" class="btn-primary mt-5" :disabled="!file || busy" @click="run">
            <i class="fa" :class="busy ? 'fa-spinner fa-spin' : 'fa-bolt'"></i>
            {{ busy ? '处理中…' : '开始处理' }}
          </button>
        </div>

        <!-- 右侧：结果 -->
        <div class="tool-col">
          <label class="field-label">结果预览</label>
          <div class="preview-box">
            <img v-if="resultUri" :src="resultUri" alt="result" class="preview-img" />
            <div v-else class="preview-placeholder">处理后在此显示</div>
          </div>
          <div v-if="result" class="result-meta">
            <span class="meta-chip">{{ result.format.toUpperCase() }}</span>
            <span class="meta-chip">{{ result.width }} × {{ result.height }}</span>
            <span class="meta-chip">{{ formatBytes(result.size_bytes) }}</span>
          </div>
          <button v-if="resultUri" type="button" class="btn-secondary mt-3" @click="download">
            <i class="fa fa-download mr-1"></i> 下载结果
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { imageAPI } from '../../api'
import type { ProcessedImage, ProcessOperation } from '../../api'
import { useMessage } from '../../composables/useMessage'

const { toast } = useMessage()

const fileRef = ref<HTMLInputElement | null>(null)
const file = ref<File | null>(null)
const dragOver = ref(false)
const busy = ref(false)

const operation = ref<ProcessOperation>('compress')
const format = ref('')
const quality = ref(85)
const resizeMode = ref<'scale' | 'dim'>('scale')
const scale = ref(2)
const width = ref<number | null>(null)
const height = ref<number | null>(null)
const maxSize = ref(256)

const outputFormats = ref<string[]>(['jpeg', 'png', 'gif', 'bmp'])
const result = ref<ProcessedImage | null>(null)
const resultUri = ref('')

// 仅当最终输出为 jpeg 时质量参数有效
const showQuality = computed(() => {
  if (operation.value === 'thumbnail') return false
  return format.value === 'jpeg' || format.value === 'jpg' || format.value === ''
})

onMounted(async () => {
  try {
    const res = await imageAPI.getProcessCapabilities()
    if (res.data?.status && res.data.data?.output_formats?.length) {
      outputFormats.value = res.data.data.output_formats
      quality.value = res.data.data.default_quality || 85
    }
  } catch {
    // 能力接口失败不阻塞，使用默认值
  }
})

const onPick = (e: Event) => {
  const t = e.target as HTMLInputElement
  if (t.files && t.files[0]) setFile(t.files[0])
}

const onDrop = (e: DragEvent) => {
  dragOver.value = false
  const f = e.dataTransfer?.files?.[0]
  if (f) setFile(f)
}

const setFile = (f: File) => {
  if (!f.type.startsWith('image/')) {
    toast.warning('请选择图片文件')
    return
  }
  file.value = f
  result.value = null
  resultUri.value = ''
}

const run = async () => {
  if (!file.value || busy.value) return
  busy.value = true
  try {
    const res = await imageAPI.processImage(file.value, {
      operation: operation.value,
      format: format.value || undefined,
      quality: showQuality.value ? quality.value : undefined,
      scale: operation.value === 'resize' && resizeMode.value === 'scale' ? scale.value : undefined,
      width: operation.value === 'resize' && resizeMode.value === 'dim' ? width.value ?? undefined : undefined,
      height: operation.value === 'resize' && resizeMode.value === 'dim' ? height.value ?? undefined : undefined,
      max_size: operation.value === 'thumbnail' ? maxSize.value : undefined,
    })
    if (res.data?.status && res.data.data) {
      result.value = res.data.data
      resultUri.value = res.data.data.data_uri
      toast.success('处理完成')
    } else {
      toast.error(res.data?.message || '处理失败')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || e?.message || '处理失败')
  } finally {
    busy.value = false
  }
}

const download = () => {
  if (!result.value) return
  const a = document.createElement('a')
  a.href = resultUri.value
  const base = (result.value.origin_name || 'image').replace(/\.[^.]+$/, '')
  a.download = `${base}_processed.${result.value.format === 'jpeg' ? 'jpg' : result.value.format}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

const formatBytes = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}
</script>

<style scoped>
.tools-page {
  max-width: 1100px;
  margin: 0 auto;
}

.tool-card {
  background: #fff;
  border-radius: 12px;
  padding: 28px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.tool-header {
  margin-bottom: 20px;
}

.tool-title {
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 6px;
}

.tool-desc {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.6;
}

.tool-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 28px;
}

.tool-col {
  display: flex;
  flex-direction: column;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 6px;
}

.mt-3 { margin-top: 12px; }
.mt-4 { margin-top: 16px; }
.mt-5 { margin-top: 20px; }

.drop-zone {
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 36px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fafafa;
}

.drop-zone:hover,
.drop-zone.active {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.tool-select,
.text-input {
  padding: 9px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background: #fff;
  color: #374151;
  width: 100%;
  transition: all 0.2s ease;
}

.tool-select:focus,
.text-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.range-input {
  width: 100%;
  accent-color: #3b82f6;
}

.dim-row {
  display: flex;
  gap: 12px;
}

.dim-field {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.btn-primary {
  padding: 11px 20px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

.btn-primary:disabled {
  background: #9ca3af;
  cursor: not-allowed;
  opacity: 0.7;
}

.btn-secondary {
  padding: 9px 16px;
  background: #10b981;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  align-self: flex-start;
}

.btn-secondary:hover {
  background: #059669;
  transform: translateY(-1px);
}

.preview-box {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f9fafb;
  min-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 12px;
}

.preview-img {
  max-width: 100%;
  max-height: 360px;
  object-fit: contain;
  border-radius: 4px;
}

.preview-placeholder {
  color: #9ca3af;
  font-size: 14px;
}

.result-meta {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.meta-chip {
  font-size: 12px;
  color: #374151;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  border-radius: 4px;
  padding: 3px 8px;
}

@media (max-width: 768px) {
  .tool-card {
    padding: 16px;
  }
  .tool-body {
    grid-template-columns: 1fr;
    gap: 20px;
  }
}
</style>
