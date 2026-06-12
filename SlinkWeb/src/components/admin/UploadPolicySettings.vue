<template>
  <div class="upload-policy-page">
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800 mb-2">
        <i class="fa fa-upload mr-3 text-primary"></i>
        上传策略
      </h1>
      <p class="text-neutral-600 text-sm">
        全站上传限流、大小、命名规则与水印等（数据保存在数据库表 global_upload_policies）
      </p>
    </div>

    <div v-if="loading" class="loading-state">
      <i class="fa fa-spinner fa-spin"></i>
      <span>加载中...</span>
    </div>

    <div v-else class="settings-card">
      <section class="settings-section">
        <h2 class="section-title"><i class="fa fa-tachometer mr-2 text-primary"></i>频率限制</h2>
        <div class="field-grid">
          <label class="field">每分钟 <input v-model.number="policy.limit_per_minute" type="number" min="0" class="text-input" /></label>
          <label class="field">每小时 <input v-model.number="policy.limit_per_hour" type="number" min="0" class="text-input" /></label>
          <label class="field">每天 <input v-model.number="policy.limit_per_day" type="number" min="0" class="text-input" /></label>
          <label class="field">每周 <input v-model.number="policy.limit_per_week" type="number" min="0" class="text-input" /></label>
          <label class="field">每月 <input v-model.number="policy.limit_per_month" type="number" min="0" class="text-input" /></label>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title"><i class="fa fa-image mr-2 text-primary"></i>文件与并发</h2>
        <div class="field-grid">
          <label class="field wide">单文件最大（KB） <input v-model.number="policy.maximum_file_size" type="number" min="0" class="text-input" /></label>
          <label class="field wide">并发上传数 <input v-model.number="policy.concurrent_upload_num" type="number" min="0" class="text-input" /></label>
        </div>
        <label class="field block mt-3">允许后缀（逗号分隔，不含点） <input v-model="suffixesInput" type="text" class="text-input" placeholder="jpeg,jpg,png,webp" /></label>
      </section>

      <section class="settings-section">
        <h2 class="section-title"><i class="fa fa-folder-open mr-2 text-primary"></i>命名规则</h2>
        <label class="field block">路径规则 <input v-model="policy.path_naming_rule" type="text" class="text-input" placeholder="{Y}/{m}/{d}" /></label>
        <label class="field block mt-3">文件规则 <input v-model="policy.file_naming_rule" type="text" class="text-input" placeholder="{uniqid}" /></label>
      </section>

      <section class="settings-section">
        <h2 class="section-title"><i class="fa fa-image mr-2 text-primary"></i>图片保存</h2>
        <label class="field block">保存格式（留空为默认） <input v-model="imageFormatInput" type="text" class="text-input" placeholder="可选" /></label>
        <label class="field block mt-3">JPEG 质量（0–100） <input v-model.number="policy.image_save_quality" type="number" min="0" max="100" class="text-input" /></label>
      </section>

      <section class="settings-section">
        <h2 class="section-title"><i class="fa fa-lock mr-2 text-primary"></i>原图保护</h2>
        <div class="field-grid">
          <label class="field">启用 <select v-model.number="policy.is_enable_original_protection" class="select-input"><option :value="0">否</option><option :value="1">是</option></select></label>
          <label class="field wide">缓存 TTL（秒） <input v-model.number="policy.image_cache_ttl" type="number" min="0" class="text-input" /></label>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title"><i class="fa fa-tint mr-2 text-primary"></i>水印（仅前端）</h2>
        <p class="text-neutral-500 text-xs mb-3">服务端不处理图片；配置写入数据库并由「图片水印」菜单与上传页前端读取。字体/水印图放在 <code class="mono-inline">static/</code> 目录。</p>
        <label class="field block">启用 <select v-model.number="policy.is_enable_watermark" class="select-input"><option :value="0">否</option><option :value="1">是</option></select></label>
        <label class="field block mt-3">模式（mode） <input v-model.number="wm.mode" type="number" min="0" class="text-input" /></label>
        <label class="field block mt-3">当前使用的驱动
          <select v-model="wm.driver" class="select-input">
            <option value="font">文字水印（font）</option>
            <option value="image">图片水印（image）</option>
          </select>
        </label>

        <div class="wm-subcard mt-4">
          <h3 class="wm-subtitle"><i class="fa fa-font mr-2"></i>文字水印参数（drivers.font）</h3>
          <div class="field-grid">
            <label class="field">边距 X <input v-model.number="wm.font.x" type="number" class="text-input" /></label>
            <label class="field">边距 Y <input v-model.number="wm.font.y" type="number" class="text-input" /></label>
            <label class="field wide">字体文件（相对 static，如 2.ttf） <input v-model="wm.font.font" type="text" class="text-input" placeholder="2.ttf" /></label>
            <label class="field">字号 <input v-model.number="wm.font.size" type="number" min="1" step="any" class="text-input" /></label>
            <label class="field wide">文字内容 <input v-model="wm.font.text" type="text" class="text-input" /></label>
            <label class="field">旋转角度 <input v-model.number="wm.font.angle" type="number" class="text-input" /></label>
            <label class="field">颜色（#RRGGBB） <input v-model="wm.font.color" type="text" class="text-input" placeholder="#ffffff" /></label>
            <label class="field wide">位置
              <select v-model="wm.font.position" class="select-input">
                <option v-for="p in positions" :key="p.v" :value="p.v">{{ p.label }}</option>
              </select>
            </label>
          </div>
        </div>

        <div class="wm-subcard mt-4">
          <h3 class="wm-subtitle"><i class="fa fa-image mr-2"></i>图片水印参数（drivers.image）</h3>
          <div class="field-grid">
            <label class="field">边距 X <input v-model.number="wm.image.x" type="number" class="text-input" /></label>
            <label class="field">边距 Y <input v-model.number="wm.image.y" type="number" class="text-input" /></label>
            <label class="field wide">水印图路径（相对 static） <input v-model="wm.image.image" type="text" class="text-input" placeholder="watermark.png" /></label>
            <label class="field">宽度（0 为原图宽） <input v-model.number="wm.image.width" type="number" min="0" class="text-input" /></label>
            <label class="field">高度（0 为原图高） <input v-model.number="wm.image.height" type="number" min="0" class="text-input" /></label>
            <label class="field">旋转 <input v-model.number="wm.image.rotate" type="number" class="text-input" /></label>
            <label class="field">不透明度（0–100） <input v-model.number="wm.image.opacity" type="number" min="0" max="100" step="1" class="text-input" /></label>
            <label class="field wide">位置
              <select v-model="wm.image.position" class="select-input">
                <option v-for="p in positions" :key="'i'+p.v" :value="p.v">{{ p.label }}</option>
              </select>
            </label>
          </div>
        </div>
      </section>

      <div class="settings-actions">
        <button type="button" class="save-button" :disabled="saving" @click="handleSave">
          <i v-if="saving" class="fa fa-spinner fa-spin mr-2"></i>
          <i v-else class="fa fa-save mr-2"></i>
          {{ saving ? '保存中…' : '保存' }}
        </button>
        <button type="button" class="reset-button" :disabled="saving" @click="reload">
          <i class="fa fa-undo mr-2"></i>
          重新加载
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminAPI } from '@/api'
import { useMessage } from '@/composables/useMessage.ts'

