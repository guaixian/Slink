<template>
  <PageLayout
    title="AI 设置"
    description="配置 AI 图像扩展、LLM 平台与向量 Embedding 平台的地址与密钥，并测试连接。"
    icon-class="fa fa-flask"
  >
    <div v-if="loading" class="loading-box">
      <i class="fa fa-spinner fa-spin mr-2"></i>加载中…
    </div>

    <template v-else>
      <p class="global-note">
        <i class="fa fa-info-circle mr-1"></i>{{ data?.note }}
      </p>

      <!-- 图像扩展 -->
      <section class="card">
        <h3 class="card-title"><i class="fa fa-magic mr-2"></i>图像处理外置扩展</h3>
        <p class="card-sub">去水印 / 超分辨率 / 智能打标，对接外部图像模型服务（契约见 docs/ai-extension.md）。</p>

        <div class="field">
          <label>服务地址 (Endpoint)</label>
          <input v-model="form.ai_image_endpoint" type="text" class="inp"
            :disabled="locked('image','endpoint')" placeholder="http://127.0.0.1:9000" />
          <span v-if="locked('image','endpoint')" class="lock-tip"><i class="fa fa-lock"></i> 由环境变量锁定</span>
        </div>
        <div class="field">
          <label>令牌 (Token，可选)</label>
          <input v-model="form.ai_image_token" type="password" class="inp"
            :disabled="locked('image','token')" :placeholder="secretPlaceholder('image','token')" autocomplete="new-password" />
          <span v-if="locked('image','token')" class="lock-tip"><i class="fa fa-lock"></i> 由环境变量锁定</span>
        </div>
        <div class="field">
          <label>超时 (秒)</label>
          <input v-model="form.ai_image_timeout" type="number" min="1" class="inp narrow"
            :disabled="locked('image','timeout')" placeholder="120" />
        </div>
        <div class="actions">
          <button class="btn-test" :disabled="testing.image" @click="test('image')">
            <i class="fa" :class="testing.image ? 'fa-spinner fa-spin' : 'fa-plug'"></i> 测试连接
          </button>
          <TestBadge :result="results.image" />
        </div>
      </section>

      <!-- LLM -->
      <section class="card">
        <h3 class="card-title"><i class="fa fa-comments mr-2"></i>LLM 平台（OpenAI 兼容）</h3>
        <p class="card-sub">用于配置与连通性测试，能力预留（如后续的图片描述 / Alt 文本生成）。</p>

        <div class="field">
          <label>Base URL</label>
          <input v-model="form.ai_llm_base_url" type="text" class="inp"
            :disabled="locked('llm','base_url')" placeholder="https://api.openai.com/v1" />
          <span v-if="locked('llm','base_url')" class="lock-tip"><i class="fa fa-lock"></i> 由环境变量锁定</span>
        </div>
        <div class="field">
          <label>API Key</label>
          <input v-model="form.ai_llm_api_key" type="password" class="inp"
            :disabled="locked('llm','api_key')" :placeholder="secretPlaceholder('llm','api_key')" autocomplete="new-password" />
          <span v-if="locked('llm','api_key')" class="lock-tip"><i class="fa fa-lock"></i> 由环境变量锁定</span>
        </div>
        <div class="field">
          <label>默认模型</label>
          <input v-model="form.ai_llm_model" type="text" class="inp"
            :disabled="locked('llm','model')" placeholder="gpt-4o-mini" />
        </div>
        <div class="actions">
          <button class="btn-test" :disabled="testing.llm" @click="test('llm')">
            <i class="fa" :class="testing.llm ? 'fa-spinner fa-spin' : 'fa-plug'"></i> 测试连接
          </button>
          <TestBadge :result="results.llm" />
        </div>
      </section>

      <!-- Embedding -->
      <section class="card">
        <h3 class="card-title"><i class="fa fa-sitemap mr-2"></i>向量 Embedding 平台（OpenAI 兼容）</h3>
        <p class="card-sub">用于配置与连通性测试，能力预留（如后续的以图搜图 / 语义检索）。</p>

        <div class="field">
          <label>Base URL</label>
          <input v-model="form.ai_embedding_base_url" type="text" class="inp"
            :disabled="locked('embedding','base_url')" placeholder="https://api.openai.com/v1" />
          <span v-if="locked('embedding','base_url')" class="lock-tip"><i class="fa fa-lock"></i> 由环境变量锁定</span>
        </div>
        <div class="field">
          <label>API Key</label>
          <input v-model="form.ai_embedding_api_key" type="password" class="inp"
            :disabled="locked('embedding','api_key')" :placeholder="secretPlaceholder('embedding','api_key')" autocomplete="new-password" />
          <span v-if="locked('embedding','api_key')" class="lock-tip"><i class="fa fa-lock"></i> 由环境变量锁定</span>
        </div>
        <div class="field">
          <label>默认模型</label>
          <input v-model="form.ai_embedding_model" type="text" class="inp"
            :disabled="locked('embedding','model')" placeholder="text-embedding-3-small" />
        </div>
        <div class="field">
          <label>向量维度</label>
          <input v-model="form.ai_embedding_dimensions" type="number" min="1" class="inp narrow"
            :disabled="locked('embedding','dimensions')" placeholder="1536" />
        </div>
        <div class="actions">
          <button class="btn-test" :disabled="testing.embedding" @click="test('embedding')">
            <i class="fa" :class="testing.embedding ? 'fa-spinner fa-spin' : 'fa-plug'"></i> 测试连接
          </button>
          <TestBadge :result="results.embedding" />
        </div>
      </section>

      <div class="save-bar">
        <button class="btn-save" :disabled="saving" @click="save">
          <i class="fa" :class="saving ? 'fa-spinner fa-spin' : 'fa-save'"></i>
          {{ saving ? '保存中…' : '保存设置' }}
        </button>
      </div>
    </template>
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import PageLayout from './PageLayout.vue'
import { adminAPI } from '../../api'
import type { AISettingsData } from '../../api'
import { useMessage } from '../../composables/useMessage'

