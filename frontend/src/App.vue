<script setup>
import { reactive, onMounted, ref, computed, watch } from 'vue'
import { 
  SelectFolder, GetDiskInfo, GetDirectoryInsight, LocateFile,
  GetTopFiles, ScanCleanup, ExecuteCleanup, GetSecurityAudit, GetFilesByExt 
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

const state = reactive({
  activeModule: 'activity', // activity, insights, cleanup, security
  monitoredPath: '',
  events: [],
  filterQuery: '',
  stats: { total: 0, create: 0, write: 0, remove: 0, rename: 0, folders: 0 },
  disk: { total: '0 B', free: '0 B', used: '0 B', freeBytes: 0, usage: 0 },
  insight: { totalSize: '0 B', fileCount: 0, dirCount: 0, categories: {}, extDetails: {} },
  topFiles: [],
  cleanupFiles: [],
  cleanupSelected: new Set(),
  securityAudit: [],
  isScanning: false,
  isCleaning: false,
  scanProgress: { scanned: 0, current: '' },
  drilldown: { active: false, ext: '', name: '', files: [] } // 新增钻取状态
})

// 计算属性
const filteredEvents = computed(() => {
  if (!state.filterQuery) return state.events
  const q = state.filterQuery.toLowerCase()
  return state.events.filter(e => e.name.toLowerCase().includes(q) || e.op.toLowerCase().includes(q))
})

// 20+ 开发语言扩展名列表
const DEVELOPER_EXTS = {
  '.go': 'Go', '.py': 'Python', '.js': 'JavaScript', '.ts': 'TypeScript', '.vue': 'Vue',
  '.java': 'Java', '.cpp': 'C++', '.c': 'C', '.h': 'Header', '.cs': 'C#',
  '.php': 'PHP', '.sql': 'SQL', '.html': 'HTML', '.css': 'CSS', '.sh': 'Shell',
  '.json': 'JSON', '.yaml': 'YAML', '.yml': 'YAML', '.rb': 'Ruby', '.rs': 'Rust',
  '.swift': 'Swift', '.kt': 'Kotlin', '.xml': 'XML', '.md': 'Markdown'
}

const languageStats = computed(() => {
  const details = state.insight.extDetails || {}
  return Object.entries(DEVELOPER_EXTS)
    .map(([ext, name]) => ({ ext, name, count: details[ext] || 0 }))
    .filter(item => item.count > 0)
    .sort((a, b) => b.count - a.count)
})

const insightSegments = computed(() => {
  if (state.insight.fileCount === 0) return []
  const cats = state.insight.categories; const total = state.insight.fileCount
  const getWeight = (exts) => exts.reduce((acc, e) => acc + (cats[e] || 0), 0)
  
  const codeExts = Object.keys(DEVELOPER_EXTS)
  const officeExts = ['.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.pdf', '.txt', '.csv']
  const imgExts = ['.jpg', '.jpeg', '.png', '.gif', '.svg', '.webp', '.bmp', '.ico']
  const vidExts = ['.mp4', '.mkv', '.mov', '.avi', '.mp3', '.wav', '.flac', '.wmv']
  
  const code = getWeight(codeExts)
  const office = getWeight(officeExts)
  const imgs = getWeight(imgExts)
  const vids = getWeight(vidExts)
  const other = total - code - office - imgs - vids

  return [
    { type: 'code', label: '代码', color: '#a855f7', icon: 'M16 18l6-6-6-6M8 6l-6 6 6 6', width: (code / total * 100) + '%' },
    { type: 'office', label: '办公', color: '#f97316', icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z', width: (office / total * 100) + '%' },
    { type: 'img', label: '图片', color: '#60a5fa', icon: 'M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z', width: (imgs / total * 100) + '%' },
    { type: 'vid', label: '媒体', color: '#f87171', icon: 'M23 7l-7 5 7 5V7zM1 5h11v14H1z', width: (vidExts / total * 100) + '%' },
    { type: 'other', label: '其他', color: '#94a3b8', icon: 'M12 2v20M2 12h20', width: (other / total * 100) + '%' }
  ].filter(s => parseFloat(s.width) > 0)
})

const selectedCleanupFiles = computed(() => Array.from(state.cleanupSelected))
const totalCleanupSize = computed(() => {
  const bytes = state.cleanupFiles
    .filter(f => state.cleanupSelected.has(f.path))
    .reduce((acc, f) => acc + (f.bytes || 0), 0)
  
  // 简易 formatSize 逻辑 (前端版)
  if (bytes === 0) return '0 B'
  const k = 1024; const dm = 2; const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
})

// 核心逻辑
const fetchModuleData = async () => {
  if (!state.monitoredPath) return
  state.isScanning = true
  state.scanProgress = { scanned: 0, current: '准备扫描...' }
  try {
    if (state.activeModule === 'insights') {
      const ins = await GetDirectoryInsight(state.monitoredPath)
      state.insight = ins
      state.topFiles = await GetTopFiles(state.monitoredPath)
    } else if (state.activeModule === 'cleanup') {
      state.cleanupFiles = await ScanCleanup(state.monitoredPath)
      state.cleanupSelected.clear()
    } else if (state.activeModule === 'security') {
      state.securityAudit = await GetSecurityAudit()
    }
    const d = await GetDiskInfo(state.monitoredPath)
    state.disk = d
  } finally {
    state.isScanning = false
  }
}

watch(() => state.activeModule, fetchModuleData)

const handleSelectFolder = async () => {
  try {
    const path = await SelectFolder()
    if (path) {
      state.monitoredPath = path
      state.events = []
      Object.keys(state.stats).forEach(k => state.stats[k] = 0)
      fetchModuleData()
    }
  } catch (err) { console.error(err) }
}

const handleCleanup = async () => {
  if (state.cleanupSelected.size === 0) return
  state.isCleaning = true
  await ExecuteCleanup(Array.from(state.cleanupSelected))
  state.cleanupFiles = state.cleanupFiles.filter(f => !state.cleanupSelected.has(f.path))
  state.cleanupSelected.clear()
  state.isCleaning = false
  fetchModuleData()
}

onMounted(() => {
  EventsOn('file-event', (data) => {
    // 使用后端推送的高精度 time
    const event = { 
      id: Math.random().toString(36).substr(2, 9), 
      ...data 
    }
    state.events.unshift(event)
    state.stats.total++
    const op = data.op.toLowerCase()
    if (state.stats[op] !== undefined) state.stats[op]++
    if (data.isDir) state.stats.folders++
    if (state.events.length > 200) state.events.pop()
    
    if (data.isSensitive && state.activeModule === 'security') {
      GetSecurityAudit().then(res => state.securityAudit = res)
    }
    // 实时更新磁盘信息
    GetDiskInfo(state.monitoredPath).then(d => state.disk = d)
  })

  // 监听扫描进度
  EventsOn('scan-progress', (data) => {
    state.scanProgress = data
  })
})

// 分类钻取逻辑 (v8.4)
const handleDrilldown = async (item) => {
  state.isScanning = true
  try {
    const files = await GetFilesByExt(state.monitoredPath, item.ext)
    state.drilldown = {
      active: true,
      ext: item.ext,
      name: item.name,
      files: files
    }
  } finally {
    state.isScanning = false
  }
}

const closeDrilldown = () => {
  state.drilldown.active = false
}

const icons = {
  activity: 'M22 12h-4l-3 9L9 3l-3 9H2',
  insights: 'M21.21 15.89A10 10 0 1 1 8 2.83M22 12A10 10 0 0 0 12 2v10z',
  cleanup: 'M3 6h18M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2',
  security: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
  folder: 'M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z',
  file: 'M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z m0 0v7h7'
}
</script>

<template>
  <div class="window-content">
    <aside class="sidebar">
      <div class="brand">
        <!-- v8.5 炫酷 SVG Logo -->
        <div class="logo-stack">
          <svg viewBox="0 0 100 100" width="48" height="48" class="main-logo">
            <defs>
              <linearGradient id="logo-grad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" style="stop-color:#0078d4;stop-opacity:1" />
                <stop offset="100%" style="stop-color:#39ff14;stop-opacity:1" />
              </linearGradient>
            </defs>
            <circle cx="50" cy="50" r="45" fill="none" stroke="url(#logo-grad)" stroke-width="2" stroke-dasharray="10 5" class="outer-ring"/>
            <path d="M30 50 Q 40 20, 50 50 T 70 50" fill="none" stroke="#39ff14" stroke-width="4" stroke-linecap="round" class="pulse-line"/>
            <rect x="42" y="42" width="16" height="16" rx="4" fill="#0078d4" class="core-box"/>
          </svg>
        </div>
        <span>Monitor Pro</span>
      </div>

      <div class="disk-status neon-border">
        <div class="disk-info-header">
          <div class="disk-title-group">
            <div class="radar-icon"></div>
            <span>磁盘态势感知</span>
          </div>
          <span style="color: var(--neon-green)">{{ Math.round(state.disk.usage) || 0 }}%</span>
        </div>
        <div class="progress-container">
          <div class="progress-bar" :style="{ width: state.disk.usage + '%' }"></div>
        </div>
        <div class="disk-detail-grid" style="margin-top: 24px; display: flex; flex-direction: column; gap: 12px;">
          <div class="detail-item" style="display: flex; justify-content: space-between; align-items: center;">
            <span class="detail-label" style="font-size: 11px; opacity: 0.6;">已使用空间</span>
            <span class="detail-value used-val" style="font-size: 15px; font-weight: 800; color: var(--neon-green);">{{ state.disk.used || '0 B' }}</span>
          </div>
          <div class="detail-item" style="display: flex; justify-content: space-between; align-items: center;">
            <span class="detail-label" style="font-size: 11px; opacity: 0.6;">可用剩余</span>
            <span class="detail-value free-val" style="font-size: 15px; font-weight: 800; color: #3b82f6;">{{ state.disk.free || '0 B' }}</span>
          </div>
        </div>
      </div>

      <nav class="nav-stack">
        <div v-for="m in ['activity', 'insights', 'cleanup', 'security']" :key="m" 
             class="nav-item neon-item" :class="{ active: state.activeModule === m }" @click="state.activeModule = m; closeDrilldown()">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path v-if="m==='activity'" :d="icons.activity"/><path v-else-if="m==='insights'" :d="icons.insights"/>
            <path v-else-if="m==='cleanup'" :d="icons.cleanup"/><path v-else :d="icons.security"/>
          </svg>
          <span style="text-transform: capitalize">{{ m==='activity'?'实时活动':m==='insights'?'空间洞察':m==='cleanup'?'清理建议':'安全审计' }}</span>
        </div>
      </nav>

      <div class="sidebar-footer" style="margin-top: auto;">
        <button class="action-btn neon-btn" @click="handleSelectFolder">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg>
          更换分析目标
        </button>
      </div>
    </aside>

    <main class="main-stage">
      <header class="top-bar">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
          <input v-model="state.filterQuery" :placeholder="'在'+(state.activeModule==='activity'?'实时记录':'历史数据')+'中检索...'" />
        </div>
        <div v-if="state.monitoredPath" class="path-chip">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path :d="icons.folder"/></svg>
          <span :data-fulltext="state.monitoredPath" style="max-width: 250px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ state.monitoredPath }}</span>
        </div>
      </header>

      <section class="viewport">
        <!-- 实时扫描指示器 -->
        <div v-if="state.isScanning" class="scan-indicator">
          <div class="scanner-icon">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
            </svg>
          </div>
          <div class="scan-text">
            <strong>正在执行多维深度扫描...</strong> (已发现 {{ state.scanProgress.scanned || 0 }} 资产项目)
            <span class="scan-path" :data-fulltext="state.scanProgress.current">{{ state.scanProgress.current || '正在枚举...' }}</span>
          </div>
        </div>

        <!-- 模块：实时活动 -->
        <div v-if="state.activeModule === 'activity' && !state.drilldown.active" class="view-module">
          <!-- 实时监控呼吸指示器 (v8.3) -->
          <div class="monitoring-status">
            <div class="pulse-dot"></div>
            <span class="status-label">ENGINE ACTIVE</span>
            <span class="status-text">实时防御保护中</span>
          </div>

          <div class="stats-panel" style="display: flex; gap: 20px; padding: 12px; margin-bottom: 20px;">
            <div class="mini-stat"><span class="dot create"></span><span>新建: {{ state.stats.create }}</span></div>
            <div class="mini-stat"><span class="dot write"></span><span>修改: {{ state.stats.write }}</span></div>
            <div class="mini-stat"><span class="dot remove"></span><span>删除: {{ state.stats.remove }}</span></div>
            <div class="mini-stat"><span class="dot folder"></span><span>目录: {{ state.stats.folders }}</span></div>
          </div>
          <div class="event-scroll">
            <TransitionGroup name="staggered">
              <div v-for="event in filteredEvents" :key="event.id" class="glass-card" @click="LocateFile(event.name)">
                <div class="card-indicator" :class="event.op.toLowerCase()"></div>
                <div class="card-body">
                  <div class="card-header"><span class="op-tag" :class="event.op.toLowerCase()">{{ event.op }}</span><span class="timestamp">{{ event.time }}</span></div>
                  <div class="filename" :data-fulltext="event.name" :data-filename="event.name">{{ event.name }}</div>
                </div>
              </div>
            </TransitionGroup>
          </div>
        </div>

        <!-- 模块：空间洞察 -->
        <div v-if="state.activeModule === 'insights' && !state.drilldown.active" class="view-module">
          <div class="section-title">📊 多维分类统计分布</div>
          <div class="insight-card" style="padding: 24px !important;">
            <div class="stat-summary-grid">
              <div class="stat-module">
                <span class="val">{{ state.insight.fileCount }}</span>
                <span class="lab">资产总数</span>
              </div>
              <div class="stat-module">
                <span class="val">{{ state.insight.dirCount }}</span>
                <span class="lab">目录深度</span>
              </div>
              <div class="stat-module">
                <span class="val" style="color: #60a5fa;">{{ state.insight.totalSize }}</span>
                <span class="lab">总体积</span>
              </div>
            </div>
            <div class="insight-bar" style="margin-top: 10px;">
              <div v-for="seg in insightSegments" :key="seg.type" class="insight-segment" :class="seg.type" :style="{ width: seg.width }"></div>
            </div>
            <div class="disk-detail" style="font-size: 11px; flex-wrap: wrap; gap: 16px; margin-top: 20px;">
              <span v-for="seg in insightSegments" :key="seg.type" :style="{ color: seg.color, display: 'flex', alignItems: 'center', gap: '6px' }">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path :d="seg.icon"/></svg>
                <span style="font-weight: 600;">{{ seg.label }}</span>
              </span>
            </div>
          </div>

          <!-- 20+ 开发语言明细 -->
          <div class="section-title" style="margin-top: 32px;">💻 编程语言独立分布 (20+ 识别)</div>
          <div v-if="languageStats.length === 0" class="welcome-view" style="padding: 20px; font-size: 12px; height: auto;">未检测到主流开发代码资产</div>
          <div v-else class="extension-grid">
            <div v-for="item in languageStats" :key="item.ext" class="ext-chip drillable" @click="handleDrilldown(item)">
              <span class="ext-name">{{ item.name }}</span>
              <span class="ext-count">{{ item.count }}</span>
            </div>
          </div>

          <div class="section-title" style="margin-top: 32px;">🔥 冗余及核心资产排行 (TOP 20)</div>
          <div class="ranking-list">
            <div v-for="file in state.topFiles" :key="file.path" class="large-file-item glass-card" @click="LocateFile(file.path)">
              <div class="file-info" style="padding: 12px; flex: 1; overflow: hidden;">
                <div class="file-name-text" :data-fulltext="file.name" style="font-weight: 700; font-size: 14px;">{{ file.name }}</div>
                <div class="file-path-text" :data-fulltext="file.path" style="margin-top: 4px; opacity: 0.5;">{{ file.path }}</div>
                <div class="timestamp" style="margin-top: 8px; font-size: 10px; opacity: 0.8;">最后修改: {{ file.timeDetail }}</div>
              </div>
              <div class="file-size-tag" style="margin-right: 16px; min-width: 80px; text-align: center;">{{ file.size }}</div>
            </div>
          </div>
        </div>

        <!-- 模块：清理建议 -->
        <div v-if="state.activeModule === 'cleanup' && !state.drilldown.active" class="view-module">
          <div class="section-title">🧹 智能清理建议</div>
          <div v-if="state.cleanupFiles.length > 0" class="cleanup-toolbar glass-effect" style="margin-bottom: 24px; padding: 20px; border-radius: 16px; display: flex; align-items: center; justify-content: space-between; background: rgba(0,0,0,0.2); border: 1px solid rgba(255,255,255,0.08);">
            <div class="toolbar-info">
              <div style="font-size: 11px; opacity: 0.5; margin-bottom: 4px; text-transform: uppercase;">Storage Cleanup Task</div>
              <span style="font-size: 14px; font-weight: 600;">已扫描到 {{ state.cleanupFiles.length }} 项冗余 | 可释放: <span style="color: var(--neon-green)">{{ totalCleanupSize }}</span></span>
            </div>
            <div class="module-btn-group">
              <button class="module-btn primary" @click="handleExecuteCleanup" :disabled="state.cleanupSelected.size === 0">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
                一键清理
              </button>
              <button class="module-btn" @click="state.cleanupSelected.clear()">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3h18v18H3zM9 9l6 6m0-6l-6 6"/></svg>
                清空选择
              </button>
            </div>
          </div>
          <div v-if="state.cleanupFiles.length === 0" class="welcome-view">系统非常整洁，暂无清理建议</div>
          <div v-else>
            <div class="layout-grid">
              <div v-for="file in state.cleanupFiles" :key="file.path" class="cleanup-card glass-card" 
                   @click="state.cleanupSelected.has(file.path)?state.cleanupSelected.delete(file.path):state.cleanupSelected.add(file.path)">
                <input type="checkbox" class="cleanup-check" :checked="state.cleanupSelected.has(file.path)" />
                <div class="file-info" style="flex: 1; min-width: 0;">
                  <div class="file-name-text" :data-fulltext="file.name">{{ file.name }}</div>
                  <div class="file-path-text" :data-fulltext="file.path">{{ file.path }}</div>
                </div>
                <div class="file-size-tag">{{ file.size }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 模块：安全审计 -->
        <div v-if="state.activeModule === 'security' && !state.drilldown.active" class="view-module">
          <div class="section-title">🛡️ 安全风险审计</div>
          <div class="event-scroll">
            <div v-for="(ev, idx) in state.securityAudit" :key="idx" class="glass-card" style="margin-bottom: 8px;">
              <div class="card-indicator" :class="ev.isSensitive?'remove':'write'"></div>
              <div class="card-body">
                <div class="card-header">
                  <span class="security-tag" :class="ev.isSensitive?'sensitive':'normal'">{{ ev.isSensitive?'敏感文件改动':'系统变动' }}</span>
                  <span class="op-tag" :class="ev.op.toLowerCase()">{{ ev.op }}</span>
                  <span class="timestamp">{{ ev.time }}</span>
                </div>
                <div class="filename" style="margin-top: 6px;">{{ ev.name }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 钻取详情试图 (v8.4) -->
        <div v-if="state.drilldown.active" class="drilldown-view">
          <div class="section-title">
            <button class="back-btn" @click="closeDrilldown">← 返回指标面板</button>
            <span>分类详情: {{ state.drilldown.name }} ({{ state.drilldown.ext }})</span>
          </div>
          <div class="ranking-list" style="margin-top: 20px;">
            <div v-for="file in state.drilldown.files" :key="file.path" class="large-file-item glass-card" @click="LocateFile(file.path)">
              <div class="file-info" style="padding: 12px; flex: 1; overflow: hidden;">
                <div class="file-name-text" :data-fulltext="file.name" style="font-weight: 700; font-size: 14px;">{{ file.name }}</div>
                <div class="file-path-text" :data-fulltext="file.path" style="margin-top: 4px; opacity: 0.5;">{{ file.path }}</div>
                <div class="timestamp" style="margin-top: 8px; font-size: 10px; opacity: 0.8;">最后修改: {{ file.timeDetail }}</div>
              </div>
              <div class="file-size-tag" style="margin-right: 16px; min-width: 80px; text-align: center;">{{ file.size }}</div>
            </div>
          </div>
        </div>

        <!-- 空白/加载状态 -->
        <div v-if="!state.monitoredPath" class="welcome-view">
          <div class="hero-icon"><svg viewBox="0 0 24 24" width="80" height="80" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></div>
          <h2>分析中心就绪</h2>
          <p>请选择驱动器或目录以激活全量高精度分析模块</p>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.window-content { display: flex; height: 100vh; width: 100vw; user-select: none; background: rgba(0, 0, 0, 0.2); }
.sidebar { width: 280px; background: rgba(0, 0, 0, 0.3); border-right: 1px solid var(--glass-border); display: flex; flex-direction: column; padding: 32px 24px; backdrop-filter: blur(20px); }
.brand { display: flex; flex-direction: column; align-items: center; gap: 16px; font-size: 20px; font-weight: 800; margin-bottom: 40px; color: var(--neon-green); text-shadow: 0 0 10px rgba(57, 255, 20, 0.3); }
.logo-stack { position: relative; width: 64px; height: 64px; display: flex; align-items: center; justify-content: center; }
.outer-ring { animation: rotate 10s linear infinite; transform-origin: center; }
.pulse-line { stroke-dasharray: 100; stroke-dashoffset: 100; animation: dash 2s ease-in-out infinite; }
@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
@keyframes dash { 0% { stroke-dashoffset: 100; opacity: 0.3; } 50% { stroke-dashoffset: 0; opacity: 1; } 100% { stroke-dashoffset: -100; opacity: 0.3; } }
.main-logo { filter: drop-shadow(0 0 8px rgba(0, 120, 212, 0.5)); transition: all 0.5s; }
.logo-stack:hover .main-logo { transform: scale(1.1); filter: drop-shadow(0 0 15px var(--neon-green)); }
.core-box { animation: breathe 3s ease-in-out infinite; transform-origin: center; }
@keyframes breathe { 0%, 100% { transform: scale(1); opacity: 0.8; } 50% { transform: scale(1.1); opacity: 1; } }
.stats-panel { border-radius: 12px; border: 1px solid var(--glass-border); background: rgba(255, 255, 255, 0.05); }
.mini-stat { display: flex; align-items: center; gap: 6px; font-size: 11px; }
.dot { width: 6px; height: 6px; border-radius: 50%; }
.dot.create { background: #4ade80; }
.dot.write { background: #60a5fa; }
.dot.remove { background: #f87171; }
.dot.folder { background: #fbbf24; }
.action-btn { width: 100%; background: var(--accent-primary); border: none; color: white; padding: 12px; border-radius: 10px; font-weight: 600; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 10px; box-shadow: 0 4px 15px rgba(0, 120, 212, 0.3); }
.main-stage { flex: 1; display: flex; flex-direction: column; }
.top-bar { height: 80px; display: flex; align-items: center; padding: 0 32px; gap: 20px; background: rgba(0, 0, 0, 0.1); border-bottom: 1px solid var(--glass-border); }
.search-box { flex: 1; max-width: 460px; position: relative; }
.search-icon { position: absolute; left: 14px; top: 50%; transform: translateY(-50%); color: var(--text-dim); }
.search-box input { width: 100%; background: rgba(0, 0, 0, 0.2); border: 1px solid var(--glass-border); padding: 10px 16px 10px 40px; border-radius: 10px; color: white; font-size: 14px; outline: none; }
.path-chip { background: rgba(255, 255, 255, 0.05); padding: 6px 14px; border-radius: 20px; font-size: 12px; color: var(--text-dim); display: flex; align-items: center; gap: 8px; }
.viewport { flex: 1; padding: 24px 32px; overflow-y: auto; overflow-x: hidden; }
.view-module { animation: fadeIn 0.4s ease-out; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.event-scroll { display: flex; flex-direction: column; gap: 10px; }
.glass-card { background: var(--glass-bg); border: 1px solid var(--glass-border); border-radius: 12px; display: flex; align-items: center; overflow: hidden; box-shadow: var(--card-shadow); transition: all 0.2s; }
.card-indicator { width: 4px; height: 100%; align-self: stretch; }
.card-indicator.create { background: #4ade80; }
.card-indicator.write { background: #60a5fa; }
.card-indicator.remove { background: #f87171; }
.card-indicator.rename { background: #fbbf24; }
.card-body { padding: 14px 16px; flex: 1; overflow: hidden; }
.card-header { display: flex; justify-content: space-between; margin-bottom: 4px; }
.op-tag { font-size: 10px; font-weight: 800; text-transform: uppercase; }
.op-tag.create { color: #4ade80; }
.op-tag.write { color: #60a5fa; }
.op-tag.remove { color: #f87171; }
.op-tag.rename { color: #fbbf24; }
.timestamp { font-size: 11px; color: var(--text-dim); }
.staggered-enter-active { transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1); }
.staggered-enter-from { opacity: 0; transform: scale(0.9) translateX(-20px); }
.welcome-view { text-align: center; margin-top: 15vh; opacity: 0.8; }
.hero-icon { color: var(--accent-primary); margin-bottom: 24px; animation: float 3s ease-in-out infinite; }
@keyframes float { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-10px); } }
</style>
