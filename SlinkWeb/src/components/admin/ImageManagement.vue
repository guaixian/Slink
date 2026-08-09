<template>
  <PageLayout 
    title="图片管理"
    description="查看、筛选与删除图库中的全部图片"
    icon-class="fa fa-cogs"
  >
    <div class="images-view">
      <div class="filters-toolbar">
        <div class="filter-item">
          <span class="filter-item-label">权限</span>
          <select v-model="selectedImageFilter" class="filter-select-sm" title="访问权限">
            <option value="all">全部</option>
            <option value="public">公开</option>
            <option value="private">私有</option>
          </select>
        </div>
        <div class="filter-item filter-item-strategy">
          <span class="filter-item-label">策略</span>
          <select v-model="selectedStrategyId" class="filter-select-sm filter-select-strategy" title="存储策略">
            <option value="all">全部</option>
            <option v-for="s in strategyOptions" :key="s.id" :value="String(s.id)">
              {{ s.name }}
            </option>
          </select>
        </div>
        <div class="filter-item">
          <span class="filter-item-label">类型</span>
          <select v-model="selectedMimeFilter" class="filter-select-sm" title="图片类型">
            <option value="all">全部</option>
            <option value="image/jpeg">JPEG</option>
            <option value="image/png">PNG</option>
            <option value="image/gif">GIF</option>
            <option value="image/webp">WebP</option>
            <option value="image/svg+xml">SVG</option>
            <option value="image/bmp">BMP</option>
            <option value="other">其它</option>
          </select>
        </div>
        <div class="filter-item">
          <span class="filter-item-label">排序</span>
          <select v-model="sortOrder" class="filter-select-sm" title="排序">
            <option value="created_desc">时间↓</option>
            <option value="created_asc">时间↑</option>
            <option value="size_desc">大小↓</option>
            <option value="size_asc">大小↑</option>
          </select>
        </div>
        <div class="filter-item filter-item-search">
          <span class="filter-item-label">搜索</span>
          <div class="search-inline">
            <input
              type="text"
              placeholder="文件名 / 路径…"
              class="search-input-compact"
              v-model="imageSearchKeyword"
              @keyup.enter="handleImageSearch"
            >
            <button type="button" class="search-btn-compact" @click="handleImageSearch" title="搜索">
              <i class="fa fa-search"></i>
            </button>
          </div>
        </div>
        <div class="filter-stats-inline" v-if="!imageLoading">
          共 <strong>{{ images.length }}</strong> · 筛选 <strong>{{ filteredImages.length }}</strong>
        </div>
      </div>

      <!-- 图片网格 -->
      <div v-if="imageLoading" class="loading-state">
        <i class="fa fa-spinner fa-spin"></i>
        <span>加载中...</span>
      </div>
      <div v-else-if="filteredImages.length === 0" class="empty-state">
        <i class="fa fa-images"></i>
        <span>暂无图片</span>
      </div>
      <div v-else class="images-grid">
        <div 
          v-for="image in paginatedImages" 
          :key="image.id" 
          class="image-card"
          @click="handleImageClick(image)"
        >
          <div class="image-container">
            <img
              :src="image.displaySrc"
              :alt="image.name"
              class="image-thumbnail"
              loading="lazy"
              @error="onImgError($event, image)"
            >
            <div class="image-overlay">
              <div class="overlay-actions">
                <button class="action-btn" @click.stop="handleView(image)">
                  <i class="fa fa-eye"></i>
                </button>
                <button class="action-btn" @click.stop="handleEdit(image)">
                  <i class="fa fa-edit"></i>
                </button>
                <button class="action-btn" @click.stop="handleDelete(image)">
                  <i class="fa fa-trash"></i>
                </button>
              </div>
            </div>
          </div>
          <div class="image-info">
            <div class="image-name">{{ image.name }}</div>
            <div class="image-meta">
              <span class="image-size">{{ image.size }}</span>
              <span class="image-date">{{ image.uploadDate }}</span>
            </div>
            <div v-if="image.strategyName" class="image-strategy" :title="image.strategyName">
              {{ image.strategyName }}
            </div>
            <div class="image-status">
              <span 
                class="status-badge" 
                :class="image.isPublic ? 'status-public' : 'status-private'"
              >
                {{ image.isPublic ? '公开' : '私有' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination">
        <button class="page-btn" :disabled="currentPage === 1" @click="prevPage">
          <i class="fa fa-chevron-left"></i>
        </button>
        <span class="page-info">第 {{ currentPage }} 页，共 {{ totalPages }} 页</span>
        <button class="page-btn" :disabled="currentPage === totalPages" @click="nextPage">
          <i class="fa fa-chevron-right"></i>
        </button>
      </div>
    </div>

    <div v-if="renameOpen" class="rename-overlay" @click.self="closeRename">
      <div class="rename-dialog" role="dialog" aria-modal="true">
        <h3 class="rename-title">修改显示名称</h3>
        <p class="rename-hint">仅更新对外展示的原始文件名，不改变已存储路径与外链路径。</p>
        <input
          v-model="renameValue"
          class="rename-input"
          type="text"
          maxlength="255"
          autocomplete="off"
          @keyup.enter="submitRename"
        >
        <div class="rename-actions">
          <button type="button" class="btn-cancel" @click="closeRename">取消</button>
          <button type="button" class="btn-save" :disabled="renameSaving" @click="submitRename">
            {{ renameSaving ? '保存中…' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { adminAPI, imageAPI } from '../../api'
import { useMessage } from '../../composables/useMessage'
import PageLayout from './PageLayout.vue'

const KNOWN_MIMES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml', 'image/bmp']

const IMG_PLACEHOLDER =
  'data:image/svg+xml,' +
  encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="120" viewBox="0 0 200 120"><rect fill="#f3f4f6" width="200" height="120"/><text x="100" y="64" fill="#9ca3af" font-size="13" font-family="sans-serif" text-anchor="middle">无法加载预览</text></svg>`
  )

const imageSearchKeyword = ref('')
const selectedImageFilter = ref<'all' | 'public' | 'private'>('all')
const selectedStrategyId = ref<string>('all')
const selectedMimeFilter = ref<string>('all')
const sortOrder = ref<'created_desc' | 'created_asc' | 'size_desc' | 'size_asc'>('created_desc')
const currentPage = ref(1)
const pageSize = ref(20)
const imageLoading = ref(false)

const renameOpen = ref(false)
const renameTarget = ref<AdminImage | null>(null)
const renameValue = ref('')
const renameSaving = ref(false)

const { toast, confirm } = useMessage()

interface AdminImage {
  id: number
  name: string
  pathname: string
  url: string
  displaySrc: string
  size: string
  sizeBytes: number
  uploadDate: string
  createdAt: string
  isPublic: boolean
  strategyId: number
  strategyName: string
  mimetype: string
}

const images = ref<AdminImage[]>([])

const fixImageUrl = (url: string): string => {
  if (!url) return ''
  return String(url).replace(/\\/g, '/')
}

/** 列表接口里的外链可能是相对路径，补全为可在 <img> 中加载的绝对地址 */
const resolveDisplaySrc = (url: string): string => {
  const u = fixImageUrl(url)
  if (!u) return ''
  if (u.startsWith('data:') || u.startsWith('blob:')) return u
  if (u.startsWith('//')) return `${window.location.protocol}${u}`
  if (u.startsWith('/')) return `${window.location.origin}${u}`
  return u
}

const loadImages = async () => {
  imageLoading.value = true
  try {
    const response = await adminAPI.getAllImages(1, 500)
    if (response.data?.status && Array.isArray(response.data.data)) {
      images.value = response.data.data.map((img: any) => {
        const perm = img.permission ?? img.Permission ?? 0
        const bytes = Number(img.size_bytes ?? img.SizeBytes ?? Math.round((img.size || 0) * 1024 * 1024)) || 0
        const rawUrl = img.links?.url || ''
        return {
          id: img.id,
          name: img.origin_name || img.pathname || '',
          pathname: String(img.pathname || ''),
          url: fixImageUrl(rawUrl),
          displaySrc: resolveDisplaySrc(rawUrl),
          size: formatSize(bytes),
          sizeBytes: bytes,
          uploadDate: formatDate(img.created_at),
          createdAt: img.created_at ? String(img.created_at) : '',
          isPublic: perm === 0,
          strategyId: Number(img.strategy_id ?? img.strategyId ?? 0),
          strategyName: String(img.strategy_name ?? img.strategyName ?? ''),
          mimetype: String(img.mimetype || '').toLowerCase(),
        }
      })
    } else {
      images.value = []
    }
  } catch (e) {
    console.error('加载图片列表失败:', e)
    images.value = []
  } finally {
    imageLoading.value = false
  }
}

const strategyOptions = computed(() => {
  const map = new Map<number, string>()
  for (const img of images.value) {
    if (!img.strategyId) continue
    if (!map.has(img.strategyId)) {
      map.set(img.strategyId, img.strategyName || `策略 #${img.strategyId}`)
    }
  }
  return Array.from(map.entries())
    .map(([id, name]) => ({ id, name }))
    .sort((a, b) => a.id - b.id)
})

