# 任务：文件监控系统 (Go + Wails)

## 环境搭建

- [ ] 检查 Go 版本并安装 Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) <!-- id: 10 -->
- [ ] 初始化 Wails 项目项目 <!-- id: 11 -->

## 后端开发

- [ ] 实现 `fsnotify` 监听循环 <!-- id: 20 -->
- [ ] 定义事件结构体并完成 Wails 绑定 <!-- id: 21 -->

## 前端开发

- [ ] 在 `main.go` 中配置 Mica 背景背景 <!-- id: 30 -->
- [ ] 构建基于 CSS 模糊效果的日志展示界面 <!-- id: 31 -->

## 验证

- [ ] 在 Windows 11 上编译并运行，确认 Mica 效果渲染 <!-- id: 40 -->
- [ ] 验证多目录监控的准确性 <!-- id: 41 -->
