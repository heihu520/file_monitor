# 设计方案：UI 统合与精致化修复

## 1. 按钮组布局 (Button Grouping)

- **方案**：在 `style.css` 中重定义 `.button-group`。
- **配置**：使用 `display: flex; gap: 12px; flex-wrap: wrap; align-items: center;`。确保在窄屏下按钮能自动换行而不会相互遮挡。

## 2. 内容网格重构 (Content Grid Refactoring)

- **单列优先**：将“占用排行榜”由 `layout-grid` (多列) 调整为独立的 `.ranking-list` (100% 宽度)。
- **原因**：长路径和长文件名在多列布局下极易溢出。单列布局能为跑马灯和文字提供足够的物理空间力。

## 3. 玻璃材质统合 (Unified Glassmorphism)

- **统一规范**：所有模块卡片（事件、排行、清理）必须继承统一的 `.glass-card` 背景与模糊。
- **调整**：
  - `background: rgba(255, 255, 255, 0.03);`
  - `border: 1px solid rgba(255, 255, 255, 0.08);`
  - `backdrop-filter: blur(20px);` (局部卡片无需全局 50px 模糊，以平衡性能表现)。

## 4. 统计面板细节优化 (Stat Panel Polish)

- **对齐**：使用 `justify-content: space-between` 重整 `.disk-info-header`。
- **留白**：增加 `padding: 20px` 以确保内容不紧贴卡片边缘。
- **色彩识别**：为 `languageStats` 的标签点引入与分类对应的标准色。