const formatSize = (bytes: number) => {
  if (!bytes || bytes < 0) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

const onImgError = (e: Event, image: AdminImage) => {
  const el = e.target as HTMLImageElement
  if (el.dataset.fallback === '1') {
    el.src = IMG_PLACEHOLDER
    return
  }
  el.dataset.fallback = '1'
  const path = (image.pathname || '').replace(/^static\//, '')
  if (path) {
    el.src = resolveDisplaySrc(`/static/${path}`)
  } else {
    el.src = IMG_PLACEHOLDER
  }
}

const filteredImages = computed(() => {
  let filtered = [...images.value]

  if (selectedImageFilter.value === 'public') {
    filtered = filtered.filter((img) => img.isPublic)
  } else if (selectedImageFilter.value === 'private') {
    filtered = filtered.filter((img) => !img.isPublic)
  }

  if (selectedStrategyId.value !== 'all') {
    const sid = Number(selectedStrategyId.value)
    filtered = filtered.filter((img) => img.strategyId === sid)
  }

  if (selectedMimeFilter.value !== 'all') {
    if (selectedMimeFilter.value === 'other') {
      filtered = filtered.filter((img) => !KNOWN_MIMES.includes(img.mimetype))
    } else {
      filtered = filtered.filter((img) => img.mimetype === selectedMimeFilter.value)
    }
  }

  if (imageSearchKeyword.value.trim()) {
    const q = imageSearchKeyword.value.trim().toLowerCase()
    filtered = filtered.filter(
      (img) =>
        img.name.toLowerCase().includes(q) ||
        img.pathname.toLowerCase().includes(q) ||
        img.strategyName.toLowerCase().includes(q)
    )
  }

  const byTime = (a: AdminImage, b: AdminImage) =>
    new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  const bySize = (a: AdminImage, b: AdminImage) => a.sizeBytes - b.sizeBytes

  switch (sortOrder.value) {
    case 'created_desc':
      filtered.sort((a, b) => -byTime(a, b))
      break
    case 'created_asc':
      filtered.sort(byTime)
      break
    case 'size_desc':
      filtered.sort((a, b) => -bySize(a, b))
      break
    case 'size_asc':
      filtered.sort(bySize)
      break
  }

  return filtered
})

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredImages.value.length / pageSize.value))
)