const { toast } = useMessage()
const loading = ref(true)
const saving = ref(false)

const policy = reactive({
  limit_per_minute: 99999,
  limit_per_hour: 99997,
  limit_per_day: 99999,
  limit_per_week: 99999,
  limit_per_month: 99999,
  maximum_file_size: 15120,
  concurrent_upload_num: 10,
  file_naming_rule: '{uniqid}',
  path_naming_rule: '{Y}/{m}/{d}',
  image_save_quality: 75,
  is_enable_original_protection: 0,
  image_cache_ttl: 2626560,
  is_enable_watermark: 0
})

const suffixesInput = ref('jpeg,jpg,png,gif,tif,bmp,ico,psd,webp')
const imageFormatInput = ref('')

const positions = [
  { v: 'top-left', label: '左上' },
  { v: 'top-right', label: '右上' },
  { v: 'bottom-left', label: '左下' },
  { v: 'bottom-right', label: '右下' },
  { v: 'center', label: '居中' }
]

const wm = reactive({
  mode: 1,
  driver: 'font' as 'font' | 'image',
  font: {
    x: 10,
    y: 10,
    font: '',
    size: 24,
    text: '',
    angle: 0,
    color: '#ffffff',
    position: 'bottom-right'
  },
  image: {
    x: 10,
    y: 10,
    image: '',
    width: 0,
    height: 0,
    rotate: 0,
    opacity: 100,
    position: 'bottom-right'
  }
})

