# Spec Delta: 响应式增强与实时监控指示 (v8.3)

## ADDED Requirements

### Requirement: 实时监控心跳指示器 (Heartbeat)

#### Scenario: 监控状态视觉反馈

- **Given**: 系统正处于文件监控模式。
- **When**: 用户进入“实时活动”模块。
- **Then**: 模块顶部 SHALL 显示一个包含呼吸脉冲圆点的指示器栏。
- **Then**: 指示圆点 SHALL 伴随 2 秒周期的透明度梯度与微缩放动效 (`pulse`)。
- **Then**: 指示文本应显示监控引擎的实时工作状态。

## MODIFIED Requirements

### Requirement: 全站流式自适应网格

#### Scenario: 跨分辨率布局自愈

- **Given**: 包含大量卡片的网格模块（排行榜、清理建议）。
- **When**: 视口宽度缩放。
- **Then**: 容器必须动态重算网格列数，而非强制单列或定宽双列。
- **Then**: 单体卡片宽度 SHALL 在 `400px` 至 `1fr` 之间自动流转，确保长路径文件名有足够的物理展示空间。

### Requirement: 精确原子间距规范

#### Scenario: 解决复选框与文字挤压

- **Given**: 带操作项的卡片（如清理列表项）。
- **Then**: 复选框与文本区之间 MUST 存在至少 `16px` 的缓冲带。
- **Then**: 文本展示区域 MUST 设置 `min-width: 0` 以保障 Flex-box 内的 `ellipsis` 或 `marquee` 逻辑生效，不得因长文本撑开卡片结构。
