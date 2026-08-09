<template>
  <div class="upload-policy-page">
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800 mb-2">
        <i class="fa fa-upload mr-3 text-primary"></i>
        上传策略组
      </h1>
      <p class="text-neutral-600 text-sm">
        管理上传策略组，每个策略组可分配给多个用户。默认策略组用于新注册用户。
      </p>
    </div>

    <!-- 策略组列表 -->
    <div class="groups-list">
      <div class="list-header">
        <span class="list-title">策略组列表</span>
        <button class="create-btn" @click="startCreate">
          <i class="fa fa-plus"></i> 新建策略组
        </button>
      </div>
      <table class="data-table" v-if="groups.length > 0">
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>描述</th>
            <th>用户数</th>
            <th>默认</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="g in groups" :key="g.id" :class="{ 'selected-row': editingId === g.id }">
            <td>{{ g.id }}</td>
            <td>{{ g.name }}</td>
            <td>{{ g.description }}</td>
            <td>{{ g.user_count ?? 0 }}</td>
            <td><span v-if="g.is_default === 1" class="badge-default">默认</span></td>
            <td>
              <button class="action-btn edit" @click="startEdit(g)">编辑</button>
              <button v-if="g.is_default !== 1" class="action-btn default" @click="handleSetDefault(g)">设为默认</button>
              <button class="action-btn delete" @click="handleDelete(g)" :disabled="g.is_default === 1">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-hint">暂无策略组</div>
    </div>

    <!-- 编辑表单 -->
    <div v-if="editingId !== null || creating" class="settings-card mt-4">
      <h2 class="section-title mb-4">
        {{ creating ? '新建上传策略组' : '编辑策略组' }}
      </h2>

      <div class="field block mb-4">
        <label>名称</label>
        <input v-model="editName" type="text" class="text-input" placeholder="策略组名称" />
      </div>
      <div class="field block mb-4">
        <label>描述</label>
        <input v-model="editDesc" type="text" class="text-input" placeholder="策略组描述" />
      </div>

      <section class="settings-section">
        <h2 class="section-title">频率限制</h2>
        <div class="field-grid">
          <label class="field">每分钟 <input v-model.number="policy.limit_per_minute" type="number" min="0" class="text-input" /></label>
          <label class="field">每小时 <input v-model.number="policy.limit_per_hour" type="number" min="0" class="text-input" /></label>
          <label class="field">每天 <input v-model.number="policy.limit_per_day" type="number" min="0" class="text-input" /></label>
          <label class="field">每周 <input v-model.number="policy.limit_per_week" type="number" min="0" class="text-input" /></label>
          <label class="field">每月 <input v-model.number="policy.limit_per_month" type="number" min="0" class="text-input" /></label>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title">文件与并发</h2>
        <div class="field-grid">
          <label class="field wide">单文件最大（KB） <input v-model.number="policy.maximum_file_size" type="number" min="0" class="text-input" /></label>
          <label class="field wide">并发上传数 <input v-model.number="policy.concurrent_upload_num" type="number" min="0" class="text-input" /></label>
        </div>
        <label class="field block mt-3">允许后缀（逗号分隔，不含点） <input v-model="suffixesInput" type="text" class="text-input" placeholder="jpeg,jpg,png,webp" /></label>
      </section>

      <section class="settings-section">
        <h2 class="section-title">命名规则</h2>
        <label class="field block">路径规则 <input v-model="policy.path_naming_rule" type="text" class="text-input" /></label>
        <label class="field block mt-3">文件规则 <input v-model="policy.file_naming_rule" type="text" class="text-input" /></label>
      </section>

      <section class="settings-section">
        <h2 class="section-title">原图保护</h2>
        <label class="field">启用 <select v-model.number="policy.is_enable_original_protection" class="select-input"><option :value="0">否</option><option :value="1">是</option></select></label>
        <label class="field wide mt-3">缓存TTL（秒） <input v-model.number="policy.image_cache_ttl" type="number" min="0" class="text-input" /></label>
      </section>

      <section class="settings-section">
        <h2 class="section-title">水印</h2>
        <label class="field block">启用 <select v-model.number="policy.is_enable_watermark" class="select-input"><option :value="0">否</option><option :value="1">是</option></select></label>
        <label class="field block mt-3">模式 <input v-model.number="wm.mode" type="number" min="0" class="text-input" /></label>
        <label class="field block mt-3">驱动
          <select v-model="wm.driver" class="select-input">
            <option value="font">文字水印</option>
            <option value="image">图片水印</option>
          </select>
        </label>
        <div class="wm-subcard mt-4">
          <h3 class="wm-subtitle">文字水印</h3>
          <div class="field-grid">
            <label class="field">X <input v-model.number="wm.font.x" type="number" class="text-input" /></label>
            <label class="field">Y <input v-model.number="wm.font.y" type="number" class="text-input" /></label>
            <label class="field wide">字体 <input v-model="wm.font.font" type="text" class="text-input" /></label>
            <label class="field">字号 <input v-model.number="wm.font.size" type="number" step="any" class="text-input" /></label>
            <label class="field wide">文字 <input v-model="wm.font.text" type="text" class="text-input" /></label>
            <label class="field">角度 <input v-model.number="wm.font.angle" type="number" class="text-input" /></label>
            <label class="field">颜色 <input v-model="wm.font.color" type="text" class="text-input" /></label>
            <label class="field wide">位置
              <select v-model="wm.font.position" class="select-input">
                <option v-for="p in positions" :key="p.v" :value="p.v">{{ p.label }}</option>
              </select>
            </label>
          </div>
        </div>
        <div class="wm-subcard mt-4">
          <h3 class="wm-subtitle">图片水印</h3>
          <div class="field-grid">
            <label class="field">X <input v-model.number="wm.image.x" type="number" class="text-input" /></label>
            <label class="field">Y <input v-model.number="wm.image.y" type="number" class="text-input" /></label>
            <label class="field wide">路径 <input v-model="wm.image.image" type="text" class="text-input" /></label>
            <label class="field">宽 <input v-model.number="wm.image.width" type="number" class="text-input" /></label>
            <label class="field">高 <input v-model.number="wm.image.height" type="number" class="text-input" /></label>
            <label class="field">旋转 <input v-model.number="wm.image.rotate" type="number" class="text-input" /></label>
            <label class="field">透明度 <input v-model.number="wm.image.opacity" type="number" min="0" max="100" class="text-input" /></label>
            <label class="field wide">位置
              <select v-model="wm.image.position" class="select-input">
                <option v-for="p in positions" :key="'i'+p.v" :value="p.v">{{ p.label }}</option>
              </select>
            </label>
          </div>
        </div>
      </section>

      <div class="settings-actions">
        <button class="save-button" :disabled="saving" @click="handleSave">
          <i v-if="saving" class="fa fa-spinner fa-spin mr-2"></i>
          {{ saving ? '保存中...' : '保存' }}
        </button>
        <button class="reset-button" @click="cancelEdit">取消</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminAPI } from '@/api'
