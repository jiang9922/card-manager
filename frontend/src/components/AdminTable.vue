<template>
  <div class="admin">
    <div class="toolbar">
      <button @click="confirmDelete" :disabled="selected.length === 0" class="btn-del">
        删除选中项
      </button>
      <button @click="exportSelected" :disabled="selected.length === 0" class="btn-export">
        导出选中项
      </button>
      <!-- 分页大小设置 -->
      <div class="page-size-control">
        <label>每页显示：</label>
        <select v-model="pageSize" @change="changePageSize">
          <option value="10">10</option>
          <option value="20">20</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </select>
        <span>条</span>
      </div>
    </div>
    
    <!-- 筛选区域 -->
    <div class="filter-section">
      <div class="filter-row">
        <div class="filter-item">
          <label>卡号搜索：</label>
          <input type="text" v-model="filters.cardNo" @input="applyFilters" placeholder="输入卡号搜索" />
        </div>
        <div class="filter-item">
          <label>日期筛选：</label>
          <input type="date" v-model="filters.date" @change="applyFilters" />
        </div>
        <div class="filter-item">
          <label>状态筛选：</label>
          <select v-model="filters.status" @change="applyFilters">
            <option value="">全部</option>
            <option value="checked">已获取</option>
            <option value="unchecked">未获取</option>
          </select>
        </div>
        <div class="filter-item">
          <button @click="clearFilters" class="btn-clear">清除筛选</button>
        </div>
      </div>
    </div>
    
    <table>
      <thead>
        <tr>
          <th>
            <label class="select-all">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
              全选
            </label>
          </th>
          <th>卡号</th><th>链接</th><th>状态</th><th>添加时间</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="c in displayedCards" :key="c.id">
          <td><input type="checkbox" :value="c.id" v-model="selected" /></td>
          <td>
            {{ c.card_no }}
            <a :href="c.query_url || `/query?card=${c.card_no}`" class="query-link">查询</a>
          </td>
          <td>
            <a :href="c.card_link" target="_blank" class="link" :title="c.card_link">
              {{ truncate(c.card_link, 50) }}
            </a>
          </td>
          <td>{{ c.card_check ? '已获取' : '未获取' }}</td>
          <td>
            <div class="time-cell">
              <div class="date">{{ formatDate(c.created_at) }}</div>
              <div class="time">{{ formatTime(c.created_at) }}</div>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- 分页控件 -->
    <div class="pagination" v-if="pagination.total_pages > 1">
      <button @click="goToPage(1)" :disabled="pagination.page === 1">首页</button>
      <button @click="goToPage(pagination.page - 1)" :disabled="pagination.page === 1">上一页</button>
      <span class="page-info">
        第 {{ pagination.page }} 页，共 {{ pagination.total_pages }} 页
      </span>
      <button @click="goToPage(pagination.page + 1)" :disabled="pagination.page === pagination.total_pages">下一页</button>
      <button @click="goToPage(pagination.total_pages)" :disabled="pagination.page === pagination.total_pages">末页</button>
    </div>

    <div class="add-section">
      <h3>批量添加卡密</h3>
      <p>格式：每行一个 卡号----链接</p>
      <textarea v-model="input" placeholder="888888----http://localhost:8081/api12828798dss
