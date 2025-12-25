<script setup>
import { reactive, onMounted, ref, computed, watch } from 'vue'
import { 
  SelectFolder, GetDiskInfo, GetDirectoryInsight, LocateFile,
  GetTopFiles, ScanCleanup, ExecuteCleanup, GetSecurityAudit, GetFilesByExt,
  GetFilesByExts, SearchFiles, GetLastPath, StartAutoMonitor
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
  drilldown: { active: false, ext: '', name: '', files: [], sortBy: 'size', sortOrder: 'desc' }, // 分类排序 (v10.1)
  search: { active: false, query: '', results: [], sortBy: 'size', sortOrder: 'desc' }, // 默认降序 (v10.0)
  restored: false
})

const searchStats = computed(() => {
  if (!state.search.results.length) return { files: 0, dirs: 0 }
  const dirs = new Set(state.search.results.map(f => {
    const parts = f.path.split(/[\\/]/)
    parts.pop()
    return parts.join('/')
  }))
  return {
    files: state.search.results.length,
    dirs: dirs.size
  }
})

const sortedSearchResults = computed(() => {
  const results = [...state.search.results]
  const order = state.search.sortOrder === 'desc' ? -1 : 1
  
  return results.sort((a, b) => {
    if (state.search.sortBy === 'size') {
      return (a.bytes - b.bytes) * order
    } else if (state.search.sortBy === 'time') {
      return (a.timestamp - b.timestamp) * order
    } else if (state.search.sortBy === 'name') {
      return a.name.localeCompare(b.name) * order
    }
    return 0
  })
})

const sortedDrilldownFiles = computed(() => {
  const results = [...state.drilldown.files]
  const order = state.drilldown.sortOrder === 'desc' ? -1 : 1
  
  return results.sort((a, b) => {
    if (state.drilldown.sortBy === 'size') {
      return (a.bytes - b.bytes) * order
    } else if (state.drilldown.sortBy === 'time') {
      return (a.timestamp - b.timestamp) * order
    } else if (state.drilldown.sortBy === 'name') {
      return a.name.localeCompare(b.name) * order
    }
    return 0
  })
})

