<script setup>
import { reactive, onMounted, ref, computed } from 'vue'
import { SelectFolder } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

const state = reactive({
  monitoredPath: '',
  events: [],
  filterQuery: '',
})

const filteredEvents = computed(() => {
  if (!state.filterQuery) return state.events
  return state.events.filter(e => e.name.toLowerCase().includes(state.filterQuery.toLowerCase()))
})

const handleSelectFolder = async () => {
  try {
    const path = await SelectFolder()
    if (path) {
      state.monitoredPath = path
      state.events = []
    }
  } catch (err) {
    console.error("选择文件夹失败:", err)
  }
}

onMounted(() => {
  EventsOn('file-event', (data) => {
    state.events.unshift({
      id: Math.random().toString(36).substr(2, 9),
      time: new Date().toLocaleTimeString(),
      ...data
    })
    if (state.events.length > 200) state.events.pop()
  })
})
</script>

<template>
  <div class="window-content">
    <aside class="sidebar">
      <div class="brand">
        <div class="logo-orb"></div>
        <span>Monitor</span>
      </div>
      
      <nav class="nav-stack">
        <div class="nav-item active">
          <i class="icon-monitor"></i> 实时监控
        </div>
      </nav>

      <div class="sidebar-footer">
        <button class="action-btn" @click="handleSelectFolder">
          <span class="plus">+</span> 监控新目录
        </button>
      </div>
    </aside>

    <main class="main-stage">
      <header class="top-bar">
        <div class="search-box">
          <input v-model="state.filterQuery" placeholder="搜索文件事件..." type="text" />
        </div>
        <div v-if="state.monitoredPath" class="path-chip">
          {{ state.monitoredPath }}
        </div>
      </header>

      <section class="viewport">
        <div v-if="state.events.length === 0" class="welcome-view">
          <div class="hero-icon">📁</div>
          <h2>开启您的监控之旅</h2>
          <p>选择一个文件夹，我们将实时捕捉其中的每一个细微变化。</p>
          <button v-if="!state.monitoredPath" class="cta-btn" @click="handleSelectFolder">立即开始</button>
        </div>

        <TransitionGroup name="staggered" tag="div" class="event-scroll">
          <div v-for="event in filteredEvents" :key="event.id" class="glass-card">
            <div class="card-indicator" :class="event.op.toLowerCase()"></div>
            <div class="card-body">
              <div class="card-header">
                <span class="op-tag" :class="event.op.toLowerCase()">{{ event.op }}</span>
                <span class="timestamp">{{ event.time }}</span>
              </div>
              <div class="filename">{{ event.name }}</div>
            </div>
          </div>
        </TransitionGroup>
      </section>
    </main>
  </div>
</template>

<style scoped>
.window-content {
  display: flex;
  height: 100vh;
  width: 100vw;
  user-select: none;
}

/* 侧边栏 */
.sidebar {
  width: 240px;
  background: rgba(255, 255, 255, 0.02);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  padding: 32px 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 40px;
  padding: 0 8px;
}

.logo-orb {
  width: 24px;
  height: 24px;
  background: linear-gradient(135deg, #0078d4, #00bcf2);
  border-radius: 50%;
  box-shadow: 0 0 15px rgba(0, 120, 212, 0.4);
}

.nav-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-dim);
  cursor: pointer;
  transition: all 0.2s;
}

.nav-item.active {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main);
  font-weight: 500;
}

.sidebar-footer {
  margin-top: auto;
}

.action-btn {
  width: 100%;
  background: var(--accent-primary);
  border: none;
  color: white;
  padding: 12px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: background 0.2s;
}

.action-btn:hover { background: var(--accent-hover); }

/* 主舞台 */
.main-stage {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: rgba(0, 0, 0, 0.05);
}

.top-bar {
  height: 70px;
  display: flex;
  align-items: center;
  padding: 0 32px;
  gap: 20px;
  border-bottom: 1px solid var(--glass-border);
}

.search-box {
  flex: 1;
  max-width: 400px;
}

.search-box input {
  width: 100%;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  padding: 8px 16px;
  border-radius: 20px;
  color: white;
  outline: none;
}

.path-chip {
  background: rgba(255, 255, 255, 0.05);
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  color: var(--text-dim);
}

.viewport {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
}

.welcome-view {
  text-align: center;
  margin-top: 100px;
}

.hero-icon { font-size: 64px; margin-bottom: 24px; }
.welcome-view h2 { font-size: 28px; margin-bottom: 12px; }
.welcome-view p { color: var(--text-dim); margin-bottom: 32px; }

.cta-btn {
  background: transparent;
  border: 1px solid var(--accent-primary);
  color: var(--accent-primary);
  padding: 12px 32px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.cta-btn:hover { background: var(--accent-primary); color: white; }

/* 玻璃卡片 */
.event-scroll {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.glass-card {
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  display: flex;
  overflow: hidden;
  box-shadow: var(--card-shadow);
  transition: transform 0.2s;
}

.glass-card:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.15);
}

.card-indicator {
  width: 4px;
  background: #666;
}

.card-indicator.create { background: #4ade80; }
.card-indicator.write { background: #60a5fa; }
.card-indicator.remove { background: #f87171; }
.card-indicator.rename { background: #fbbf24; }

.card-body {
  padding: 16px;
  flex: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.op-tag {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.op-tag.create { color: #4ade80; }
.op-tag.write { color: #60a5fa; }
.op-tag.remove { color: #f87171; }
.op-tag.rename { color: #fbbf24; }

.timestamp { font-size: 11px; color: var(--text-dim); }

.filename {
  font-size: 14px;
  font-weight: 500;
  word-break: break-all;
}

/* 动画 */
.staggered-enter-active {
  transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
}
.staggered-enter-from {
  opacity: 0;
  transform: scale(0.95) translateY(20px);
}
</style>