888888----http://localhost:8081/api871282"></textarea>
      <div class="validate">
        <span class="ok">有效 {{ validCount }} 条</span>
        <span class="bad" v-if="errors.length">无效 {{ errors.length }} 条</span>
        <ul v-if="errors.length">
          <li v-for="e in errors.slice(0,5)" :key="e">{{ e }}</li>
        </ul>
      </div>
      <button @click="add" :disabled="adding" class="btn-add">
        {{ adding ? '提交中...' : '提交添加' }}
      </button>
      <div v-if="msg" :class="msgType">{{ msg }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
// 管理表格逻辑说明：
// - 列表加载：`GET /api/cards`，支持日期与状态筛选，分页展示
// - 批量添加：`POST /api/cards`，按行解析“卡号----链接”
// - 批量删除：`DELETE /api/admin/batch-delete` 按 id 删除
// - 批量导出（前端生成）：导出为“卡号----查询页链接(card_enc)”文本
// - 选中项支持全选，分页切换时清空已选，以避免误删
import { ref, onMounted, computed, watch } from 'vue'
import { useToast } from '../composables/useToast'

const toast = useToast()
// 列表与分页数据
const cards = ref<any[]>([])
const displayedCards = ref<any[]>([])
const pagination = ref({
  page: 1,
  page_size: 10,
  total: 0,
  total_pages: 0
})

// 分页大小
const pageSize = ref(10)

// 批量添加输入与校验
const input = ref('')
const adding = ref(false)
const msg = ref('')
const msgType = ref('')
// 选中项与校验错误统计
const selected = ref<number[]>([])
const errors = ref<string[]>([])
const validCount = ref(0)

// 筛选条件
const filters = ref({
  cardNo: '',
  date: '',
  status: ''
})

onMounted(() => {
  load(1)
})

async function load(page: number = 1) {
  // 构建列表查询 URL，并应用日期/状态筛选
  try {
    // 构建查询参数
    let url = `/api/cards?page=${page}&page_size=${pageSize.value}`
    
    // 添加筛选参数
    if (filters.value.date) {
      url += `&date=${filters.value.date}`
    }
    if (filters.value.status) {
      url += `&status=${filters.value.status}`
    }
    
    const res = await fetch(url)
    const json = await res.json()
    
    if (json.code === 0 && json.data) {
      cards.value = Array.isArray(json.data.cards) ? json.data.cards : []
      displayedCards.value = cards.value
      // 应用当前筛选
      applyFilters()
      pagination.value = json.data.pagination || {
        page: 1,
        page_size: pageSize.value,
        total: 0,
        total_pages: 0
      }
    } else {
      cards.value = []
      displayedCards.value = []
      pagination.value = {
        page: 1,
        page_size: pageSize.value,
        total: 0,
        total_pages: 0
      }
      toast('加载失败', 'error')
    }
  } catch (err) {
    console.error('加载失败:', err)
    toast('加载失败', 'error')
  }
}

function changePageSize() {
  // 改变分页大小后回到第一页并重新拉取数据
  load(1)
}

function applyFilters() {
  // 前端筛选：根据卡号搜索、日期、状态筛选
  let result = cards.value
  
  // 卡号搜索
  if (filters.value.cardNo.trim()) {
    const keyword = filters.value.cardNo.toLowerCase()
    result = result.filter(c => c.card_no.toLowerCase().includes(keyword))
  }
  
  // 状态筛选
  if (filters.value.status === 'checked') {
    result = result.filter(c => c.card_check)
  } else if (filters.value.status === 'unchecked') {
    result = result.filter(c => !c.card_check)
  }
  
  // 日期筛选
  if (filters.value.date) {
    result = result.filter(c => {
      const cardDate = c.created_at.split('T')[0]
      return cardDate === filters.value.date
    })
  }
  
  displayedCards.value = result
}

function clearFilters() {
  // 清除筛选条件
  filters.value.cardNo = ''
  filters.value.date = ''
  filters.value.status = ''
  // 恢复显示所有数据
  displayedCards.value = cards.value
}

function goToPage(page: number) {
  // 分页跳转并清空选中项，避免跨页误操作
  if (page >= 1 && page <= pagination.value.total_pages && page !== pagination.value.page) {
    load(page)
    // 清空选中项
    selected.value = []
  }
}

async function del(ids: number[]) {
  // 后端批量删除接口：传递 id 数组
  try {
    await fetch('/api/admin/batch-delete', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids })
    })
    toast('删除成功', 'success')
    selected.value = []
    // 重新加载当前页
    load(pagination.value.page)
  } catch { toast('删除失败', 'error') }
}

function confirmDelete() {
  // 弹窗确认后执行删除
  if (selected.value.length === 0) return
  if (confirm(`确定要删除选中的 ${selected.value.length} 项吗？`)) {
    del(selected.value)
  }
}

