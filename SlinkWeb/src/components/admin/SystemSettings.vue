<template>
  <div class="system-settings">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="text-2xl font-bold text-neutral-800 mb-2">
        <i class="fa fa-cog mr-3 text-primary"></i>
        系统设置
      </h1>
      <p class="text-neutral-600 text-sm">
        管理系统基本配置和功能开关
      </p>
    </div>

    <!-- 设置卡片 -->
    <div v-if="loading" class="loading-state">
      <i class="fa fa-spinner fa-spin"></i>
      <span>加载中...</span>
    </div>
    <div v-else class="settings-card">
      <div class="settings-section">
        <h2 class="section-title">
          <i class="fa fa-shield mr-2 text-primary"></i>
          基本设置
        </h2>
        
        <!-- 设置项列表 -->
        <div class="settings-list">
          <!-- 画廊功能 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">是否启用画廊</div>
              <div class="setting-description">
                启用或关闭画廊功能，画廊只有已登录的用户可见，画廊中的图片均为所有用户公开的图片。
              </div>
            </div>
            <div class="setting-control">
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="settings.enableGallery"
                  @change="handleSettingChange('enableGallery', $event)"
                >
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <!-- 接口功能 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">是否启用接口</div>
              <div class="setting-description">
                启用或关闭接口功能，关闭后将无法通过接口上传图片、管理图片等操作。
              </div>
            </div>
            <div class="setting-control">
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="settings.enableApi"
                  @change="handleSettingChange('enableApi', $event)"
                >
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <!-- 游客上传 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">是否允许游客上传</div>
              <div class="setting-description">
                启用或关闭游客上传功能；限额与允许类型等在「上传策略」页面配置。
              </div>
            </div>
            <div class="setting-control">
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="settings.allowGuestUpload"
                  @change="handleSettingChange('allowGuestUpload', $event)"
                >
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <!-- 账号验证 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">账号验证</div>
              <div class="setting-description">
                是否强制用户验证邮箱，开启后用户必须经过验证邮箱后才能上传图片，请确保邮件配置正常。
              </div>
            </div>
            <div class="setting-control">
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="settings.requireEmailVerification"
                  @change="handleSettingChange('requireEmailVerification', $event)"
                >
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>
        </div>
      </div>

      <!-- 应用信息设置 -->
      <div class="settings-section">
        <h2 class="section-title">
          <i class="fa fa-info-circle mr-2 text-primary"></i>
          应用信息
        </h2>

        <div class="settings-list">
          <!-- 应用名称 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">应用程序名字</div>
              <div class="setting-description">
                显示在界面和标题中的应用名称
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.appName"
                @input="handleSettingChange('appName', $event)"
                class="text-input"
                placeholder="输入应用名称"
              >
            </div>
          </div>

          <!-- 应用版本 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">应用程序版本</div>
              <div class="setting-description">
                当前应用程序的版本号
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.appVersion"
                @input="handleSettingChange('appVersion', $event)"
                class="text-input"
                placeholder="1.0.0"
              >
            </div>
          </div>

          <!-- ICP许可证号 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">ICP许可证号</div>
              <div class="setting-description">
                网站的ICP备案号
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.icpNo"
                @input="handleSettingChange('icpNo', $event)"
                class="text-input"
                placeholder="输入ICP号"
              >
            </div>
          </div>
        </div>
      </div>

      <!-- 站点信息设置 -->
      <div class="settings-section">
        <h2 class="section-title">
          <i class="fa fa-globe mr-2 text-primary"></i>
          站点信息
        </h2>

        <div class="settings-list">
          <!-- 站点描述 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">站点信息</div>
              <div class="setting-description">
                网站的描述信息，用于SEO优化
              </div>
            </div>
            <div class="setting-control">
              <textarea
                v-model="settings.siteDescription"
                @input="handleSettingChange('siteDescription', $event)"
                class="textarea-input"
                placeholder="输入站点描述"
                rows="2"
              ></textarea>
            </div>
          </div>

          <!-- SEO关键词 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">SEO关键词</div>
              <div class="setting-description">
                网站的关键词，用逗号分隔，用于SEO优化
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.siteKeywords"
                @input="handleSettingChange('siteKeywords', $event)"
                class="text-input"
                placeholder="关键词1,关键词2,关键词3"
              >
            </div>
          </div>

          <!-- 站点通知 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">站点通知</div>
              <div class="setting-description">
                显示在网站首页的公告信息
              </div>
            </div>
            <div class="setting-control">
              <textarea
                v-model="settings.siteNotice"
                @input="handleSettingChange('siteNotice', $event)"
                class="textarea-input"
                placeholder="输入站点通知"
                rows="2"
              ></textarea>
            </div>
          </div>
        </div>
      </div>

      <!-- 用户设置 -->
      <div class="settings-section">
        <h2 class="section-title">
          <i class="fa fa-users mr-2 text-primary"></i>
          用户设置
        </h2>

        <div class="settings-list">
          <!-- 默认存储空间 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">默认存储空间(bytes)</div>
              <div class="setting-description">
                新注册用户默认分配的存储空间大小（字节）
              </div>
            </div>
            <div class="setting-control">
              <input
                type="number"
                v-model="settings.userInitialCapacity"
                @input="handleSettingChange('userInitialCapacity', $event)"
                class="text-input"
                placeholder="5120000"
                min="0"
              >
            </div>
          </div>
        </div>
      </div>

      <!-- 邮箱服务配置 -->
      <div class="settings-section">
        <h2 class="section-title">
          <i class="fa fa-envelope mr-2 text-primary"></i>
          邮箱服务配置
        </h2>

        <div class="mail-config-section">
          <!-- SMTP主机 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">SMTP主机</div>
              <div class="setting-description">
                SMTP服务器地址
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.mailConfig.mailers.smtp.host"
                @input="handleSettingChange('mailHost', $event)"
                class="text-input"
                placeholder="smtp.example.com"
              >
            </div>
          </div>

          <!-- SMTP端口 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">SMTP端口</div>
              <div class="setting-description">
                SMTP服务器端口
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.mailConfig.mailers.smtp.port"
                @input="handleSettingChange('mailPort', $event)"
                class="text-input"
                placeholder="587"
              >
            </div>
          </div>

          <!-- SMTP用户名 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">SMTP用户名</div>
              <div class="setting-description">
                SMTP认证用户名
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.mailConfig.mailers.smtp.username"
                @input="handleSettingChange('mailUsername', $event)"
                class="text-input"
                placeholder="your-email@example.com"
              >
            </div>
          </div>

          <!-- SMTP密码 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">SMTP密码</div>
              <div class="setting-description">
                SMTP认证密码
              </div>
            </div>
            <div class="setting-control">
              <input
                type="password"
                v-model="settings.mailConfig.mailers.smtp.password"
                @input="handleSettingChange('mailPassword', $event)"
                class="text-input"
                placeholder="输入密码"
              >
            </div>
          </div>

          <!-- 加密方式 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">加密方式</div>
              <div class="setting-description">
                SMTP连接的加密方式
              </div>
            </div>
            <div class="setting-control">
              <select
                v-model="settings.mailConfig.mailers.smtp.encryption"
                @change="handleSettingChange('mailEncryption', $event)"
                class="select-input"
              >
                <option value="tls">TLS</option>
                <option value="ssl">SSL</option>
                <option value="none">无</option>
              </select>
            </div>
          </div>

          <!-- 发件人地址 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">发件人地址</div>
              <div class="setting-description">
                发送邮件的发件人邮箱地址
              </div>
            </div>
            <div class="setting-control">
              <input
                type="email"
                v-model="settings.mailConfig.from.address"
                @input="handleSettingChange('mailFromAddress', $event)"
                class="text-input"
                placeholder="noreply@example.com"
              >
            </div>
          </div>

          <!-- 发件人名称 -->
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">发件人名称</div>
              <div class="setting-description">
                发送邮件的发件人显示名称
              </div>
            </div>
            <div class="setting-control">
              <input
                type="text"
                v-model="settings.mailConfig.from.name"
                @input="handleSettingChange('mailFromName', $event)"
                class="text-input"
                placeholder="Slink Pro"
              >
            </div>
          </div>
        </div>
      </div>

      <!-- 数据备份与恢复 -->
      <div class="settings-section">
        <h2 class="section-title">
          <i class="fa fa-database mr-2 text-primary"></i>
          数据备份与恢复
        </h2>
        <p class="text-neutral-600 text-sm mb-4">
          将 <code class="text-xs bg-neutral-100 px-1 rounded">config</code> 目录、SQLite 数据文件
          与 <code class="text-xs bg-neutral-100 px-1 rounded">static</code> 本地图片打成 zip 下载；恢复时会覆盖本机对应内容。
        </p>
        <div class="settings-list">
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">导出全量备份</div>
              <div class="setting-description">
                非 SQLite 部署时仅包含配置文件与 static，数据库需自行用外部工具迁移。
              </div>
            </div>
            <div class="setting-control">
              <button
                type="button"
                class="save-button"
                :disabled="backupExporting"
                @click="downloadFullBackup"
              >
                <i v-if="backupExporting" class="fa fa-spinner fa-spin mr-2" />
                <i v-else class="fa fa-download mr-2" />
                {{ backupExporting ? '打包中…' : '下载 zip' }}
              </button>
            </div>
          </div>
          <div class="setting-item">
            <div class="setting-info">
              <div class="setting-title">从 zip 恢复</div>
              <div class="setting-description">
                将覆盖 config、当前 SQLite 库文件和 static 目录。请先确认本机与备份包中
                <code class="text-xs">config/database.json</code> 里数据库路径在目标机可用。
              </div>
            </div>
            <div class="setting-control flex flex-col items-end gap-2">
              <input
                ref="importInputRef"
                type="file"
                accept=".zip,application/zip"
                class="text-sm"
                @change="onImportFileChange"
              />
              <button
                type="button"
                class="reset-button"
                :disabled="backupImporting || !importFile"
                @click="runFullImport"
              >
                <i v-if="backupImporting" class="fa fa-spinner fa-spin mr-2" />
                确认恢复（覆盖本机数据）
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 保存按钮 -->
      <div class="settings-actions">
        <button
          @click="saveSettings"
          class="save-button"
          :disabled="!hasChanges || saving"
        >
          <i v-if="saving" class="fa fa-spinner fa-spin mr-2"></i>
          <i v-else class="fa fa-save mr-2"></i>
          {{ saving ? '保存中...' : '保存更改' }}
        </button>
        <button
          @click="resetSettings"
          class="reset-button"
          :disabled="!hasChanges || saving"
        >
          <i class="fa fa-undo mr-2"></i>
          重置
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { adminAPI } from '@/api'
import { useMessage } from '@/composables/useMessage.ts'
import { isAxiosError } from 'axios'

