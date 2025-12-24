# 设计方案：UI 视觉美化

## 架构决策

### 1. 多彩粉蓝渐变 (Aesthetic Gradient)

- **方案**：在 `#app::before` 层使用 `linear-gradient` 或 `radial-gradient`。
- **动效**：通过 `background-size: 200% 200%` 并结合 CSS 动画 (`keyframe`) 实现背景色的缓慢流动感。

### 2. 长文件名滚动 (Scrolling Filenames)

- **方案**：使用 `overflow: hidden` 与 CSS `@keyframes scrollText`。
- **触发机制**：仅在文件名长度超过容器宽度时启用。或者简单实现为一个固定宽度的容器，内容在 Hover 或自动状态下左右平移。

### 3. 深层模糊 (Heavy Backdrop Blur)

- **方案**：将 `backdrop-filter: blur(20px)` 提升至 `40px` - `60px`，并配合 `background-color` 的透明度调整（降低至 `0.1` - `0.2`），使光影穿透效果更明显。

## 性能考量

深度的 `backdrop-filter` 在部分低端设备上可能导致 GPU 压力，需确保在主流配置下流畅。