// function exportSelected() {
//   // 前端导出：生成 “卡号 空格 查询页链接” 文本并下载
//   if (selected.value.length === 0) return
  
//   try {
//     // 获取选中的卡片数据
//     const selectedCards = cards.value.filter(card => selected.value.includes(card.id))
    
//     // 生成导出内容
//     const origin = import.meta.env.VITE_EXPORT_ORIGIN || 'http://8.138.84.103'
//     const lines = selectedCards.map(card => {
//       const encodedCardNo = encCard(card.card_no)
//       const linkParam = encUrl(card.card_link)
//       return `${card.card_no} ${origin}/query?card_enc=${encodedCardNo}&link_enc=${linkParam}`
//     })
    
//     const content = lines.join('\n')
    
//     // 创建并下载文件
//     const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
//     const url = window.URL.createObjectURL(blob)
//     const a = document.createElement('a')
//     a.href = url
//     a.download = `cards_export_${new Date().toISOString().slice(0, 19).replace(/:/g, '')}.txt`
//     document.body.appendChild(a)
//     a.click()
//     document.body.removeChild(a)
//     window.URL.revokeObjectURL(url)
    
//     toast('导出成功', 'success')
//   } catch {
//     toast('导出失败', 'error')
//   }
// }
function exportSelected() {
  // 前端导出：生成 “卡号 空格 查询页链接” 文本并下载
  if (selected.value.length === 0) return
  
  try {
    // 获取选中的卡片数据
    const selectedCards = cards.value.filter(card => selected.value.includes(card.id))
    
    // 生成导出内容（只包含卡号，不包含链接参数）
    const origin = import.meta.env.VITE_EXPORT_ORIGIN || 'http://8.138.84.103'
    const lines = selectedCards.map(card => {
      // 只导出卡号参数，不导出链接参数
      return `${card.card_no} ${origin}/query?card=${card.card_no}`
    })
    
    const content = lines.join('\n')
    
    // 创建并下载文件
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cards_export_${new Date().toISOString().slice(0, 19).replace(/:/g, '')}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
    
    toast('导出成功', 'success')
  } catch {
    toast('导出失败', 'error')
  }
}

async function add() {
  // 批量添加：校验文本后提交到后端
  const text = input.value.trim()
  if (!text) return toast('请输入内容', 'error')
  validate()
  if (errors.value.length) return toast('存在格式错误，请修正后再提交', 'error')

  adding.value = true
  try {
    const res = await fetch('/api/cards', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text })
    })
    const json = await res.json()
    if (json.code === 0) {
      input.value = ''
      toast(json.message || '添加成功！', 'success')
      // 添加成功后回到第一页并重新加载
      load(1)
    } else {
      toast(json.message, 'error')
    }
  } catch {
    toast('网络错误', 'error')
  } finally {
    adding.value = false
  }
}

function validate() {
  // 文本校验：每行需满足 “数字卡号 + 有效 URL”，并去重
  const errs: string[] = []
  let ok = 0
  const seen = new Set<string>()
  for (const raw of input.value.split('\n')) {
    const line = raw.trim()
    if (!line) continue
    const parts = line.split('----').map(s => s.trim())
    if (parts.length !== 2) { errs.push(`格式错误: ${line}`); continue }
    const [no, link] = parts
    // 修复TypeScript类型错误，确保no和link不为undefined
    if (!no || !link) { errs.push(`格式错误: ${line}`); continue }
    if (!/^\d+$/.test(no)) { errs.push(`卡号需为数字: ${no}`); continue }
    try { new URL(link) } catch { errs.push(`链接无效: ${link}`); continue }
    if (seen.has(no)) { errs.push(`重复卡号: ${no}`); continue }
    seen.add(no)
    ok++
  }
  errors.value = errs
  validCount.value = ok
}

watch(input, validate)

function truncate(str: string, n: number) {
  return str.length > n ? str.slice(0, n - 3) + '...' : str
}

function encCard(no: string) {
  // 生成可分享的查询链接参数：base64(encodeURIComponent(card_no))
  try { return btoa(encodeURIComponent(no)) } catch { return '' }
}

