# 磁盘管理与安全中心 (Monitor Pro)

一款基于 Wails + Vue 3 构建的高性能、高颜值 Windows 磁盘监控与资产管理工具。

## 🌟 核心特性 (v7.1)

### 1. 🔍 多维空间洞察 (Disk Insights)

- **资产深度扫描**：全量枚举磁盘文件，提供分钟级的深度资产透视。
- **20+ 开发语言识别**：自动统计并分类 Go, Python, Rust, Java, C++, Vue 等主流代码资产。
- **占用排行榜**：实时呈现 Top 20 大文件，并支持毫秒级修改时间追踪。

### 2. ⚡ 实时活动流 (Live Activity)

- **高精度监听**：基于 `fsnotify` 实时捕获文件系统的新建、修改、删除及重命名操作。
- **毫秒级时戳**：所有变动精确对齐至 `HH:mm:ss.ms`，确保操作溯源无死角。

### 3. 🧹 智能清理建议 (Smart Cleanup)

- **冗余资产识别**：智能扫描系统临时文件、回收站及大面积重复数据。
- **一键清理**：安全管控，支持自定义选中后的快速物理释放。

### 4. 🛡️ 安全风险审计 (Security Watch)

- **敏感监控**：针对 `.config`, `.ssh`, `.env` 等敏感文件及系统核心目录的变动进行实时预警。
- **风险标记**：直观区分系统变动与潜在的权限敏感操作。

### 5. 🏗️ 全盘架构性能

- **驱动器直挂**：彻底优化递归挂载算法，支持直接监控整个磁盘分区（如 `C:\`, `D:\`）。
- **鲁棒扫描**：静默跳过系统受限目录（System Volume Information, Recycle等），挂载过程流畅无卡顿。

## 🚀 快速开始

### 开发环境

- Go 1.21+
- Node.js 18+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 运行调试

```bash
wails dev
```

### 生产构建

```bash
wails build
```

## 🛠️ 技术栈

- **后端**: Go + [Wails](https://wails.io/) + [fsnotify](https://github.com/fsnotify/fsnotify)
- **前端**: Vue 3 + Vite + Vanilla CSS (Glassmorphism Design)

---
© 2025 Monitor Pro 团队. 保留所有权利。