import { useMessage } from '@/composables/useMessage.ts'

const { toast, confirm } = useMessage()
const loading = ref(true)
const saving = ref(false)

const groups = ref<any[]>([])
const editingId = ref<number | null>(null)
const creating = ref(false)
const editName = ref('')
const editDesc = ref('')

const policy = reactive({
  limit_per_minute: 99999, limit_per_hour: 99997, limit_per_day: 99999,
  limit_per_week: 99999, limit_per_month: 99999,
  maximum_file_size: 15120, concurrent_upload_num: 10,
  file_naming_rule: '{uniqid}', path_naming_rule: '{Y}/{m}/{d}',
  is_enable_original_protection: 0,
  image_cache_ttl: 2626560, is_enable_watermark: 0
})

const suffixesInput = ref('jpeg,jpg,png,gif,tif,bmp,ico,psd,webp')

const positions = [
  { v: 'top-left', label: '左上' }, { v: 'top-right', label: '右上' },
  { v: 'bottom-left', label: '左下' }, { v: 'bottom-right', label: '右下' },
  { v: 'center', label: '居中' }
]

const wm = reactive({
  mode: 1, driver: 'font' as 'font' | 'image',
  font: { x: 10, y: 10, font: '', size: 24, text: '', angle: 0, color: '#ffffff', position: 'bottom-right' },
  image: { x: 10, y: 10, image: '', width: 0, height: 0, rotate: 0, opacity: 100, position: 'bottom-right' }
})

function resetForm() {
  Object.assign(policy, {
    limit_per_minute: 99999, limit_per_hour: 99997, limit_per_day: 99999,
    limit_per_week: 99999, limit_per_month: 99999, maximum_file_size: 15120,
    concurrent_upload_num: 10, file_naming_rule: '{uniqid}', path_naming_rule: '{Y}/{m}/{d}',
    is_enable_original_protection: 0,
    image_cache_ttl: 2626560, is_enable_watermark: 0
  })
  suffixesInput.value = 'jpeg,jpg,png,gif,tif,bmp,ico,psd,webp'
  wm.mode = 1; wm.driver = 'font'
  Object.assign(wm.font, { x: 10, y: 10, font: '', size: 24, text: '', angle: 0, color: '#ffffff', position: 'bottom-right' })
  Object.assign(wm.image, { x: 10, y: 10, image: '', width: 0, height: 0, rotate: 0, opacity: 100, position: 'bottom-right' })
}

