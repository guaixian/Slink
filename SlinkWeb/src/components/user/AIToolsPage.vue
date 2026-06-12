<template>
  <div class="tools-page">
    <div class="tool-card">
      <div class="tool-header">
        <div>
          <h2 class="tool-title">
            <i class="fa fa-flask mr-2 text-primary"></i>
            AI 图片处理
          </h2>
          <p class="tool-desc">
            智能去水印、AI 超分辨率（高清放大）、智能识别打标。这些能力依赖真实模型，由
            <strong>外置扩展服务</strong>提供（非本程序内置）。
          </p>
        </div>
      </div>

      <!-- 扩展状态横幅 -->
      <div v-if="statusLoaded" class="banner" :class="configured ? 'banner-ok' : 'banner-warn'">
        <template v-if="configured">
          <i class="fa fa-check-circle mr-1"></i>
          外置 AI 扩展已配置：<code class="mono">{{ endpoint }}</code>
        </template>
        <template v-else>
          <i class="fa fa-exclamation-triangle mr-1"></i>
          未配置外置 AI 扩展。可在后台<strong>「AI 设置」</strong>页填写图像扩展服务地址，或设置环境变量
          <code class="mono">SLINK_AI_ENDPOINT</code>；对接说明见
          <code class="mono">docs/ai-extension.md</code>。仅需压缩/格式转换/高清缩放可改用「图片处理」页。
        </template>
      </div>

      <!-- 能力切换 -->
      <div class="cap-tabs">
        <button
          v-for="cap in caps"
          :key="cap.key"
          class="cap-tab"
          :class="{ active: activeCap === cap.key }"
          @click="switchCap(cap.key)"
        >
          <i class="fa mr-1" :class="cap.icon"></i>{{ cap.label }}
        </button>
      </div>

      <div class="tool-body">
        <!-- 左侧：输入与参数 -->
        <div class="tool-col">
          <label class="field-label">选择图片</label>
          <div
            class="drop-zone"
            :class="{ active: dragOver, disabled: !configured }"
            @dragenter.prevent="configured && (dragOver = true)"
            @dragover.prevent="configured && (dragOver = true)"
            @dragleave.prevent="dragOver = false"
            @drop.prevent="onDrop"
            @click="configured && fileRef?.click()"
          >
            <i class="fa fa-image text-3xl text-neutral-300 mb-2"></i>
            <span v-if="!file">点击或拖拽图片到此处</span>
            <span v-else class="text-sm">{{ file.name }}</span>
          </div>
          <input ref="fileRef" type="file" accept="image/*" class="hidden" @change="onPick" />

          <!-- 超分倍数 -->
          <template v-if="activeCap === 'upscale'">
            <label class="field-label mt-4">放大倍数：{{ scale }}x</label>
            <input v-model.number="scale" type="range" min="2" max="4" step="1" class="range-input" :disabled="!configured" />
          </template>

          <!-- 去水印提示 -->
          <template v-if="activeCap === 'dewatermark'">
            <p class="hint mt-3">
              <i class="fa fa-info-circle mr-1"></i>
              水印区域由外部服务自动检测；如外部服务支持手动区域，可后续扩展传参。
            </p>
          </template>

          <button
            type="button"
            class="btn-primary mt-5"
            :disabled="!file || busy || !configured"
            @click="run"
          >
            <i class="fa" :class="busy ? 'fa-spinner fa-spin' : runIcon"></i>
            {{ busy ? '处理中…' : runLabel }}
          </button>
          <p v-if="!configured && statusLoaded" class="hint mt-2">外置扩展未配置，操作已禁用。</p>
        </div>

        <!-- 右侧：结果 -->
        <div class="tool-col">
          <label class="field-label">结果</label>

          <!-- 图像类结果（去水印/超分） -->
          <template v-if="activeCap !== 'tag'">
            <div class="preview-box">
              <img v-if="resultUri" :src="resultUri" alt="result" class="preview-img" />
              <div v-else class="preview-placeholder">处理后在此显示</div>
            </div>
            <button v-if="resultUri" type="button" class="btn-secondary mt-3" @click="download">
              <i class="fa fa-download mr-1"></i> 下载结果
            </button>
          </template>

          <!-- 识别类结果（打标） -->
          <template v-else>
            <div class="preview-box tags-box">
              <div v-if="tags.length" class="tag-list">
                <span v-for="t in tags" :key="t.name" class="tag-chip">
                  {{ t.name }}
                  <em>{{ (t.score * 100).toFixed(0) }}%</em>
                </span>
              </div>
              <div v-else class="preview-placeholder">识别后在此显示标签</div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { imageAPI } from '../../api'
import type { AITag } from '../../api'
import { useMessage } from '../../composables/useMessage'

const { toast } = useMessage()

type CapKey = 'dewatermark' | 'upscale' | 'tag'