function defaultWmFromPolicy(d: Record<string, unknown> | undefined) {
  const wc = (d?.watermark_configs ?? {}) as Record<string, unknown>
  wm.mode = typeof wc.mode === 'number' ? wc.mode : 1
  wm.driver = wc.driver === 'image' ? 'image' : 'font'

  const drivers = (wc.drivers ?? {}) as Record<string, Record<string, unknown>>
  const f = drivers.font ?? {}
  wm.font.x = Number(f.x ?? 10)
  wm.font.y = Number(f.y ?? 10)
  wm.font.font = String(f.font ?? '')
  wm.font.size = Number(f.size ?? 24)
  wm.font.text = String(f.text ?? '')
  wm.font.angle = Number(f.angle ?? 0)
  wm.font.color = String(f.color ?? '#ffffff')
  wm.font.position = String(f.position ?? 'bottom-right')

  const im = drivers.image ?? {}
  wm.image.x = Number(im.x ?? 10)
  wm.image.y = Number(im.y ?? 10)
  wm.image.image = im.image != null && im.image !== '' ? String(im.image) : ''
  wm.image.width = Number(im.width ?? 0)
  wm.image.height = Number(im.height ?? 0)
  wm.image.rotate = Number(im.rotate ?? 0)
  wm.image.opacity = Number(im.opacity ?? 100)
  wm.image.position = String(im.position ?? 'bottom-right')
}

function applySuffixesFromPolicy(arr: string[] | undefined) {
  if (!arr || !arr.length) {
    suffixesInput.value = ''
    return
  }
  suffixesInput.value = arr.join(',')
}

async function reload() {
  loading.value = true
  try {
    const res = await adminAPI.getUploadPolicy()
    if (!res.data?.status || !res.data.data) {
      toast.error(res.data?.message || '加载失败')
      return
    }
    const d = res.data.data as Record<string, unknown>
    Object.assign(policy, {
      limit_per_minute: d.limit_per_minute ?? 0,
      limit_per_hour: d.limit_per_hour ?? 0,
      limit_per_day: d.limit_per_day ?? 0,
      limit_per_week: d.limit_per_week ?? 0,
      limit_per_month: d.limit_per_month ?? 0,
      maximum_file_size: d.maximum_file_size ?? 0,
      concurrent_upload_num: d.concurrent_upload_num ?? 0,
      file_naming_rule: d.file_naming_rule ?? '',
      path_naming_rule: d.path_naming_rule ?? '',
      image_save_quality: d.image_save_quality ?? 75,
      is_enable_original_protection: d.is_enable_original_protection ?? 0,
      image_cache_ttl: d.image_cache_ttl ?? 0,
      is_enable_watermark: d.is_enable_watermark ?? 0
    })
    imageFormatInput.value = (d.image_save_format as string) ?? ''
    applySuffixesFromPolicy(d.accepted_file_suffixes as string[] | undefined)
    defaultWmFromPolicy(d)
  } catch (e: any) {
    toast.error(e?.response?.data?.error || e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function buildWatermarkConfigsPayload() {
  return {
    mode: wm.mode,
    driver: wm.driver,
    drivers: {
      font: { ...wm.font },
      image: { ...wm.image }
    }
  }
}

async function handleSave() {
  const accepted = suffixesInput.value
    .split(',')
    .map(s => s.trim().toLowerCase().replace(/^\./, ''))
    .filter(Boolean)

  const payload: Record<string, unknown> = {
    ...policy,
    image_save_format: imageFormatInput.value.trim() || null,
    accepted_file_suffixes: accepted,
    watermark_configs: buildWatermarkConfigsPayload()
  }

  saving.value = true
  try {
    const res = await adminAPI.updateUploadPolicy(payload)
    if (res.data?.status) {
      toast.success('已保存')
      await reload()
    } else {
      toast.error(res.data?.message || '保存失败')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(reload)
</script>

<style scoped>
.upload-policy-page {
  padding: 24px;
  max-width: 880px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  padding: 24px;
}

.settings-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.settings-section {
  margin-bottom: 28px;
  padding-bottom: 24px;
  border-bottom: 1px solid #f3f4f6;
}

.settings-section:last-of-type {
  border-bottom: none;
}

.section-title {
  font-size: 17px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: #4b5563;
}

.field.block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field.wide {
  grid-column: span 2;
}

.text-input,
.select-input {
  padding: 10px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
}

.mono-inline {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  background: #f3f4f6;
  padding: 1px 6px;
  border-radius: 4px;
}

.wm-subcard {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 16px;
}

.wm-subtitle {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 12px;
}

.settings-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.save-button {
  padding: 10px 20px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
}

.save-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.reset-button {
  padding: 10px 20px;
  background: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.mt-3 {
  margin-top: 12px;
}

.mt-4 {
  margin-top: 16px;
}
</style>
