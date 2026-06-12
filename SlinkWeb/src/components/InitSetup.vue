<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
    <div class="max-w-2xl w-full bg-white rounded-2xl shadow-2xl p-8">
      <!-- 头部 -->
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-gray-800 mb-2">个人图床初始化</h1>
        <p class="text-gray-600">首次运行请创建站长账号并配置数据库（本系统为单用户，不提供公开注册）</p>
      </div>

      <!-- 步骤指示器 -->
      <div class="flex justify-between mb-8">
        <div v-for="(step, index) in steps" :key="index" class="flex-1">
          <div class="flex items-center">
            <div
              :class="[
                'w-10 h-10 rounded-full flex items-center justify-center font-semibold',
                currentStep >= index + 1
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-200 text-gray-500'
              ]"
            >
              {{ index + 1 }}
            </div>
            <div
              v-if="index < steps.length - 1"
              :class="[
                'flex-1 h-1 mx-2',
                currentStep > index + 1 ? 'bg-indigo-600' : 'bg-gray-200'
              ]"
            ></div>
          </div>
          <div class="text-xs mt-2 text-center">{{ step }}</div>
        </div>
      </div>

      <!-- 表单内容 -->
      <form @submit.prevent="handleSubmit">
        <!-- 步骤1: 管理员账号 -->
        <div v-show="currentStep === 1" class="space-y-4">
          <h2 class="text-2xl font-semibold text-gray-800 mb-4">站长账号（唯一）</h2>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">登录账号</label>
            <input
              v-model="formData.adminAccount"
              type="text"
              required
              autocomplete="username"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="任意账号，如 admin"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">登录密码</label>
            <input
              v-model="formData.adminPassword"
              type="password"
              required
              minlength="6"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="至少6位字符"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">确认密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              required
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="再次输入密码"
            />
            <p v-if="passwordMismatch" class="text-red-500 text-sm mt-1">密码不匹配</p>
          </div>
        </div>

        <!-- 步骤2: 数据库配置 -->
        <div v-show="currentStep === 2" class="space-y-4">
          <h2 class="text-2xl font-semibold text-gray-800 mb-4">数据库配置</h2>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">数据库类型</label>
            <select
              v-model="formData.databaseConfig.type"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            >
              <option value="sqlite">SQLite (推荐)</option>
              <option value="mysql">MySQL</option>
              <option value="postgres">PostgreSQL</option>
              <option value="sqlserver">SQL Server</option>
            </select>
          </div>

          <!-- SQLite 配置 -->
          <div v-if="formData.databaseConfig.type === 'sqlite'">
            <label class="block text-sm font-medium text-gray-700 mb-2">数据库文件路径</label>
            <input
              v-model="formData.databaseConfig.dbname"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="data.db"
            />
          </div>

          <!-- MySQL/PostgreSQL/SQL Server 配置 -->
          <template v-else>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">主机地址</label>
                <input
                  v-model="formData.databaseConfig.host"
                  type="text"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  placeholder="localhost"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">端口</label>
                <input
                  :value="formData.databaseConfig.port != null ? String(formData.databaseConfig.port) : ''"
                  @input="onPortInput($event as InputEvent)"
                  type="text"
                  pattern="\d*"
                  inputmode="numeric"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="String(getDefaultPort())"
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">数据库名</label>
              <input
                v-model="formData.databaseConfig.dbname"
                type="text"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="slink"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">用户名</label>
                <input
                  v-model="formData.databaseConfig.user"
                  type="text"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  placeholder="root"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">密码</label>
                <input
                  v-model="formData.databaseConfig.password"
                  type="password"
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  placeholder="数据库密码"
                />
              </div>
            </div>

            <div v-if="formData.databaseConfig.type === 'postgres'">
              <label class="block text-sm font-medium text-gray-700 mb-2">SSL 模式</label>
              <select
                v-model="formData.databaseConfig.sslmode"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option value="disable">禁用</option>
                <option value="require">需要</option>
                <option value="verify-ca">验证CA</option>
                <option value="verify-full">完全验证</option>
              </select>
            </div>

            <div v-if="formData.databaseConfig.type === 'mysql'">
              <label class="block text-sm font-medium text-gray-700 mb-2">字符集</label>
              <input
                v-model="formData.databaseConfig.charset"
                type="text"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="utf8mb4"
              />
            </div>
          </template>

          <button
            type="button"
            @click="testDatabaseConnection"
            :disabled="testing"
            class="w-full py-2 px-4 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-gray-400 transition-colors"
          >
            {{ testing ? '测试中...' : '测试数据库连接' }}
          </button>
          <p v-if="testResult" :class="testResult.success ? 'text-green-600' : 'text-red-600'" class="text-sm">
            {{ testResult.message }}
          </p>
        </div>

        <!-- 步骤3: 缓存配置 -->
        <div v-show="currentStep === 3" class="space-y-4">
          <h2 class="text-2xl font-semibold text-gray-800 mb-4">缓存配置</h2>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">缓存类型</label>
            <select
              v-model="formData.cacheType"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            >
              <option value="memory">内存缓存 (推荐)</option>
              <option value="file">文件缓存</option>
            </select>
          </div>

          <div v-if="formData.cacheType === 'file'">
            <label class="block text-sm font-medium text-gray-700 mb-2">缓存目录</label>
            <input
              v-model="formData.cachePath"
              type="text"
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="cache"
            />
          </div>

          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h3 class="font-semibold text-blue-800 mb-2">配置摘要</h3>
            <ul class="text-sm text-blue-700 space-y-1">
              <li>登录账号: {{ formData.adminAccount }}</li>
              <li>数据库类型: {{ getDatabaseTypeName() }}</li>
              <li>缓存类型: {{ formData.cacheType === 'memory' ? '内存缓存' : '文件缓存' }}</li>
            </ul>
          </div>
        </div>

        <!-- 按钮组 -->
        <div class="flex justify-between mt-8">
          <button
            v-if="currentStep > 1"
            type="button"
            @click="previousStep"
            class="px-6 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
          >
            上一步
          </button>
          <div v-else></div>

          <button
            v-if="currentStep < 3"
            type="button"
            @click="nextStep"
            :disabled="!canProceed"
            class="px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:bg-gray-400 transition-colors"
          >
            下一步
          </button>
          <button
            v-else
            type="submit"
            :disabled="submitting"
            class="px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:bg-gray-400 transition-colors"
          >
            {{ submitting ? '初始化中...' : '完成初始化' }}
          </button>
        </div>
      </form>

      <!-- 错误提示 -->
      <div v-if="error" class="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
        <p class="text-red-700">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '../utils/request'
