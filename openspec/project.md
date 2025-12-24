# 项目上下文

## Purpose

实现一个具有 Win11 原生高斯模糊（Mica/Acrylic）效果的高性能文件监控系统。

## 技术栈

- **Backend**: Go 1.21+ (Wails v2)
- **Frontend**: Vue 3 (Vite), Vanilla CSS
- **Win11 Integration**: Wails builtin (Windows/Mac/Linux support), CSS Backdrop Filter
- **File Monitoring**: `fsnotify` (Go native)

## 项目约定

### 代码风格

- PEP 8 命名规范。
- 界面逻辑与业务逻辑分离。

### 架构模式

- MVC (Model-View-Controller)
- 异步事件处理，避免界面卡顿。

### 测试策略

[解释您的测试方法和要求]

### Git 工作流

[描述您的分支策略和提交约定]

## 领域上下文

[添加AI助手需要理解的领域特定知识]

## 重要约束

[列出任何技术、业务或法规约束]

## 外部依赖

[记录关键的外部服务、API或系统]