function loadFromGroupConfig(d: Record<string, unknown>) {
  Object.assign(policy, {
    limit_per_minute: d.limit_per_minute ?? 0, limit_per_hour: d.limit_per_hour ?? 0,
    limit_per_day: d.limit_per_day ?? 0, limit_per_week: d.limit_per_week ?? 0,
    limit_per_month: d.limit_per_month ?? 0, maximum_file_size: d.maximum_file_size ?? 0,
    concurrent_upload_num: d.concurrent_upload_num ?? 0,
    file_naming_rule: d.file_naming_rule ?? '', path_naming_rule: d.path_naming_rule ?? '',
    is_enable_original_protection: d.is_enable_original_protection ?? 0,
    image_cache_ttl: d.image_cache_ttl ?? 0, is_enable_watermark: d.is_enable_watermark ?? 0
  })
  const suffixes = d.accepted_file_suffixes as string[] | undefined
  suffixesInput.value = suffixes?.length ? suffixes.join(',') : ''

  const wc = (d.watermark_configs ?? {}) as Record<string, unknown>
  wm.mode = typeof wc.mode === 'number' ? wc.mode : 1
  wm.driver = wc.driver === 'image' ? 'image' : 'font'
  const drivers = (wc.drivers ?? {}) as Record<string, Record<string, unknown>>
  const f = drivers.font ?? {}
  wm.font.x = Number(f.x ?? 10); wm.font.y = Number(f.y ?? 10)
  wm.font.font = String(f.font ?? ''); wm.font.size = Number(f.size ?? 24)
  wm.font.text = String(f.text ?? ''); wm.font.angle = Number(f.angle ?? 0)
  wm.font.color = String(f.color ?? '#ffffff'); wm.font.position = String(f.position ?? 'bottom-right')
  const im = drivers.image ?? {}
  wm.image.x = Number(im.x ?? 10); wm.image.y = Number(im.y ?? 10)
  wm.image.image = String(im.image ?? ''); wm.image.width = Number(im.width ?? 0)
  wm.image.height = Number(im.height ?? 0); wm.image.rotate = Number(im.rotate ?? 0)
  wm.image.opacity = Number(im.opacity ?? 100); wm.image.position = String(im.position ?? 'bottom-right')
}

async function loadGroups() {
  loading.value = true
  try {
    const res = await adminAPI.getPolicyGroups()
    if (res.data?.status) groups.value = res.data.data || []
  } catch (e: any) {
    toast.error('加载策略组失败')
  } finally {
    loading.value = false
  }
}

async function startCreate() {
  creating.value = true
  editingId.value = null
  editName.value = ''
  editDesc.value = ''
  resetForm()
}

async function startEdit(g: any) {
  creating.value = false
  editingId.value = g.id
  editName.value = g.name || ''
  editDesc.value = g.description || ''
  try {
    const res = await adminAPI.getPolicyGroup(g.id)
    if (res.data?.status && res.data.data) {
      loadFromGroupConfig(res.data.data as Record<string, unknown>)
    }
  } catch { /* use defaults */ }
}

function cancelEdit() {
  editingId.value = null
  creating.value = false
}

function buildPayload() {
  const accepted = suffixesInput.value.split(',').map(s => s.trim().toLowerCase().replace(/^\./, '')).filter(Boolean)
  return {
    limit_per_minute: policy.limit_per_minute, limit_per_hour: policy.limit_per_hour,
    limit_per_day: policy.limit_per_day, limit_per_week: policy.limit_per_week,
    limit_per_month: policy.limit_per_month, maximum_file_size: policy.maximum_file_size,
    concurrent_upload_num: policy.concurrent_upload_num,
    file_naming_rule: policy.file_naming_rule, path_naming_rule: policy.path_naming_rule,
    accepted_file_suffixes: accepted,
    is_enable_original_protection: policy.is_enable_original_protection,
    image_cache_ttl: policy.image_cache_ttl,
    is_enable_watermark: policy.is_enable_watermark,
    watermark_configs: { mode: wm.mode, driver: wm.driver, drivers: { font: { ...wm.font }, image: { ...wm.image } } }
  }
}

