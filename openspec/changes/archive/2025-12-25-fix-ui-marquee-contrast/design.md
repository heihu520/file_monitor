# 设计方案：UI 体验修复

## 1. 跑马灯逻辑重构 (Unified Marquee)

- **问题**：当前的 `::after` 方案依赖于 `content: attr(...)`，在某些布局下会导致宽度计算偏差，且未覆盖所有模块。
- **改进**：
  - 在 `App.vue` 中为所有潜在的长文本（实时活动文件名、排行榜路径、排行榜文件名）统一注入 `data-fulltext` 属性。
  - 在 `style.css` 中定义 `.marquee-on-hover` 类。
  - 使用更稳定的 `mask-image` 实现边缘渐隐。
  - 调整 `translateX(-100%)` 或动态偏移，确保文件名完全可见。

## 2. 顶栏对比度增强 (Contrast Optimization)

- **遮罩层**：将 `.top-bar` 的背景由 `rgba(0, 0, 0, 0.1)` 提升至 `rgba(0, 0, 0, 0.4)`，并增强 `backdrop-filter` 密度。
- **文字阴影**：为 `.top-bar` 内部的文字、搜索图标及路径芯片添加轻微的 `text-shadow` (e.g., `0 1px 4px rgba(0,0,0,0.5)`)。
- **背景调整**：微调 `#app::before` 的渐变透明度，减益过亮的粉色区域。