const paginatedImages = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredImages.value.slice(start, start + pageSize.value)
})

const handleImageSearch = () => {
  currentPage.value = 1
}

watch([selectedImageFilter, selectedStrategyId, selectedMimeFilter, sortOrder], () => {
  currentPage.value = 1
})

watch(
  () => filteredImages.value.length,
  () => {
    const tp = Math.max(1, Math.ceil(filteredImages.value.length / pageSize.value))
    if (currentPage.value > tp) currentPage.value = tp
  }
)

// 图片点击处理
const handleImageClick = (image: AdminImage) => {
  const openUrl = image.url || resolveDisplaySrc(image.pathname ? `/static/${image.pathname.replace(/^static\//, '')}` : '')
  if (openUrl) window.open(openUrl, '_blank')
}

const handleView = (image: AdminImage) => {
  handleImageClick(image)
}

const closeRename = () => {
  renameOpen.value = false
  renameTarget.value = null
  renameValue.value = ''
}

// 编辑：修改展示用原始文件名（后端 OriginName）
const handleEdit = (image: AdminImage) => {
  renameTarget.value = image
  renameValue.value = image.name
  renameOpen.value = true
}

const submitRename = async () => {
  const img = renameTarget.value
  if (!img) return
  const name = renameValue.value.trim()
  if (!name) {
    toast.error('名称不能为空')
    return
  }
  if (name === img.name) {
    closeRename()
    return
  }
  renameSaving.value = true
  try {
    const res = await imageAPI.renameImage(img.id, name)
    if (res.data?.status) {
      toast.success('已更新名称')
      closeRename()
      await loadImages()
    } else {
      toast.error((res.data as any)?.message || '重命名失败')
    }
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.response?.data?.error || e?.message || '重命名失败'
    toast.error(typeof msg === 'string' ? msg : '重命名失败')
  } finally {
    renameSaving.value = false
  }
}