// 设置数据
const originalSettings = reactive({
  // 功能开关设置
  enableGallery: false,
  enableApi: true,
  allowGuestUpload: false,
  requireEmailVerification: false,
  // 应用信息设置
  appName: 'Slink Pro',
  appVersion: '1.0.0',
  icpNo: '5',
  // 站点信息设置
  siteDescription: 'Slink Pro, Your photo album on the cloud.',
  siteKeywords: 'Slink Pro,slink,丝灵图床',
  siteNotice: 'Welcome to Slink Pro!',
  // 用户设置
  userInitialCapacity: '5120000.00',
  // 邮件配置
  mailConfig: {
    mailers: {
      smtp: {
        host: 'smtp.mailgun.org',
        port: '587',
        username: null,
        password: null,
        encryption: 'tls',
        timeout: '10',
        transport: 'smtp'
      }
    },
    from: {
      address: null,
      name: null
    },
    default: 'smtp'
  }
})

const settings = reactive({
  // 功能开关设置
  enableGallery: false,
  enableApi: true,
  allowGuestUpload: false,
  requireEmailVerification: false,
  // 应用信息设置
  appName: 'Slink Pro',
  appVersion: '1.0.0',
  icpNo: '5',
  // 站点信息设置
  siteDescription: 'Slink Pro, Your photo album on the cloud.',
  siteKeywords: 'Slink Pro,slink,丝灵图床',
  siteNotice: 'Welcome to Slink Pro!',
  // 用户设置
  userInitialCapacity: '5120000.00',
  // 邮件配置
  mailConfig: {
    mailers: {
      smtp: {
        host: 'smtp.mailgun.org',
        port: '587',
        username: null,
        password: null,
        encryption: 'tls',
        timeout: '10',
        transport: 'smtp'
      }
    },
    from: {
      address: null,
      name: null
    },
    default: 'smtp'
  }
})

