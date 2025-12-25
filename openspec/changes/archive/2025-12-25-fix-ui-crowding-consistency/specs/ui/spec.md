# Spec Delta: UI 统合与精致化 (v8.2)

## MODIFIED Requirements

### Requirement: 统一玻璃材质规范

#### Scenario: 跨模块卡片风格一致

- **Given**: 排行榜项、清理建议项及事件流卡片。
- **Then**: 必须使用统一的背景色 `rgba(255,255,255,0.03)` 与边框规范。
- **Then**: 卡片必须支持 Hover 时的色值加深与位移动效，以增强交互反馈。

### Requirement: 响应式按钮组布局

#### Scenario: 处理拥挤的操作项

- **Given**: 清理建议模块的操作按钮。
- **Then**: 按钮组 SHALL 采用 Flex 布局且具 `12px` 间距。
- **Then**: 窗口缩窄时，按钮组必须自动换行，且行间距保持一致。

### Requirement: 信息架构优先级

#### Scenario: 排行榜单列展示

- **Given**: 占用排行榜 (Top 20) 内容。
- **Then**: 为平衡长路径显示，排行榜 SHALL 采用单一垂直列布局而非多列网格。
- **Then**: 各项之间 SHALL 保持标准间隙以消除视觉压迫感。
