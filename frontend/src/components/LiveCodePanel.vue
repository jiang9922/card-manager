<template>
  <div class="live-panel">
    <h4>实时验证码</h4>
    <div class="scroll-container" ref="scrollContainer">
      <div v-for="item in liveCodes" :key="item.id" class="code-item">
        <span class="card-tail">{{ item.cardTail }}</span>
        <span class="code">{{ item.code }}</span>
        <span class="time">{{ formatTime(item.time) }}</span>
      </div>
      <div v-if="liveCodes.length === 0" class="empty">暂无数据</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

interface LiveCode {
  id: number
  cardTail: string
  code: string
  time: string
}

const liveCodes = ref<LiveCode[]>([])
const scrollContainer = ref<HTMLElement>()
let timer: number

// 加载最新验证码
async function loadLatestCodes() {
  try {
    const res = await fetch('/api/cards/live-codes?limit=20')
    const json = await res.json()
    if (json.code === 0 && json.data) {
      liveCodes.value = json.data.map((c: any) => ({
        id: c.id,
        cardTail: c.card_no.slice(-4),
        code: c.card_code || '----',
        time: c.card_expired_date || c.created_at
      }))
    }
  } catch {
    console.error('加载实时验证码失败')
  }
}

function formatTime(t: string) {
  if (!t) return '--:--'
  try {
    const d = new Date(t)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  } catch {
    return '--:--'
  }
}

onMounted(() => {
  loadLatestCodes()
  // 每3秒刷新一次
  timer = window.setInterval(loadLatestCodes, 3000)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<style scoped>
.live-panel {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  padding: 16px;
  width: 280px;
  height: 400px;
  display: flex;
  flex-direction: column;
}

h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #333;
  font-weight: 600;
}

.scroll-container {
  flex: 1;
  overflow-y: auto;
  border-radius: 8px;
  background: #f8f9fa;
}

.code-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-bottom: 1px solid #eee;
  font-size: 14px;
}

.code-item:last-child {
  border-bottom: none;
}

.card-tail {
  color: #666;
  font-weight: 500;
  min-width: 60px;
}

.code {
  flex: 1;
  color: #007bff;
  font-weight: 600;
  font-size: 16px;
  letter-spacing: 2px;
  text-align: center;
}

.time {
  color: #999;
  font-size: 12px;
  min-width: 50px;
  text-align: right;
}

.empty {
  text-align: center;
  padding: 40px 20px;
  color: #999;
  font-size: 14px;
}

/* 滚动条样式 */
.scroll-container::-webkit-scrollbar {
  width: 4px;
}

.scroll-container::-webkit-scrollbar-track {
  background: transparent;
}

.scroll-container::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 2px;
}
</style>