const { toast } = useMessage()

type Section = 'image' | 'llm' | 'embedding'
type TestResult = { ok: boolean; status?: number; message: string }

const loading = ref(true)
const saving = ref(false)
const data = ref<AISettingsData | null>(null)

const testing = reactive<Record<Section, boolean>>({ image: false, llm: false, embedding: false })
const results = reactive<Record<Section, TestResult | null>>({ image: null, llm: null, embedding: null })

// 表单：非密钥字段填入当前值；密钥字段留空（占位提示是否已配置）
const form = reactive<Record<string, string>>({
  ai_image_endpoint: '', ai_image_token: '', ai_image_timeout: '',
  ai_llm_base_url: '', ai_llm_api_key: '', ai_llm_model: '',
  ai_embedding_base_url: '', ai_embedding_api_key: '', ai_embedding_model: '', ai_embedding_dimensions: '',
})

// 内联的测试结果徽标组件
const TestBadge = (props: { result: TestResult | null }) => {
  if (!props.result) return null
  const ok = props.result.ok
  return h('span', { class: ['test-badge', ok ? 'ok' : 'err'] }, [
    h('i', { class: ['fa', ok ? 'fa-check-circle' : 'fa-times-circle', 'mr-1'] }),
    props.result.message,
  ])
}

const fieldOf = (section: Section, name: string) => {
  const d = data.value as any
  return d?.[section]?.[name]
}

const locked = (section: Section, name: string) => !!fieldOf(section, name)?.from_env

// 密钥占位：已配置则提示「已配置」，否则留空提示
const secretPlaceholder = (section: Section, name: string) => {
  const f = fieldOf(section, name)
  if (f?.from_env) return f.value || '由环境变量锁定'
  return f?.has_value ? '已配置（如需修改请输入新值）' : '未配置'
}

