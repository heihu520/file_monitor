# Spec Delta: UI 交互与对比度修复 (v8.1)

## MODIFIED Requirements

### Requirement: 全局文本滚动系统 (Marquee)

#### Scenario: 悬停展示完整文件名

- **Given**: 一个被赋予 `.marquee-on-hover` 类且具有 `data-fulltext` 属性的文本元素。
- **Given**: 该文本内容由于宽度限制被省略。
- **When**: 用户将鼠标悬停在该元素上。
- **Then**: 文本应平滑滚动至末尾并回弹或循环，确保 100% 内容可见。
- **Then**: 滚动应涵盖：实时活动流文件名、排行榜 Top 20 文件名、清理列表路径。

### Requirement: 顶栏对比度优化

#### Scenario: 增强文字清晰度

- **Given**: 应用顶部的 `.top-bar` 及其子组件。
- **Then**: 顶栏遮罩透明度 SHALL 提升。
- **Then**: 顶栏内文字与图标 SHALL 具备 `drop-shadow` 以抵消高亮渐变底色。
- **Then**: 搜索框背景 SHALL 加深至 `rgba(0,0,0,0.5)` 以形成视觉反差。