function encUrl(u: string) {
  try { return btoa(u) } catch { return '' }
}

const isAllSelected = computed(() => cards.value.length > 0 && selected.value.length === cards.value.length)

function toggleSelectAll() {
  // 全选/取消全选
  if (isAllSelected.value) {
    selected.value = []
  } else {
    selected.value = cards.value.map((c: any) => c.id)
  }
}

function formatDate(s: string) {
  // 显示日期（YYYY-MM-DD），异常时返回原字符串
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function formatTime(s: string) {
  // 显示时间（HH:mm:ss），异常时为空
  const d = new Date(s)
  if (isNaN(d.getTime())) return ''
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}
</script>

<style scoped>
.admin { background:#fff; padding:24px; border-radius:16px; box-shadow:0 4px 20px rgba(0,0,0,.08); }
.toolbar { 
  margin-bottom: 16px; 
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}
.toolbar button { 
  padding: 8px 16px; 
  margin-right: 10px; 
  border: none; 
  border-radius: 6px; 
  cursor: pointer;
}
.toolbar button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-del { background:#e74c3c; color:#fff; }
.btn-del:hover:not(:disabled) { background:#c0392b; }
.btn-export { background:#3498db; color:#fff; }
.btn-export:hover:not(:disabled) { background:#2980b9; }

/* 分页大小控制样式 */
.page-size-control {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.page-size-control label {
  font-weight: 500;
  color: #333;
}

.page-size-control select {
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.page-size-control span {
  color: #666;
}

/* 筛选样式 */
.filter-section {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 8px;
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  align-items: center;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item label {
  font-weight: 500;
  color: #333;
  white-space: nowrap;
}

.filter-item input, .filter-item select {
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.btn-clear {
  padding: 6px 12px;
  background: #6c757d;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-clear:hover {
  background: #5a6268;
}

table { width:100%; border-collapse:collapse; margin-bottom:30px; }
th, td { padding:14px; text-align:left; border-bottom:1px solid #eee; }
th { background:#4CAF50; color:#fff; }
.link { color:#0066cc; max-width:200px; display:inline-block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.btn-del-table { background:#e74c3c; color:#fff; border:none; padding:6px 12px; border-radius:6px; }
.btn-del-table:hover { background:#c0392b; }
.time-cell { display:flex; flex-direction:column; }
.time-cell .date { font-weight:600; color:#333; }
.time-cell .time { color:#888; font-size:12px; }
.admin input[type="checkbox"] { width:16px; height:16px; }
.select-all { display:flex; align-items:center; gap:6px; }
.query-link { margin-left:8px; font-size:12px; color:#007bff; }
.add-section h3 { margin:20px 0 10px; color:#333; }
textarea { width:100%; height:120px; padding:12px; border:1px solid #ddd; border-radius:8px; margin:10px 0; resize:vertical; }
.validate { display:block; margin:6px 0 10px; }
.validate .ok { color:#27ae60; margin-right:10px; }
.validate .bad { color:#e74c3c; }
.validate ul { margin:6px 0 0 0; padding-left:18px; color:#e74c3c; }
.btn-add { background:#27ae60; color:#fff; padding:10px 20px; border:none; border-radius:8px; }
.btn-add:disabled { background:#95a5a6; }

/* 分页样式 */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.pagination button {
  padding: 8px 12px;
  border: 1px solid #ddd;
  background: #fff;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
}

.pagination button:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
  color: #999;
}

.pagination button:not(:disabled):hover {
  background: #f0f0f0;
}

.page-info {
  font-size: 14px;
  color: #666;
  white-space: nowrap;
}

@media (max-width: 600px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .page-size-control {
    margin-left: 0;
    justify-content: center;
  }
  
  .filter-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-item {
    justify-content: space-between;
  }
  
  .pagination {
    gap: 5px;
  }
  
  .pagination button {
    padding: 6px 10px;
    font-size: 13px;
  }
  
  .page-info {
    font-size: 13px;
  }
}
</style>