const loading = ref(false)
const saving = ref(false)
const backupExporting = ref(false)
const backupImporting = ref(false)
const importFile = ref<File | null>(null)
const importInputRef = ref<HTMLInputElement | null>(null)

const { toast } = useMessage()

// 计算是否有更改
const hasChanges = computed(() => {
  return JSON.stringify(settings) !== JSON.stringify(originalSettings)
})

// 设置变更处理
const handleSettingChange = (key: string, event: Event) => {
  const target = event.target as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
  let value: any = target.value

  // 对于复选框，使用checked属性
  if (target.type === 'checkbox') {
    value = (target as HTMLInputElement).checked
  }

  console.log(`设置 ${key} 已更改为: ${value}`)
}

// 保存设置
const saveSettings = async () => {
  if (saving.value) return

  saving.value = true
  try {
    // 构建请求数据，映射前端字段到后端字段
    const requestData = {
      // 功能开关设置（个人图床固定关闭注册）
      enable_register: false,
      enable_gallery: settings.enableGallery,
      enable_api: settings.enableApi,
      guest_upload: settings.allowGuestUpload,
      email_verify: settings.requireEmailVerification,
      // 应用信息设置
      app_name: settings.appName,
      app_version: settings.appVersion,
      icp_no: settings.icpNo,
      // 站点信息设置
      site_description: settings.siteDescription,
      site_keywords: settings.siteKeywords,
      site_notice: settings.siteNotice,
      // 用户设置
      user_initial_capacity: settings.userInitialCapacity,
      // 邮件配置
      mail: JSON.stringify(settings.mailConfig)
    }

    console.log('保存设置数据:', requestData)

    const response = await adminAPI.updateSystemConfigs(requestData)
    console.log('保存设置响应:', response)

    if (response.data && response.data.status) {
      // 更新原始设置
      Object.assign(originalSettings, settings)
      toast.success('设置已保存成功！')
    } else {
      toast.error(`保存失败: ${response.data?.message || '未知错误'}`)
    }
  } catch (error: any) {
    console.error('保存设置失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    toast.error(`保存失败: ${errorMsg}`)
  } finally {
    saving.value = false
  }
}

// 重置设置
const resetSettings = () => {
  // 重置为原始设置
  Object.assign(settings, originalSettings)
  console.log('设置已重置')
}

const onImportFileChange = (e: Event) => {
  const t = e.target as HTMLInputElement
  importFile.value = t.files && t.files[0] ? t.files[0] : null
}

const downloadFullBackup = async () => {
  if (backupExporting.value) return
  backupExporting.value = true
  try {
    const res = await adminAPI.exportFullBackup()
    const blob = res.data
    if (!(blob instanceof Blob) || blob.size < 10) {
      toast.error('下载失败：响应不是有效的压缩包')
      return
    }
    const raw = new Date().toISOString().slice(0, 19)
    const name = `slink-backup-${raw.replace(/-/g, '').replace(/:/g, '').replace('T', '')}.zip`
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = name
    a.click()
    URL.revokeObjectURL(url)
    toast.success('备份已开始下载')
  } catch (e: unknown) {
    if (isAxiosError(e) && e.response?.data instanceof Blob) {
      try {
        const t = await (e.response.data as Blob).text()
        const j = JSON.parse(t) as { message?: string }
        toast.error(j.message || '导出失败')
      } catch {
        toast.error('导出失败')
      }
    } else {
      toast.error('导出失败')
    }
  } finally {
    backupExporting.value = false
  }
}

const runFullImport = async () => {
  if (!importFile.value || backupImporting.value) return
  if (!window.confirm('将覆盖本机 config、数据库与 static，是否继续？')) {
    return
  }
  backupImporting.value = true
  try {
    const res = await adminAPI.importFullBackup(importFile.value)
    const d = res.data
    if (d?.status === true) {
      toast.success(d.message || '恢复成功')
      importFile.value = null
      if (importInputRef.value) importInputRef.value.value = ''
    } else {
      toast.error(d?.message || '恢复失败')
    }
  } catch (e: unknown) {
    const msg = isAxiosError(e) ? (e.response?.data as { message?: string })?.message : undefined
    toast.error(msg || (e as Error).message || '恢复失败')
  } finally {
    backupImporting.value = false
  }
}

// 加载设置
const loadSettings = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getSystemConfigs()
    console.log('系统配置API响应:', response)

    if (response.data && response.data.status) {
      const data = response.data.data
      console.log('系统配置数据:', data)

      // 更新设置，映射后端字段到前端字段
      settings.enableGallery = data.enable_gallery || false
      settings.enableApi = data.enable_api || false
      settings.allowGuestUpload = data.guest_upload || false
      settings.requireEmailVerification = data.email_verify || false

      // 应用信息设置
      settings.appName = data.app_name || 'Slink Pro'
      settings.appVersion = data.app_version || '1.0.0'
      settings.icpNo = data.icp_no || '5'

      // 站点信息设置
      settings.siteDescription = data.site_description || 'Slink Pro, Your photo album on the cloud.'
      settings.siteKeywords = data.site_keywords || 'Slink Pro,slink,丝灵图床'
      settings.siteNotice = data.site_notice || 'Welcome to Slink Pro!'

      // 用户设置
      settings.userInitialCapacity = data.user_initial_capacity || '5120000.00'

      // 邮件配置
      if (data.mail) {
        try {
          const mailConfig = typeof data.mail === 'string' ? JSON.parse(data.mail) : data.mail
          settings.mailConfig = {
            mailers: {
              smtp: {
                host: mailConfig.mailers?.smtp?.host || 'smtp.mailgun.org',
                port: mailConfig.mailers?.smtp?.port || '587',
                username: mailConfig.mailers?.smtp?.username || null,
                password: mailConfig.mailers?.smtp?.password || null,
                encryption: mailConfig.mailers?.smtp?.encryption || 'tls',
                timeout: mailConfig.mailers?.smtp?.timeout || '10',
                transport: mailConfig.mailers?.smtp?.transport || 'smtp'
              }
            },
            from: {
              address: mailConfig.from?.address || null,
              name: mailConfig.from?.name || null
            },
            default: mailConfig.default || 'smtp'
          }
        } catch (error) {
          console.error('解析邮件配置失败:', error)
          // 使用默认配置
        }
      }

      // 保存原始设置
      Object.assign(originalSettings, settings)
    } else {
      console.error('加载设置失败:', response.data?.message)
    }
  } catch (error: any) {
    console.error('加载设置失败:', error)
    const errorMsg = error.response?.data?.error || error.message || '未知错误'
    console.error('错误详情:', errorMsg)
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(async () => {
  console.log('系统设置组件已加载')
  await loadSettings()
})
</script>

<style scoped>
.system-settings {
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
}

.settings-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.settings-section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 24px;
  display: flex;
  align-items: center;
}

