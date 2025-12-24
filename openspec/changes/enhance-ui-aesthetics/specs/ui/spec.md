# Spec Delta: UI 视觉增强

## MODIFIED Requirements

### Requirement: 全局背景系统增强

#### Scenario: 启用多彩流动渐变

- **Given**: 用户启动应用或在任何模块中。
- **Then**: 背景应显示粉色 (`#ff7eb3`) 与天蓝色 (`#7afcff`) 的混合渐变。
- **Then**: 渐变色应以 10-20s 的周期产生平滑流动感。

### Requirement: 扫描指示器交互优化

#### Scenario: 自动滚动长文件名

- **Given**: 文件名长度超过 `scan-text` 容器剩余空间。
- **Then**: 文案应自动执行水平左右往复滚动，确保完整路径可见。

### Requirement: 表面材质升级

#### Scenario: 加重背景模糊

- **Given**: 所有的 `glass-card` 或 `sidebar` 元素。
- **Then**: 其 `backdrop-filter: blur(...)` 参数应设为 `50px`。
- **Then**: 遮罩颜色透明度调至物理感知最舒适的阈值（推荐 0.15）。