// 删除图片
const handleDelete = async (image: AdminImage) => {
  const confirmed = await confirm.danger(`确定要删除图片"${image.name}"吗？此操作不可恢复。`, '删除图片')
  if (!confirmed) {
    return
  }

  try {
    const response = await imageAPI.deleteImage(image.id)
    if (response.data?.status) {
      toast.success('删除成功！')
      await loadImages()
    } else {
      toast.error(`删除失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('删除图片失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`删除失败: ${errorMsg}`)
  }
}

// 分页处理
const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

onMounted(() => {
  loadImages()
})
</script>

<style scoped>
.image-management {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.filters-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px 10px;
  margin-bottom: 14px;
  padding: 6px 10px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
}

.filter-item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.filter-item-label {
  font-size: 11px;
  color: #9ca3af;
  flex-shrink: 0;
  user-select: none;
}

.filter-select-sm {
  height: 28px;
  padding: 2px 8px;
  font-size: 12px;
  line-height: 1.2;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background: #fff;
  color: #374151;
  min-width: 72px;
  max-width: 118px;
}

.filter-select-sm:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
}

.filter-item-strategy .filter-select-strategy {
  max-width: 140px;
}

.filter-item-search {
  flex: 1 1 160px;
  min-width: 140px;
}

.search-inline {
  display: flex;
  align-items: stretch;
  flex: 1;
  min-width: 0;
  max-width: 280px;
}

.search-input-compact {
  flex: 1;
  min-width: 0;
  height: 28px;
  padding: 2px 8px;
  font-size: 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px 0 0 4px;
  border-right: none;
  background: #fff;
}

.search-input-compact:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
}

.search-btn-compact {
  flex-shrink: 0;
  width: 32px;
  height: 28px;
  padding: 0;
  border: 1px solid #d1d5db;
  border-radius: 0 4px 4px 0;
  background: #fff;
  color: #6b7280;
  cursor: pointer;
  font-size: 12px;
}

.search-btn-compact:hover {
  background: #f3f4f6;
  color: #3b82f6;
}

.filter-stats-inline {
  margin-left: auto;
  font-size: 11px;
  color: #9ca3af;
  white-space: nowrap;
}

.filter-stats-inline strong {
  color: #6b7280;
  font-weight: 600;
}

@media (max-width: 640px) {
  .filter-stats-inline {
    width: 100%;
    margin-left: 0;
  }
}

.loading-state,
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 48px 16px;
  color: #6b7280;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.filter-dropdown {
  position: relative;
  width: 150px;
}

.filter-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: #f9fafb;
  appearance: none;
}

.filter-select:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: white;
}

.dropdown-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  pointer-events: none;
}

.search-box {
  position: relative;
  width: 300px;
}

.search-input {
  width: 100%;
  padding: 10px 40px 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: #f9fafb;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: white;
}

.search-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
}

/* 用户列表样式 */
.users-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.user-card {
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  transition: all 0.2s;
  border: 1px solid #e5e7eb;
  overflow: hidden;
  min-height: 120px;
}

.user-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  border-color: #3b82f6;
}

.card-content {
  padding: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-avatar {
  width: 56px;
  height: 56px;
  background-color: #f3f4f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #9ca3af;
  flex-shrink: 0;
  border: 2px solid #e5e7eb;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 16px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-email {
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 8px;
}

.user-stats {
  display: flex;
  gap: 20px;
  margin-bottom: 12px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #6b7280;
  background-color: #f9fafb;
  padding: 4px 8px;
  border-radius: 6px;
}

.stat-item i {
  font-size: 12px;
  color: #9ca3af;
}

.user-status {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-active {
  background-color: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.status-inactive {
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.view-images-btn {
  padding: 6px 12px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.view-images-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.action-btn:hover {
  background-color: #2563eb;
}

/* 返回头部 */
.back-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 16px 20px;
  background-color: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
}

.back-btn:hover {
  background-color: #f3f4f6;
  border-color: #9ca3af;
}

.current-user {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-label {
  font-size: 14px;
  color: #6b7280;
}

.user-name {
  font-size: 16px;
  font-weight: 600;
  color: #374151;
}

/* 图片网格 */
.images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.image-card {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
}

.image-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.image-container {
  position: relative;
  height: 150px;
  overflow: hidden;
}

.image-thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.image-card:hover .image-overlay {
  opacity: 1;
}

.overlay-actions {
  display: flex;
  gap: 8px;
}

.overlay-actions .action-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: white;
  border: none;
  color: #374151;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.overlay-actions .action-btn:hover {
  background-color: #3b82f6;
  color: white;
}

.image-info {
  padding: 12px;
}

.image-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.image-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 4px;
}

.image-strategy {
  font-size: 11px;
  color: #9ca3af;
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.image-status {
  display: flex;
  justify-content: flex-end;
}

.status-public {
  background-color: #dcfce7;
  color: #166534;
}

.status-private {
  background-color: #fef2f2;
  color: #dc2626;
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.page-btn {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  background-color: white;
  border: 1px solid #d1d5db;
  color: #374151;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.page-btn:disabled {
  background-color: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
}

.page-info {
  font-size: 14px;
  color: #6b7280;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .users-grid {
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 18px;
  }
  
  .card-content {
    padding: 18px;
    gap: 14px;
  }
  
  .user-avatar {
    width: 52px;
    height: 52px;
    font-size: 22px;
  }
  
  .user-name {
    font-size: 15px;
  }
  
  .view-images-btn {
    padding: 5px 9px;
    font-size: 10px;
  }
}

@media (max-width: 768px) {
  .action-bar {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .search-box {
    width: 100%;
  }
  
  .users-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .card-content {
    flex-direction: column;
    text-align: center;
    gap: 14px;
    padding: 18px;
  }
  
  .user-status {
    justify-content: center;
  }
  
  .user-stats {
    justify-content: center;
    gap: 16px;
  }
  
  .user-status {
    justify-content: center;
  }
  
  .back-header {
    flex-direction: column;
    gap: 12px;
    text-align: center;
  }
  
  .images-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }
  
  .image-container {
    height: 120px;
  }
}

.rename-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.rename-dialog {
  background: #fff;
  border-radius: 12px;
  padding: 20px 22px;
  max-width: 400px;
  width: 100%;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.rename-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 8px;
}

.rename-hint {
  font-size: 12px;
  color: #6b7280;
  margin: 0 0 14px;
  line-height: 1.45;
}

.rename-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  margin-bottom: 16px;
}

.rename-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.rename-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.rename-actions .btn-cancel {
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  cursor: pointer;
  font-size: 13px;
}

.rename-actions .btn-save {
  padding: 8px 14px;
  border-radius: 8px;
  border: none;
  background: #3b82f6;
  color: #fff;
  cursor: pointer;
  font-size: 13px;
}

.rename-actions .btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style> 