async function handleSave() {
  saving.value = true
  try {
    const config = buildPayload()
    if (creating.value) {
      const res = await adminAPI.createPolicyGroup({ name: editName.value || '新策略组', description: editDesc.value, config })
      if (res.data?.status) toast.success('创建成功')
      else { toast.error(res.data?.message || '创建失败'); return }
    } else if (editingId.value) {
      const res = await adminAPI.updatePolicyGroup(editingId.value, { name: editName.value, description: editDesc.value, config })
      if (res.data?.status) toast.success('保存成功')
      else { toast.error(res.data?.message || '保存失败'); return }
    }
    cancelEdit()
    await loadGroups()
  } catch (e: any) {
    toast.error(e?.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleSetDefault(g: any) {
  try {
    const res = await adminAPI.setDefaultPolicyGroup(g.id)
    if (res.data?.status) { toast.success(`已将 "${g.name}" 设为默认策略组`); await loadGroups() }
    else toast.error(res.data?.message || '操作失败')
  } catch (e: any) { toast.error(e?.response?.data?.message || '操作失败') }
}

async function handleDelete(g: any) {
  if (g.is_default === 1) return
  const ok = await confirm.danger(`确定删除策略组 "${g.name}"？该组下的用户将移回默认组。`, '删除策略组')
  if (!ok) return
  try {
    const res = await adminAPI.deletePolicyGroup(g.id)
    if (res.data?.status) { toast.success('已删除'); await loadGroups() }
    else toast.error(res.data?.message || '删除失败')
  } catch (e: any) {
    toast.error(e?.response?.data?.message || '删除失败')
  }
}

onMounted(loadGroups)
</script>

<style scoped>
.upload-policy-page { padding: 24px; max-width: 960px; margin: 0 auto; }
.page-header { margin-bottom: 24px; }

.groups-list { background: #fff; border-radius: 12px; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.05); margin-bottom: 16px; }
.list-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.list-title { font-size: 16px; font-weight: 600; color: #374151; }

.create-btn { padding: 8px 16px; background: #3b82f6; color: #fff; border: none; border-radius: 6px; cursor: pointer; font-size: 13px; display: flex; align-items: center; gap: 6px; }
.create-btn:hover { background: #2563eb; }

.data-table { width: 100%; border-collapse: collapse; font-size: 14px; }
.data-table th { background: #f9fafb; padding: 10px 8px; text-align: left; font-weight: 600; color: #374151; border-bottom: 1px solid #e5e7eb; }
.data-table td { padding: 10px 8px; border-bottom: 1px solid #f3f4f6; color: #374151; }
.selected-row { background: #eff6ff; }

.badge-default { padding: 2px 8px; background: #dcfce7; color: #166534; border-radius: 4px; font-size: 12px; }

.action-btn { padding: 4px 12px; border: none; border-radius: 4px; cursor: pointer; font-size: 12px; margin-right: 4px; }
.action-btn.edit { background: #e0e7ff; color: #3730a3; }
.action-btn.edit:hover { background: #c7d2fe; }
.action-btn.default { background: #dbeafe; color: #1e40af; }
.action-btn.default:hover { background: #bfdbfe; }
.action-btn.delete { background: #fee2e2; color: #991b1b; }
.action-btn.delete:hover { background: #fecaca; }
.action-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-hint { text-align: center; color: #9ca3af; padding: 24px; }

.settings-card { background: #fff; border-radius: 12px; padding: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.05); }
.settings-section { margin-bottom: 24px; padding-bottom: 20px; border-bottom: 1px solid #f3f4f6; }
.settings-section:last-of-type { border-bottom: none; }
.section-title { font-size: 16px; font-weight: 600; color: #374151; margin-bottom: 12px; }

.field-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 12px; }
.field { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #4b5563; }
.field.block { display: flex; flex-direction: column; gap: 6px; }
.field.wide { grid-column: span 2; }

.text-input, .select-input { padding: 10px 12px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 14px; }

.wm-subcard { background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 10px; padding: 16px; }
.wm-subtitle { font-size: 14px; font-weight: 600; color: #374151; margin-bottom: 12px; }

.settings-actions { display: flex; gap: 12px; margin-top: 8px; }
.save-button { padding: 10px 20px; background: #3b82f6; color: #fff; border: none; border-radius: 8px; cursor: pointer; font-weight: 500; }
.save-button:disabled { opacity: 0.6; cursor: not-allowed; }
.reset-button { padding: 10px 20px; background: #f3f4f6; color: #374151; border: none; border-radius: 8px; cursor: pointer; }

.field-hint { font-size: 12px; color: #9ca3af; margin-top: 4px; }
.mt-3 { margin-top: 12px; }
.mt-4 { margin-top: 16px; }
.mb-4 { margin-bottom: 16px; }
</style>