const load = async () => {
  loading.value = true
  try {
    const res = await adminAPI.getAISettings()
    if (res.data?.status && res.data.data) {
      data.value = res.data.data
      const d = res.data.data
      // 非密钥字段回填当前值，便于直接编辑
      form.ai_image_endpoint = d.image.endpoint.from_env ? '' : d.image.endpoint.value
      form.ai_image_timeout = d.image.timeout.from_env ? '' : d.image.timeout.value
      form.ai_llm_base_url = d.llm.base_url.from_env ? '' : d.llm.base_url.value
      form.ai_llm_model = d.llm.model.from_env ? '' : d.llm.model.value
      form.ai_embedding_base_url = d.embedding.base_url.from_env ? '' : d.embedding.base_url.value
      form.ai_embedding_model = d.embedding.model.from_env ? '' : d.embedding.model.value
      form.ai_embedding_dimensions = d.embedding.dimensions.from_env ? '' : d.embedding.dimensions.value
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.message || '加载AI设置失败')
  } finally {
    loading.value = false
  }
}

const save = async () => {
  saving.value = true
  try {
    // 仅发送非锁定字段；密钥为空表示保留原值，后端处理
    const payload: Record<string, string> = {}
    for (const [k, v] of Object.entries(form)) {
      const [sec, name] = sectionName(k)
      if (sec && locked(sec, name)) continue
      payload[k] = v
    }
    const res = await adminAPI.updateAISettings(payload)
    if (res.data?.status) {
      toast.success('保存成功')
      await load() // 重新拉取以刷新脱敏展示与 has_value
      // 清空密钥输入框
      form.ai_image_token = ''
      form.ai_llm_api_key = ''
      form.ai_embedding_api_key = ''
    } else {
      toast.error(res.data?.message || '保存失败')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

// 把 config_key 拆成 section 与字段名（如 ai_image_endpoint -> image, endpoint）
const sectionName = (key: string): [Section | '', string] => {
  if (key.startsWith('ai_image_')) return ['image', key.replace('ai_image_', '')]
  if (key.startsWith('ai_llm_')) return ['llm', key.replace('ai_llm_', '')]
  if (key.startsWith('ai_embedding_')) return ['embedding', key.replace('ai_embedding_', '')]
  return ['', key]
}

const test = async (target: Section) => {
  testing[target] = true
  results[target] = null
  try {
    // 测试基于已保存配置；提示用户先保存再测
    const res = await adminAPI.testAISettings(target)
    if (res.data?.status && res.data.data) {
      results[target] = res.data.data
    } else {
      results[target] = { ok: false, message: res.data?.message || '测试失败' }
    }
  } catch (e: any) {
    results[target] = { ok: false, message: e?.response?.data?.message || '测试失败' }
  } finally {
    testing[target] = false
  }
}

onMounted(load)
</script>

<style scoped>
.loading-box {
  padding: 40px;
  text-align: center;
  color: #6b7280;
}

.global-note {
  background: #eff6ff;
  border: 1px solid #dbeafe;
  color: #1e40af;
  font-size: 13px;
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 20px;
  line-height: 1.6;
}

.card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.card-sub {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 18px;
  line-height: 1.5;
}

.field {
  margin-bottom: 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
}

.inp {
  padding: 9px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  color: #374151;
  width: 100%;
  max-width: 480px;
  transition: all 0.2s ease;
}

.inp.narrow {
  max-width: 160px;
}

.inp:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.inp:disabled {
  background: #f3f4f6;
  color: #9ca3af;
  cursor: not-allowed;
}

.lock-tip {
  font-size: 12px;
  color: #d97706;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

.btn-test {
  padding: 8px 16px;
  background: #fff;
  color: #3b82f6;
  border: 1px solid #3b82f6;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-test:hover:not(:disabled) {
  background: #eff6ff;
}

.btn-test:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.test-badge {
  font-size: 13px;
  display: inline-flex;
  align-items: center;
}

.test-badge.ok {
  color: #059669;
}

.test-badge.err {
  color: #dc2626;
}

.save-bar {
  position: sticky;
  bottom: 0;
  padding: 16px 0;
  display: flex;
  justify-content: flex-end;
}

.btn-save {
  padding: 11px 28px;
  background: #3b82f6;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
}

.btn-save:hover:not(:disabled) {
  background: #2563eb;
}

.btn-save:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}
</style>