const caps: Array<{ key: CapKey; label: string; icon: string }> = [
  { key: 'dewatermark', label: '智能去水印', icon: 'fa-eraser' },
  { key: 'upscale', label: 'AI 超分辨率', icon: 'fa-expand' },
  { key: 'tag', label: '智能打标', icon: 'fa-tags' },
]

const fileRef = ref<HTMLInputElement | null>(null)
const file = ref<File | null>(null)
const dragOver = ref(false)
const busy = ref(false)
const activeCap = ref<CapKey>('dewatermark')

const statusLoaded = ref(false)
const configured = ref(false)
const endpoint = ref('')

const scale = ref(2)
const resultUri = ref('')
const resultMime = ref('image/png')
const tags = ref<AITag[]>([])

const runLabel = computed(() => {
  switch (activeCap.value) {
    case 'upscale':
      return '开始超分'
    case 'tag':
      return '识别标签'
    default:
      return '去除水印'
  }
})

const runIcon = computed(() => {
  switch (activeCap.value) {
    case 'upscale':
      return 'fa-expand'
    case 'tag':
      return 'fa-tags'
    default:
      return 'fa-eraser'
  }
})

onMounted(async () => {
  try {
    const res = await imageAPI.getAIStatus()
    if (res.data?.status && res.data.data) {
      configured.value = res.data.data.configured
      endpoint.value = res.data.data.endpoint
    }
  } catch {
    configured.value = false
  } finally {
    statusLoaded.value = true
  }
})

const switchCap = (key: CapKey) => {
  activeCap.value = key
  resultUri.value = ''
  tags.value = []
}

const onPick = (e: Event) => {
  const t = e.target as HTMLInputElement
  if (t.files && t.files[0]) setFile(t.files[0])
}

const onDrop = (e: DragEvent) => {
  dragOver.value = false
  if (!configured.value) return
  const f = e.dataTransfer?.files?.[0]
  if (f) setFile(f)
}

const setFile = (f: File) => {
  if (!f.type.startsWith('image/')) {
    toast.warning('请选择图片文件')
    return
  }
  file.value = f
  resultUri.value = ''
  tags.value = []
}

const run = async () => {
  if (!file.value || busy.value || !configured.value) return
  busy.value = true
  try {
    if (activeCap.value === 'tag') {
      const res = await imageAPI.aiTag(file.value)
      if (res.data?.status) {
        tags.value = res.data.data?.tags || []
        toast.success(`识别到 ${tags.value.length} 个标签`)
      } else {
        toast.error(res.data?.message || '识别失败')
      }
    } else {
      const call = activeCap.value === 'upscale'
        ? imageAPI.aiUpscale(file.value, { scale: scale.value })
        : imageAPI.aiDewatermark(file.value)
      const res = await call
      if (res.data?.status && res.data.data) {
        resultUri.value = res.data.data.data_uri
        resultMime.value = res.data.data.mimetype || 'image/png'
        toast.success('处理完成')
      } else {
        toast.error(res.data?.message || '处理失败')
      }
    }
  } catch (e: any) {
    // 501（未配置）会带 hint；其它错误带 message/error
    const d = e?.response?.data
    toast.error(d?.message || d?.error || e?.message || '处理失败')
  } finally {
    busy.value = false
  }
}

const download = () => {
  if (!resultUri.value) return
  const a = document.createElement('a')
  a.href = resultUri.value
  const ext = resultMime.value.includes('png') ? 'png' : resultMime.value.includes('jpeg') ? 'jpg' : 'png'
  a.download = `${activeCap.value}_result.${ext}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
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
  margin-bottom: 16px;
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

.banner {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 16px;
}

.banner-ok {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
}

.banner-warn {
  background: #fffbeb;
  border: 1px solid #fde68a;
  color: #92400e;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  background: rgba(0, 0, 0, 0.05);
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 12px;
}

.cap-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.cap-tab {
  padding: 8px 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
  color: #6b7280;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.cap-tab:hover {
  border-color: #3b82f6;
  color: #3b82f6;
}

.cap-tab.active {
  background: #eff6ff;
  border-color: #3b82f6;
  color: #3b82f6;
  font-weight: 500;
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

.mt-2 { margin-top: 8px; }
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

.drop-zone:hover:not(.disabled),
.drop-zone.active {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.drop-zone.disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.range-input {
  width: 100%;
  accent-color: #3b82f6;
}

.hint {
  font-size: 12px;
  color: #9ca3af;
  line-height: 1.5;
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

.tags-box {
  align-items: flex-start;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-chip {
  font-size: 13px;
  color: #1f2937;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  border-radius: 6px;
  padding: 4px 10px;
}

.tag-chip em {
  color: #3b82f6;
  font-style: normal;
  font-weight: 600;
  margin-left: 4px;
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