.settings-list {
  margin: 0;
  padding: 0;
}

.setting-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 20px 0;
  border-bottom: 1px solid #f3f4f6;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  flex: 1;
  margin-right: 24px;
}

.setting-title {
  font-size: 16px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 4px;
}

.setting-description {
  font-size: 14px;
  color: #6b7280;
  line-height: 1.5;
}

.setting-control {
  flex-shrink: 0;
}

/* 开关样式 */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #d1d5db;
  transition: 0.3s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: #3b82f6;
}

input:checked + .toggle-slider:before {
  transform: translateX(26px);
}

/* 输入框样式 */
.text-input, .textarea-input, .select-input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  width: 100%;
  max-width: 300px;
}

.text-input:focus, .textarea-input:focus, .select-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.textarea-input {
  resize: vertical;
  min-height: 60px;
}

.select-input {
  background-color: #fff;
  cursor: pointer;
}

/* 按钮样式 */
.settings-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding-top: 24px;
  border-top: 1px solid #f3f4f6;
}

.save-button, .reset-button {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
}

.save-button {
  background-color: #3b82f6;
  color: white;
  border: none;
}

.save-button:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.save-button:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
}

.reset-button {
  background-color: #f3f4f6;
  color: #6b7280;
  border: 1px solid #d1d5db;
}

.reset-button:hover:not(:disabled) {
  background-color: #e5e7eb;
  transform: translateY(-1px);
}

.reset-button:disabled {
  background-color: #f9fafb;
  color: #d1d5db;
  cursor: not-allowed;
  transform: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .system-settings {
    padding: 16px;
  }
  
  .settings-card {
    padding: 16px;
  }
  
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .setting-info {
    margin-right: 0;
    margin-bottom: 12px;
  }
  
  .settings-actions {
    flex-direction: column;
  }
  
  .save-button, .reset-button {
    width: 100%;
    justify-content: center;
  }
}

/* 动画效果 */
.setting-item {
  transition: background-color 0.2s ease;
}

.setting-item:hover {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 20px 12px;
  margin: 0 -12px;
}

/* 加载动画 */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.settings-card {
  animation: fadeInUp 0.5s ease-out;
}
</style> 
