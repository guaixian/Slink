<template>
  <div class="watermark-page">
    <div class="wm-card">
      <div class="wm-header">
        <div>
          <h2 class="wm-title">
            <i class="fa fa-tint mr-2 text-primary"></i>
            图片水印
          </h2>
          <p class="wm-desc">
            在浏览器中按「上传策略」里的水印配置叠加水印后再上传。GIF 仅处理首帧；字体/水印图需可通过本站
            <code class="mono">/static/</code> 访问。
          </p>
        </div>
        <div class="strategy-row">
          <label class="strategy-label"><i class="fa fa-database mr-2"></i>存储策略</label>
          <select v-model.number="selectedStrategy" class="strategy-select">
            <option v-for="s in strategies" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
      </div>

      <div v-if="loadError" class="banner banner-err">{{ loadError }}</div>
      <div v-else-if="policyLoaded" class="banner" :class="policy?.is_enable_watermark ? 'banner-ok' : 'banner-warn'">
        <template v-if="policy?.is_enable_watermark">
          已同步上传策略：水印驱动 <strong>{{ policy?.watermark_configs?.driver || 'font' }}</strong>
        </template>
        <template v-else>
          上传策略中未启用水印，仍将使用策略 JSON 中的 <code>watermark_configs</code> 作为默认参数（可在「上传策略」中开启）。
        </template>
      </div>

      <div class="wm-body">
        <div class="wm-col">
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

          <label v-if="(policy?.watermark_configs?.driver || 'font') === 'font'" class="field-label mt-4"
            >覆盖文字（可选，留空用策略里的 text）</label
          >
          <input
            v-if="(policy?.watermark_configs?.driver || 'font') === 'font'"
            v-model="textOverride"
            type="text"
            class="text-input"
            placeholder="策略中的默认文案"
          />

          <button type="button" class="btn-primary mt-4" :disabled="!file || busy" @click="processAndUpload">
            <i class="fa" :class="busy ? 'fa-spinner fa-spin' : 'fa-cloud-upload'"></i>
            {{ busy ? '处理并上传…' : '应用水印并上传' }}
          </button>
        </div>

        <div class="wm-col">
          <label class="field-label">结果预览</label>
          <div class="preview-box">
            <img v-if="previewUrl" :src="previewUrl" alt="preview" class="preview-img" />
            <div v-else class="preview-placeholder">上传成功后显示链接</div>
          </div>
          <div v-if="lastLink" class="link-row">
            <input :value="lastLink" readonly class="link-input" />
            <button type="button" class="btn-copy" @click="copyLink"><i class="fa fa-copy"></i></button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { userAPI, imageAPI } from '@/api'
import { useMessage } from '@/composables/useMessage'
import { watermarkedFileFromPolicy, type GroupPolicyLike } from '@/utils/watermarkCanvas'

const { toast } = useMessage()

const fileRef = ref<HTMLInputElement | null>(null)
const file = ref<File | null>(null)
const dragOver = ref(false)
const busy = ref(false)
const policy = ref<GroupPolicyLike | null>(null)
const policyLoaded = ref(false)
const loadError = ref('')
const textOverride = ref('')
const strategies = ref<Array<{ id: number; name: string }>>([])
const selectedStrategy = ref(1)
const previewUrl = ref('')
const lastLink = ref('')

onMounted(async () => {
  try {
    const res = await imageAPI.getUploadGroupConfig()
    if (res.data?.status && res.data.data) {
      policy.value = res.data.data as GroupPolicyLike
    }
  } catch (e: any) {
    loadError.value = e?.response?.data?.error || '无法读取上传策略（水印配置）'
  } finally {
    policyLoaded.value = true
  }

  try {
    const res = await userAPI.getUserInfo()
    if (res.data?.status && res.data.data) {
      const d = res.data.data
      if (d.strategies?.length) {
        strategies.value = d.strategies.map((s: { id: number; name: string }) => ({
          id: s.id,
          name: s.name
        }))
      }
      if (d.config?.default_strategy) {
        selectedStrategy.value = d.config.default_strategy
      }
    }
  } catch {
    /* 策略列表可选 */
  }
})