import { useMessage } from '../composables/useMessage'

const router = useRouter()

const steps = ['管理员账号', '数据库配置', '缓存配置']
const currentStep = ref(1)
const submitting = ref(false)
const testing = ref(false)
const error = ref('')
const confirmPassword = ref('')
const testResult = ref<{ success: boolean; message: string } | null>(null)

const { toast } = useMessage()

const formData = ref({
  adminAccount: '',
  adminPassword: '',
  databaseConfig: {
    type: 'sqlite',
    host: 'localhost',
    port: 3306,
    user: '',
    password: '',
    dbname: 'data.db',
    sslmode: 'disable',
    charset: 'utf8mb4'
  },
  cacheType: 'memory',
  cachePath: 'cache'
})

const passwordMismatch = computed(() => {
  return confirmPassword.value && formData.value.adminPassword !== confirmPassword.value
})

const canProceed = computed(() => {
  if (currentStep.value === 1) {
    return (
      formData.value.adminAccount &&
      formData.value.adminPassword.length >= 6 &&
      formData.value.adminPassword === confirmPassword.value
    )
  }
  if (currentStep.value === 2) {
    if (formData.value.databaseConfig.type === 'sqlite') {
      return formData.value.databaseConfig.dbname
    }
    return (
      formData.value.databaseConfig.host &&
      formData.value.databaseConfig.port &&
      formData.value.databaseConfig.dbname &&
      formData.value.databaseConfig.user
    )
  }
  return true
})

const getDefaultPort = () => {
  const ports: Record<string, number> = {
    mysql: 3306,
    postgres: 5432,
    sqlserver: 1433
  }
  return ports[formData.value.databaseConfig.type] || 3306
}

const getDatabaseTypeName = () => {
  const names: Record<string, string> = {
    sqlite: 'SQLite',
    mysql: 'MySQL',
    postgres: 'PostgreSQL',
    sqlserver: 'SQL Server'
  }
  return names[formData.value.databaseConfig.type] || 'SQLite'
}

const nextStep = () => {
  if (canProceed.value && currentStep.value < 3) {
    currentStep.value++
    error.value = ''
  }
}

const previousStep = () => {
  if (currentStep.value > 1) {
    currentStep.value--
    error.value = ''
  }
}

const testDatabaseConnection = async () => {
  testing.value = true
  testResult.value = null
  error.value = ''

  try {
    const response = await http.post('/api/init/test-db', formData.value.databaseConfig)
    if (response.data.code === 200) {
      testResult.value = {
        success: true,
        message: '数据库连接成功！'
      }
    } else {
      testResult.value = {
        success: false,
        message: response.data.msg || '数据库连接失败'
      }
    }
  } catch (err: any) {
    testResult.value = {
      success: false,
      message: err.response?.data?.msg || '数据库连接失败'
    }
  } finally {
    testing.value = false
  }
}

const handleSubmit = async () => {
  if (!canProceed.value) return

  submitting.value = true
  error.value = ''

  try {
    const response = await http.post('/api/init/setup', {
      admin_account: formData.value.adminAccount,
      admin_password: formData.value.adminPassword,
      database_config: formData.value.databaseConfig,
      cache_type: formData.value.cacheType,
      cache_path: formData.value.cachePath
    })

    if (response.data.code === 200) {
      toast.success('初始化成功！请使用管理员账号登录。')
      // 重新加载页面，让后端重新检查初始化状态
      window.location.href = '/'
    } else {
      error.value = response.data.msg || '初始化失败'
    }
  } catch (err: any) {
    error.value = err.response?.data?.msg || '初始化失败，请检查配置'
  } finally {
    submitting.value = false
  }
}

const onPortInput = (event: InputEvent) => {
  const target = event.target as HTMLInputElement
  const value = target.value.replace(/[^0-9]/g, '')
  target.value = value
  formData.value.databaseConfig.port = value ? Number(value) : getDefaultPort()
}
</script>
