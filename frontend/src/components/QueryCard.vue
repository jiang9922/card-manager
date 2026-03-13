<template>
  <div class="card">
    <h3>验证码查询</h3>
    <div class="field">
      <label>手机号</label>
      <input v-model="phone" @keyup.enter="loadCards" placeholder="请输入手机号查询" />
    </div>
    <div class="field" v-if="cardList.length > 0">
      <label>选择卡号</label>
      <select v-model="selectedCard" @change="query">
        <option value="">请选择卡号</option>
        <option v-for="c in cardList" :key="c.id" :value="c.query_url?.split('?card=')[1] || c.card_no">
          {{ c.card_no }}
        </option>
      </select>
    </div>
    <div class="field" v-else-if="phone && !loading">
      <label>卡号</label>
      <input v-model="cardNo" @keyup.enter="query" placeholder="请输入卡号直接查询" />
    </div>
    <div class="field" v-else>
      <label>卡号</label>
      <input v-model="cardNo" @keyup.enter="query" placeholder="请输入卡号" />
    </div>
    <div class="field">
      <label>验证码</label>
      <div class="code-line">
        <input :value="code" readonly placeholder="点击查询获取" />
        <button v-show="code" @click="copyCode" class="btn-copy">
          {{ copyText }}
        </button>
      </div>
    </div>
    <button @click="query" :disabled="loading" class="btn-query">
      {{ loading ? '查询中...' : '立即查询' }}
    </button>

    <div v-if="result" class="result">
      <div class="item"><span>验证码</span><strong>{{ code }}</strong></div>
      <div class="item"><span>获取时间</span><span>{{ formatTime(codeTime) }}</span></div>
    </div>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
// 组件职责：
// - 输入卡号并请求后端 `/api/cards/query?card=...`
// - 展示验证码、获取时间与过期时间
// - 支持从 URL（`card_enc`/`card`/`card_no`）预填卡号，便于分享或从后台跳转
// - 提供复制验证码到剪贴板
import { ref, onMounted, watch } from 'vue'
import { useToast } from '../composables/useToast'

const toast = useToast()
// 展示数据与状态
const phone = ref('')
const cardList = ref<any[]>([])
const selectedCard = ref('')
const cardNo = ref('')
const code = ref('')
const codeTime = ref('')
const cardLink = ref('')
const result = ref(false)
const error = ref('')
const loading = ref(false)
const copyText = ref('复制')

onMounted(() => {
  // 从 URL 参数恢复
  const params = new URLSearchParams(window.location.search)
  const enc = params.get('card_enc')
  if (enc) {
    try { 
      const decoded = atob(enc)
      cardNo.value = decoded.split('_')[0] || decoded
    } catch {}
  } else {
    const plain = params.get('card') || params.get('card_no')
    if (plain) {
      cardNo.value = plain.split('_')[0] || plain
    }
  }
  // 监听手机号变化，自动加载卡密列表
  watch(phone, loadCards)
})

// 根据手机号加载卡密列表
async function loadCards() {
  if (!phone.value.trim()) {
    cardList.value = []
    return
  }
  loading.value = true
  try {
    const res = await fetch(`/api/cards?phone=${encodeURIComponent(phone.value)}&page_size=100`)
    const json = await res.json()
    if (json.code === 0 && json.data && json.data.cards) {
      cardList.value = json.data.cards
    } else {
      cardList.value = []
    }
  } catch {
    cardList.value = []
  } finally {
    loading.value = false
  }
}

async function query() {
  // 优先使用选择的卡号
  const queryCardNo = selectedCard.value || cardNo.value
  if (!queryCardNo.trim()) return toast('请输入卡号', 'error')
  
  loading.value = true
  error.value = ''
  result.value = false
  code.value = ''
  codeTime.value = ''

  try {
    const url = `/api/cards/query?card=${encodeURIComponent(queryCardNo)}`
    const res = await fetch(url)
    const json = await res.json()
    if (json.code !== 0 || !json.data) {
      code.value = ''
      codeTime.value = ''
      result.value = false
      throw new Error(json.message || '未获取到验证码')
    }
    const d = json.data
    code.value = d.card_code || ''
    if (d.card_note) {
      const note = JSON.parse(d.card_note)
      codeTime.value = note.data?.code_time || ''
    }
    if (!code.value && d.card_note) {
      const m = JSON.parse(d.card_note).data?.code?.match(/\d+/)
      if (m) code.value = m[0]
    }
    if (!code.value) {
      code.value = ''
      result.value = false
      throw new Error('未提取到验证码')
    }
    result.value = true
  } catch (e: any) {
    code.value = ''
    codeTime.value = ''
    result.value = false
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function copyCode() {
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(code.value)
      copyText.value = '已复制！'
      toast('复制成功', 'success')
      setTimeout(() => (copyText.value = '复制'), 1500)
      return
    }
    const ta = document.createElement('textarea')
    ta.value = code.value
    ta.style.position = 'fixed'
    ta.style.top = '-9999px'
    document.body.appendChild(ta)
    ta.focus()
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    if (ok) {
      copyText.value = '已复制！'
      toast('复制成功', 'success')
      setTimeout(() => (copyText.value = '复制'), 1500)
    } else {
      throw new Error('复制失败')
    }
  } catch {
    toast('复制失败', 'error')
  }
}

function formatTime(t: string) {
  // 显示时间：若为 RFC3339 则转本地中文格式，否则原样显示
  if (!t) return '—'
  try { return t.includes('T') ? new Date(t).toLocaleString('zh-CN') : t }
  catch { return t }
}
</script>

<style scoped>
.card { background:#fff; padding:24px; border-radius:16px; box-shadow:0 4px 20px rgba(0,0,0,.08); }
h3 { margin-bottom:20px; color:#007bff; font-size:20px; text-align:center; }
.field { margin-bottom:18px; }
label { display:block; margin-bottom:8px; font-weight:500; color:#555; }
input { width:100%; padding:14px 16px; border:1px solid #ddd; border-radius:10px; font-size:16px; }
input:focus { outline:none; border-color:#007bff; }
.code-line { display:flex; gap:10px; align-items:center; }
input[readonly] { background:#f8f9fa; }
.btn-query, .btn-copy {
  padding:14px; border:none; border-radius:10px; font-size:16px; transition:all .2s;
}
.btn-query { background:#007bff; color:#fff; width:100%; font-weight:500; }
.btn-query:hover { background:#0056b3; }
.btn-query:disabled { background:#aaa; cursor:not-allowed; }
.btn-copy { background:#6c757d; color:#fff; padding:10px 16px; font-size:14px; }
.btn-copy:hover { background:#545b62; }
.result { margin-top:24px; padding:16px; background:#f8f9fa; border-radius:10px; }
.item { display:flex; justify-content:space-between; padding:8px 0; border-bottom:1px solid #eee; }
.item:last-child { border:none; }
strong { color:#007bff; font-weight:600; }
.error { color:#e54545; margin-top:12px; text-align:center; }
</style>