onUnmounted(() => {
  if (previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value)
  }
})

function onPick(e: Event) {
  const t = e.target as HTMLInputElement
  const f = t.files?.[0]
  if (f && f.type.startsWith('image/')) {
    file.value = f
    lastLink.value = ''
    if (previewUrl.value.startsWith('blob:')) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

function onDrop(e: DragEvent) {
  dragOver.value = false
  const f = e.dataTransfer?.files?.[0]
  if (f && f.type.startsWith('image/')) {
    file.value = f
    lastLink.value = ''
    if (previewUrl.value.startsWith('blob:')) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

async function processAndUpload() {
  if (!file.value || !policy.value || busy.value) return
  busy.value = true
  try {
    const out = await watermarkedFileFromPolicy(
      file.value,
      policy.value,
      textOverride.value.trim() || undefined
    )
    const res = await imageAPI.uploadImage(out, undefined, selectedStrategy.value)
    if (res.data?.status && res.data.data?.links?.url) {
      toast.success('已上传')
      lastLink.value = res.data.data.links.url
      if (previewUrl.value.startsWith('blob:')) URL.revokeObjectURL(previewUrl.value)
      previewUrl.value = URL.createObjectURL(out)
    } else {
      toast.error(res.data?.message || '上传失败')
    }
  } catch (e: any) {
    toast.error(e?.message || e?.response?.data?.error || '处理失败')
  } finally {
    busy.value = false
  }
}

async function copyLink() {
  if (!lastLink.value) return
  try {
    await navigator.clipboard.writeText(lastLink.value)
    toast.success('已复制')
  } catch {
    toast.error('复制失败')
  }
}
</script>

<style scoped>
.watermark-page {
  padding: 24px;
  max-width: 960px;
  margin: 0 auto;
}

.wm-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.wm-header {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.wm-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 8px;
}

.wm-desc {
  font-size: 13px;
  color: #6b7280;
  max-width: 520px;
  line-height: 1.5;
}

.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  background: #f3f4f6;
  padding: 1px 6px;
  border-radius: 4px;
}

.strategy-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.strategy-label {
  font-size: 13px;
  color: #4b5563;
}

.strategy-select {
  min-width: 200px;
  padding: 8px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.banner {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  margin-bottom: 16px;
}

.banner-ok {
  background: #ecfdf5;
  color: #047857;
}

.banner-warn {
  background: #fffbeb;
  color: #b45309;
}

.banner-err {
  background: #fef2f2;
  color: #b91c1c;
}

.wm-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

@media (max-width: 768px) {
  .wm-body {
    grid-template-columns: 1fr;
  }
}

.wm-col {
  display: flex;
  flex-direction: column;
}

.field-label {
  font-size: 13px;
  color: #4b5563;
  margin-bottom: 8px;
}

.drop-zone {
  border: 2px dashed #e5e7eb;
  border-radius: 12px;
  padding: 32px 16px;
  text-align: center;
  cursor: pointer;
  color: #6b7280;
  transition: border-color 0.15s, background 0.15s;
}

.drop-zone:hover,
.drop-zone.active {
  border-color: #3b82f6;
  background: #f8fafc;
}

.text-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
}

.btn-primary {
  padding: 10px 18px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.preview-box {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  overflow: hidden;
}

.preview-img {
  max-width: 100%;
  max-height: 320px;
  object-fit: contain;
}

.preview-placeholder {
  color: #9ca3af;
  font-size: 14px;
}

.link-row {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.link-input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 13px;
}

.btn-copy {
  padding: 8px 14px;
  border: 1px solid #e5e7eb;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
}

.mt-4 {
  margin-top: 16px;
}
</style>