const viewportRef = ref(null)

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
  const vidExts = ['.mp4', '.mkv', '.mov', '.avi', '.wmv']
  const zipExts = ['.zip', '.rar', '.7z', '.tar', '.gz', '.bz2']
  const audioExts = ['.mp3', '.wav', '.flac', '.aac', '.ogg', '.m4a']
  const exeExts = ['.exe', '.msi', '.bat', '.sh', '.cmd', '.py', '.js']
  const dbExts = ['.sqlite', '.db', '.sql', '.json', '.xml', '.yaml']
  const tmpExts = ['.tmp', '.log', '.bak', '.old', '.swp', '.dmp']

  const code = getWeight(codeExts)
  const office = getWeight(officeExts)
  const imgs = getWeight(imgExts)
  const vids = getWeight(vidExts)
  const zips = getWeight(zipExts)
  const audio = getWeight(audioExts)
  const exes = getWeight(exeExts)
  const dbs = getWeight(dbExts)
  const tmps = getWeight(tmpExts)
  const other = total - code - office - imgs - vids - zips - audio - exes - dbs - tmps

  return [
    { type: 'code', label: '核心代码', val: code, exts: codeExts, color: '#a855f7', icon: 'M16 18l6-6-6-6M8 6l-6 6 6 6', width: (code / total * 100) + '%' },
    { type: 'office', label: '办公文档', val: office, exts: officeExts, color: '#f97316', icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z', width: (office / total * 100) + '%' },
    { type: 'img', label: '设计图形', val: imgs, exts: imgExts, color: '#60a5fa', icon: 'M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z', width: (imgs / total * 100) + '%' },
    { type: 'vid', label: '视频资产', val: vids, exts: vidExts, color: '#f87171', icon: 'M21 15V9a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2z', width: (vids / total * 100) + '%' },
    { type: 'zip', label: '压缩归档', val: zips, exts: zipExts, color: '#fbbf24', icon: 'M21 10h-6V4l-4 4 4 4V8h4v2h-4v2h4v2h-4v2h4v2h-4v2h4v2h-4v2z', width: (zips / total * 100) + '%' },
    { type: 'db', label: '数据结构', val: dbs, exts: dbExts, color: '#2dd4bf', icon: 'M4 6c0 1.66 3.58 3 8 3s8-1.34 8-3-3.58-3-8-3-8 1.34-8 3M4 10c0 1.66 3.58 3 8 3s8-1.34 8-3M4 14c0 1.66 3.58 3 8 3s8-1.34 8-3', width: (dbs / total * 100) + '%' },
    { type: 'exe', label: '执行脚本', val: exes, exts: exeExts, color: '#ef4444', icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z', width: (exes / total * 100) + '%' },
    { type: 'audio', label: '音频素材', val: audio, exts: audioExts, color: '#34d399', icon: 'M9 18V5l12-2v13', width: (audio / total * 100) + '%' },
    { type: 'tmp', label: '系统冗余', val: tmps, exts: tmpExts, color: '#94a3b8', icon: 'M19 6L5 20M5 6l14 14', width: (tmps / total * 100) + '%' }
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

onMounted(async () => {
  // 恢复上次路径 (v9.6)
  const lastPath = await GetLastPath()
  if (lastPath) {
    state.monitoredPath = lastPath
    await StartAutoMonitor(lastPath)
    state.restored = true
    setTimeout(() => state.restored = false, 3000)
    fetchModuleData()
  }

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

const toggleSort = (key, mode = 'search') => {
  if (mode === 'search') {
    if (state.search.sortBy === key) {
      state.search.sortOrder = state.search.sortOrder === 'desc' ? 'asc' : 'desc'
    } else {
      state.search.sortBy = key
      state.search.sortOrder = 'desc'
    }
  } else {
    if (state.drilldown.sortBy === key) {
      state.drilldown.sortOrder = state.drilldown.sortOrder === 'desc' ? 'asc' : 'desc'
    } else {
      state.drilldown.sortBy = key
      state.drilldown.sortOrder = 'desc'
    }
  }
}

// 增强型钻取逻辑 (v9.2) - 支持多扩展名批量扫描
const handleDrilldown = async (item) => {
  state.isScanning = true
  try {
    // 根据 item 是否包含 exts 数组选择接口
    const files = item.exts 
      ? await GetFilesByExts(state.monitoredPath, item.exts)
      : await GetFilesByExt(state.monitoredPath, item.ext)
      
    state.drilldown = {
      active: true,
      ext: item.ext || (item.exts ? item.exts[0] : ''),
      name: item.label || item.name,
      files: files
    }
  } finally {
    state.isScanning = false
  }
}

// 全局递归检索逻辑 (v9.5)
const handleGlobalSearch = async () => {
  if (!state.filterQuery) {
    state.search.active = false
    return
  }
  state.isScanning = true
  try {
    const results = await SearchFiles(state.monitoredPath, state.filterQuery)
    state.search = {
      active: true,
      query: state.filterQuery,
      results: results
    }
    // 强制关闭其它视图，确保结果显示
    state.drilldown.active = false
  } finally {
    state.isScanning = false
  }
}

const closeSearch = () => {
  state.search.active = false
  state.filterQuery = ''
}

const closeDrilldown = () => {
  state.drilldown.active = false
}

const scrollToTop = () => {
  if (viewportRef.value) {
    viewportRef.value.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

const scrollToBottom = () => {
  if (viewportRef.value) {
    viewportRef.value.scrollTo({ top: viewportRef.value.scrollHeight, behavior: 'smooth' })
  }
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
      <div class="brand" style="margin-bottom: 24px;">
        <!-- v8.5 炫酷 SVG Logo -->
        <div class="logo-stack" style="width: 54px; height: 54px;">
          <svg viewBox="0 0 100 100" width="42" height="42" class="main-logo">
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
          <div class="disk-title-group" style="display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0;">
            <div class="radar-icon" style="flex-shrink: 0;"></div>
            <span style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis; font-size: 13px;">磁盘态势感知</span>
          </div>
          <span style="color: var(--neon-green)">{{ Math.round(state.disk.usage) || 0 }}%</span>
        </div>
        <div class="progress-container">
          <div class="progress-bar" :style="{ width: state.disk.usage + '%' }"></div>
        </div>
        <div class="disk-detail-grid" style="margin-top: 24px; display: flex; flex-direction: column; gap: 14px;">
          <div class="detail-item" style="display: flex; justify-content: space-between; align-items: center; gap: 8px;">
            <span class="detail-label" style="font-size: 11px; font-weight: 800; color: var(--neon-green); opacity: 0.9; white-space: nowrap;">已使用空间</span>
            <span class="detail-value used-val" style="font-size: 15px; font-weight: 800; color: var(--neon-green); white-space: nowrap;">{{ state.disk.used || '0 B' }}</span>
          </div>
          <div class="detail-item" style="display: flex; justify-content: space-between; align-items: center; gap: 8px;">
            <span class="detail-label" style="font-size: 11px; font-weight: 800; color: #3b82f6; opacity: 0.9; white-space: nowrap;">可用剩余</span>
            <span class="detail-value free-val" style="font-size: 15px; font-weight: 800; color: #3b82f6; white-space: nowrap;">{{ state.disk.free || '0 B' }}</span>
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
      <header class="top-bar" @dblclick="scrollToTop">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
          <input v-model="state.filterQuery" 
                 :placeholder="state.search.active ? '搜索中...' : '输入关键词，回车全局搜索资产...'" 
                 @keyup.enter="handleGlobalSearch" />
        </div>
        <div v-if="state.monitoredPath" class="path-chip" :class="{ restored: state.restored }">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path :d="icons.folder"/></svg>
          <span :data-fulltext="state.monitoredPath" style="max-width: 250px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ state.monitoredPath }}</span>
          <span v-if="state.restored" style="font-size: 10px; color: var(--neon-green); margin-left: 8px;">(已自动恢复)</span>
        </div>
      </header>

      <section class="viewport" ref="viewportRef">
        <!-- 底部快速置底感应器 -->
        <div class="bottom-navigator" @dblclick="scrollToBottom"></div>
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
        <div v-if="state.activeModule === 'activity' && !state.drilldown.active && !state.search.active" class="view-module">
          <!-- 实时监控呼吸指示器 (v8.3) -->
          <div class="monitoring-status">
            <div class="pulse-dot"></div>
            <span class="status-label">ENGINE ACTIVE</span>
            <span class="status-text">实时防御保护中</span>
          </div>

          <div class="stats-panel glass-card" style="display: flex; align-items: center; justify-content: space-between; padding: clamp(12px, 2vh, 24px) clamp(16px, 4vw, 60px); margin-bottom: clamp(16px, 3vh, 32px); background: rgba(0,0,0,0.3); border: 1px solid rgba(255,255,255,0.06); border-radius: 18px; position: relative; overflow: hidden;">
            <!-- 背景微弱扫掠光 -->
            <div style="position: absolute; top: 0; left: 0; right: 0; height: 1px; background: linear-gradient(90deg, transparent, rgba(57, 255, 20, 0.2), transparent); animation: scan 4s linear infinite;"></div>
            <div class="mini-stat" style="flex: 1; display: flex; justify-content: center;"><span class="dot create"></span><span style="font-weight: 800; font-size: clamp(11px, 1.2vw, 13px); letter-spacing: 0.5px; white-space: nowrap;">新建: {{ state.stats.create }}</span></div>
            <div class="orange-divider"></div>
            <div class="mini-stat" style="flex: 1; display: flex; justify-content: center;"><span class="dot write"></span><span style="font-weight: 800; font-size: clamp(11px, 1.2vw, 13px); letter-spacing: 0.5px; white-space: nowrap;">修改: {{ state.stats.write }}</span></div>
            <div class="orange-divider"></div>
            <div class="mini-stat" style="flex: 1; display: flex; justify-content: center;"><span class="dot remove"></span><span style="font-weight: 800; font-size: clamp(11px, 1.2vw, 13px); letter-spacing: 0.5px; white-space: nowrap;">删除: {{ state.stats.remove }}</span></div>
            <div class="orange-divider"></div>
            <div class="mini-stat" style="flex: 1; display: flex; justify-content: center;"><span class="dot folder"></span><span style="font-weight: 800; font-size: clamp(11px, 1.2vw, 13px); letter-spacing: 0.5px; white-space: nowrap;">目录: {{ state.stats.folders }}</span></div>
          </div>

          <!-- v11.3 旗舰级响应式看板 (Fluid Visuality) -->
          <div class="live-chart-container glass-card" style="margin-bottom: clamp(20px, 4vh, 36px); padding: clamp(24px, 4vw, 48px); display: flex; flex-direction: column; gap: clamp(16px, 3vh, 32px); background: rgba(0,0,0,0.2); border: 1px solid rgba(255,255,255,0.04); border-radius: 24px; position: relative; overflow: hidden;">
            <!-- 度量网格背景 -->
            <div class="modern-grid-overlay"></div>
            
            <div style="display: flex; justify-content: center; position: relative; z-index: 1;">
              <div style="font-size: clamp(9px, 1vw, 11px); color: #ff9800; font-weight: 900; text-transform: uppercase; letter-spacing: clamp(2px, 0.5vw, 6px); opacity: 1; text-shadow: 0 0 15px rgba(255,152,0,0.4); white-space: nowrap;">Live Dynamic Pulse Matrix</div>
            </div>
            
            <div class="chart-stage" style="display: flex; align-items: flex-end; justify-content: center; gap: clamp(20px, 5vw, 80px); height: clamp(120px, 15vh, 160px); position: relative; z-index: 1; padding: 0 10px;">
              <div v-for="s in [
                {key:'create', label:'CREATE', color:'#4ade80'},
                {key:'write', label:'UPDATE', color:'#60a5fa'},
                {key:'remove', label:'DELETE', color:'#f87171'},
                {key:'folders', label:'DIRECTORY', color:'#fbbf24'}
              ]" :key="s.key" class="chart-item" style="flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: center; gap: clamp(10px, 2vh, 18px);">
                <div class="bar-track-capsule" :style="{'--glow-color': s.color, width: 'clamp(20px, 3vw, 32px)', height: '100%'}">
                  <div class="glow-support"></div>
                  <div class="bar-fill" 
                       :style="{ 
                         height: Math.min(100, (state.stats[s.key] * 6 + 2)) + '%', 
                         background: `linear-gradient(to top, ${s.color}ee, ${s.color}77)`,
                         boxShadow: `0 0 25px ${s.color}55, inset 0 0 12px rgba(255,255,255,0.3)`
                       }"
                       style="width: 100%; position: absolute; bottom: 0; transition: height 1s cubic-bezier(0.19, 1, 0.22, 1); border-radius: 16px;">
                    <div style="height: 12px; width: 100%; background: rgba(255,255,255,0.4); border-radius: 16px 16px 0 0; filter: blur(2px);"></div>
                  </div>
                </div>
                <div style="display: flex; flex-direction: column; align-items: center; white-space: nowrap;">
                  <span style="font-size: clamp(7px, 0.8vw, 9px); font-weight: 900; color: #fff; opacity: 0.6; letter-spacing: 1px;">{{ s.label }}</span>
                  <span style="font-size: clamp(14px, 2vw, 24px); font-weight: 950; color: #fff; margin-top: 2px; font-family: 'JetBrains Mono', monospace;">{{ state.stats[s.key] }}</span>
                </div>
              </div>
            </div>
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
        <div v-if="state.activeModule === 'insights' && !state.drilldown.active && !state.search.active" class="view-module">
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
            <div class="legend-grid">
              <div v-for="seg in insightSegments" :key="seg.type" class="legend-module" @click="handleDrilldown(seg)">
                <div class="leg-icon-box" :style="{ color: seg.color }">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5"><path :d="seg.icon"/></svg>
                </div>
                <div class="leg-content">
                  <div class="leg-label">{{ seg.label }}</div>
                  <div class="leg-val" :style="{ color: seg.color }">{{ seg.val }}</div>
                </div>
              </div>
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
        <div v-if="state.activeModule === 'cleanup' && !state.drilldown.active && !state.search.active" class="view-module">
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

        <div v-if="state.search.active" class="view-module search-view">
          <div class="search-result-header" style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;">
            <div style="display: flex; align-items: center; gap: 24px;">
              <div class="back-action-chip" title="退出检索" @click="closeSearch">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
              </div>
              <div class="search-title-box">
                <!-- v10.0 再次延长撞色分割线 -->
                <div style="width: 120px; height: 3px; background: var(--neon-green); margin-bottom: 8px; border-radius: 2px; box-shadow: 0 0 12px rgba(57, 255, 20, 0.5);"></div>
                <!-- v9.9 暖橙色英文 -->
                <div style="font-size: 11px; color: #ff9800; opacity: 1; text-transform: uppercase; letter-spacing: 2px; margin-bottom: 6px; font-weight: 800; text-shadow: 0 0 8px rgba(255, 152, 0, 0.2);">Search Analytics Engine</div>
                <h3 style="margin: 0; font-size: 26px; font-weight: 800; color: #fff;">检索：{{ state.search.query }}</h3>
              </div>
            </div>
            
            <div style="display: flex; align-items: center; gap: 20px;">
              <!-- v10.0 多维排序中枢 -->
              <div class="sort-control-group" style="display: flex; background: rgba(255,255,255,0.04); padding: 4px; border-radius: 14px; border: 1px solid rgba(255,255,255,0.08); backdrop-filter: blur(10px);">
                <button v-for="s in [
                          {key:'size', label:'大小'}, 
                          {key:'time', label:'时间'}, 
                          {key:'name', label:'名称'}
                        ]" 
                        :key="s.key" 
                        class="sort-tab" 
                        :class="{ active: state.search.sortBy === s.key }"
                        style="padding: 8px 16px; border: none; background: transparent; color: #fff; font-size: 12px; font-weight: 700; cursor: pointer; border-radius: 10px; transition: all 0.3s; display: flex; align-items: center; gap: 6px;"
                        @click="toggleSort(s.key, 'search')">
                  {{ s.label }}
                  <span v-if="state.search.sortBy === s.key" style="font-size: 10px; opacity: 0.8;">
                    {{ state.search.sortOrder === 'desc' ? '▼' : '▲' }}
                  </span>
                </button>
              </div>

              <div class="search-stats-chip" style="display: flex; gap: 12px;">
                <div class="mini-stat" style="background: rgba(57, 255, 20, 0.05); border: 1px solid rgba(57, 255, 20, 0.2); height: 44px; padding: 0 16px; border-radius: 12px; display: flex; align-items: center; justify-content: center;">
                  <span style="opacity: 0.6; font-size: 10px; margin-right: 8px;">关联目录</span>
                  <span style="color: var(--neon-green); font-weight: 800; font-size: 14px;">{{ searchStats.dirs }}</span>
                </div>
                <div class="mini-stat" style="background: rgba(96, 165, 250, 0.05); border: 1px solid rgba(96, 165, 250, 0.2); height: 44px; padding: 0 16px; border-radius: 12px; display: flex; align-items: center; justify-content: center;">
                  <span style="opacity: 0.6; font-size: 10px; margin-right: 8px;">匹配文件</span>
                  <span style="color: #60a5fa; font-weight: 800; font-size: 14px;">{{ searchStats.files }}</span>
                </div>
              </div>
            </div>
          </div>
          
          <div v-if="state.search.results.length === 0" class="welcome-view" style="margin-top: 10vh;">
            未找到匹配的资产项目。试着更换关键词或目标路径。
          </div>
          
          <div v-else class="ranking-list" style="margin-top: 10px;">
            <div v-for="file in sortedSearchResults" :key="file.path" 
                 class="glass-card" 
                 style="padding: 16px 20px; cursor: pointer;"
                 @click="LocateFile(file.path)">
              <div class="card-indicator" style="background: var(--neon-green)"></div>
              <div class="card-body" style="padding: 12px 16px; overflow: hidden;">
                <div class="card-header">
                  <span class="op-tag" style="color: var(--neon-green)">SEARCH RESULT</span>
                  <span class="timestamp">最后修改: {{ file.timeDetail }}</span>
                </div>
                <div class="file-name-text" :data-fulltext="file.name" style="font-weight: 800; font-size: 14px; margin-top: 4px;">{{ file.name }}</div>
                <div class="file-path-text" :data-fulltext="file.path" style="font-size: 10px; opacity: 0.4; margin-top: 2px;">{{ file.path }}</div>
              </div>
              <div class="file-size-tag" style="margin-right: 16px; min-width: 90px; text-align: center;">{{ file.size }}</div>
            </div>
          </div>
        </div>

        <!-- 模块：安全审计 -->
        <div v-if="state.activeModule === 'security' && !state.drilldown.active && !state.search.active" class="view-module">
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
                <div class="mini-stat" style="background: rgba(96, 165, 250, 0.05); border: 1px solid rgba(96, 165, 250, 0.2); height: 44px; padding: 0 16px; border-radius: 14px; display: flex; align-items: center; justify-content: center;">
                  <span style="opacity: 0.6; font-size: 10px; margin-right: 8px;">分类资产</span>
                  <span style="color: #60a5fa; font-weight: 800; font-size: 14px;">{{ state.drilldown.files.length }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="state.drilldown.active && !state.search.active" class="view-module drilldown-view">
          <div class="search-result-header" style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px;">
            <div style="display: flex; align-items: center; gap: 24px;">
              <div class="back-action-chip" title="返回概览" @click="closeDrilldown">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
              </div>
              <div class="search-title-box">
                <!-- v10.1 镜像视觉：120px 绿条 -->
                <div style="width: 120px; height: 3px; background: var(--neon-green); margin-bottom: 8px; border-radius: 2px; box-shadow: 0 0 12px rgba(57, 255, 20, 0.5);"></div>
                <!-- v10.1 镜像视觉：暖橙色英文 -->
                <div style="font-size: 11px; color: #ff9800; opacity: 1; text-transform: uppercase; letter-spacing: 2px; margin-bottom: 6px; font-weight: 800; text-shadow: 0 0 8px rgba(255, 152, 0, 0.2);">Classification Insights Engine</div>
                <h3 style="margin: 0; font-size: 26px; font-weight: 800; color: #fff;">分类：{{ state.drilldown.name }}</h3>
              </div>
            </div>

            <div style="display: flex; align-items: center; gap: 20px;">
              <!-- v10.1 对标排序中枢 -->
              <div class="sort-control-group" style="display: flex; background: rgba(255,255,255,0.04); padding: 4px; border-radius: 14px; border: 1px solid rgba(255,255,255,0.08); backdrop-filter: blur(10px); height: 44px; align-items: center;">
                <button v-for="s in [
                          {key:'size', label:'大小'}, 
                          {key:'time', label:'时间'}, 
                          {key:'name', label:'名称'}
                        ]" 
                        :key="s.key" 
                        class="sort-tab" 
                        :class="{ active: state.drilldown.sortBy === s.key }"
                        style="height: 36px; padding: 0 16px; border: none; background: transparent; color: #fff; font-size: 12px; font-weight: 700; cursor: pointer; border-radius: 10px; transition: all 0.3s; display: flex; align-items: center; gap: 6px;"
                        @click="toggleSort(s.key, 'drilldown')">
                  {{ s.label }}
                  <span v-if="state.drilldown.sortBy === s.key" style="font-size: 10px; opacity: 0.8;">
                    {{ state.drilldown.sortOrder === 'desc' ? '▼' : '▲' }}
                  </span>
                </button>
              </div>

              <div class="search-stats-chip" style="display: flex; gap: 12px;">
                <div class="mini-stat" style="background: rgba(96, 165, 250, 0.05); border: 1px solid rgba(96, 165, 250, 0.2); height: 44px; padding: 0 16px; border-radius: 14px; display: flex; align-items: center; justify-content: center;">
                  <span style="opacity: 0.6; font-size: 10px; margin-right: 8px;">分类资产</span>
                  <span style="color: #60a5fa; font-weight: 800; font-size: 14px;">{{ state.drilldown.files.length }}</span>
                </div>
              </div>
            </div>
          </div>
          
          <div class="ranking-list" style="margin-top: 24px;">
            <div v-for="file in sortedDrilldownFiles" :key="file.path" 
                 class="glass-card" 
                 style="padding: 16px 20px; cursor: pointer;"
                 @click="LocateFile(file.path)">
              <div class="card-indicator" style="background: var(--neon-green)"></div>
              <div class="card-body" style="padding: 12px 16px; overflow: hidden;">
                <div class="card-header">
                  <span class="op-tag" style="color: var(--neon-green)">FILE ASSET</span>
                  <span class="timestamp">最后修改: {{ file.timeDetail }}</span>
                </div>
                <div class="file-name-text" :data-fulltext="file.name" style="font-weight: 800; font-size: 14px; margin-top: 4px;">{{ file.name }}</div>
                <div class="file-path-text" :data-fulltext="file.path" style="font-size: 10px; opacity: 0.4; margin-top: 2px;">{{ file.path }}</div>
              </div>
              <div class="file-size-tag" style="margin-right: 16px; min-width: 90px; text-align: center;">{{ file.size }}</div>
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
.sidebar { width: var(--sidebar-width); background: rgba(0, 0, 0, 0.3); border-right: 1px solid var(--glass-border); display: flex; flex-direction: column; padding: clamp(24px, 4vh, 32px) clamp(12px, 1.5vw, 20px); backdrop-filter: blur(20px); flex-shrink: 0; transition: width 0.3s ease; }
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
.viewport { flex: 1; overflow-y: auto; }
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

.sort-tab.active {
  background: var(--neon-green) !important;
  color: #000 !important;
  box-shadow: 0 0 12px var(--neon-green);
}

.sort-tab:hover:not(.active) {
  background: rgba(255,255,255,0.1) !important;
}

/* 底部快速导航感应器 (v9.6) */
.bottom-navigator {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 48px;
  z-index: 5;
  cursor: pointer;
  background: linear-gradient(to top, rgba(57, 255, 20, 0.05), transparent);
  pointer-events: all;
  transition: background 0.3s;
}
.bottom-navigator:hover {
  background: linear-gradient(to top, rgba(57, 255, 20, 0.12), transparent);
}
</style>
