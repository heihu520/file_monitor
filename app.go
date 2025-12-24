package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DiskInfo 磁盘空间信息
type DiskInfo struct {
	Total string  `json:"total"`
	Free  string  `json:"free"`
	Used  string  `json:"used"`
	Usage float64 `json:"usage"`
}

// DirInsight 目录深度统计
type DirInsight struct {
	TotalSize  string         `json:"totalSize"`
	TotalBytes int64          `json:"totalBytes"`
	FileCount  int            `json:"fileCount"`
	DirCount   int            `json:"dirCount"`
	Categories map[string]int `json:"categories"`
	ExtDetails map[string]int `json:"extDetails"` // 新增：扩展名明细统计
}

// FileStat 用于排序的大文件结构
type FileStat struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       string `json:"size"`
	Bytes      int64  `json:"bytes"`
	TimeDetail string `json:"timeDetail"` // 新增：毫秒级时间字符串
}

// App struct
type App struct {
	ctx            context.Context
	watcher        *fsnotify.Watcher
	securityEvents []map[string]interface{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		securityEvents: make([]map[string]interface{}, 0),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	a.watcher = watcher

	// Start a goroutine to listen for events
	go a.listenForEvents()
}

// addRecursive 递归添加目录监听（增强：跳过系统受限目录与无权限目录）
func (a *App) addRecursive(path string) error {
	skipDirs := map[string]bool{
		"System Volume Information": true,
		"$RECYCLE.BIN":              true,
		"Recovery":                  true,
		"Windows":                   true,
	}

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 遇到权限错误静默跳过，继续扫描其它
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			// 尝试添加监听，忽略单目录失败
			_ = a.watcher.Add(walkPath)
		}
		return nil
	})
	return err
}

// SelectFolder opens a folder dialog and starts monitoring it recursively
func (a *App) SelectFolder() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要监控的文件夹",
	})
	if err != nil {
		return "", err
	}

	if path != "" {
		err = a.addRecursive(path)
		if err != nil {
			return "", err
		}
		return path, nil
	}
	return "", nil
}

// GetDiskInfo 获取指定路径所属驱动器的磁盘信息
func (a *App) GetDiskInfo(path string) (*DiskInfo, error) {
	if path == "" {
		path = "C:\\"
	}
	root := filepath.VolumeName(path) + "\\"

	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytes, totalBytes, availBytes int64
	_, _, _ = c.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(root))),
		uintptr(unsafe.Pointer(&availBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&freeBytes)),
	)

	usedBytes := totalBytes - freeBytes
	usage := 0.0
	if totalBytes > 0 {
		usage = float64(usedBytes) / float64(totalBytes) * 100
	}

	return &DiskInfo{
		Total: a.formatSize(totalBytes),
		Free:  a.formatSize(freeBytes),
		Used:  a.formatSize(usedBytes),
		Usage: usage,
	}, nil
}

// GetDirectoryInsight 深度扫描目录（增强版：支持明细统计与进度推送）
func (a *App) GetDirectoryInsight(path string) (*DirInsight, error) {
	insight := &DirInsight{
		Categories: make(map[string]int),
		ExtDetails: make(map[string]int),
	}

	// 排除 Windows 受保护及挂载敏感目录
	skipDirs := map[string]bool{
		"System Volume Information": true,
		"$RECYCLE.BIN":              true,
		"Recovery":                  true,
		"Windows":                   true, // 跳过系统目录提速
	}

	totalScanned := 0
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			insight.DirCount++
		} else {
			insight.FileCount++
			insight.TotalBytes += info.Size()

			ext := strings.ToLower(filepath.Ext(walkPath))
			if ext == "" {
				ext = "其他"
			}
			insight.Categories[ext]++
			insight.ExtDetails[ext]++ // 明细统计
		}

		totalScanned++
		if totalScanned%100 == 0 {
			runtime.EventsEmit(a.ctx, "scan-progress", map[string]interface{}{
				"scanned": totalScanned,
				"current": filepath.Base(walkPath),
			})
		}
		return nil
	})

	insight.TotalSize = a.formatSize(insight.TotalBytes)
	return insight, err
}

// LocateFile 在资源管理器中定位并选中文件
func (a *App) LocateFile(path string) error {
	return exec.Command("explorer", "/select,", path).Run()
}

// GetTopFiles 获取受控目录下最大的20个文件
func (a *App) GetTopFiles(path string) ([]FileStat, error) {
	files := make([]FileStat, 0)
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		files = append(files, FileStat{
			Name:       info.Name(),
			Path:       walkPath,
			Size:       a.formatSize(info.Size()),
			Bytes:      info.Size(),
			TimeDetail: info.ModTime().Format("15:04:05.000"),
		})
		return nil
	})

	// 按大小降序排序
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].Bytes < files[j].Bytes {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	if len(files) > 20 {
		return files[:20], err
	}
	return files, err
}

// ScanCleanup 扫描可清理的冗余文件
func (a *App) ScanCleanup(path string) ([]FileStat, error) {
	redundant := make([]FileStat, 0)
	targetExts := map[string]bool{".tmp": true, ".log": true, ".bak": true, ".cache": true}

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(walkPath))
		if targetExts[ext] || strings.Contains(strings.ToLower(walkPath), "cache") {
			redundant = append(redundant, FileStat{
				Name:  info.Name(),
				Path:  walkPath,
				Size:  a.formatSize(info.Size()),
				Bytes: info.Size(),
			})
		}
		return nil
	})
	return redundant, err
}

// ExecuteCleanup 执行物理删除
func (a *App) ExecuteCleanup(paths []string) error {
	for _, p := range paths {
		_ = os.Remove(p)
	}
	return nil
}

// GetSecurityAudit 获取安全审计记录
func (a *App) GetSecurityAudit() []map[string]interface{} {
	return a.securityEvents
}

func (a *App) listenForEvents() {
	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}

			isDir := false
			info, err := os.Stat(event.Name)
			modTime := ""
			if err == nil {
				isDir = info.IsDir()
				modTime = info.ModTime().Format("15:04:05.000") // 毫秒级时戳
				if isDir && event.Op&fsnotify.Create == fsnotify.Create {
					a.addRecursive(event.Name)
				}
			}

			// 安全监控：记录敏感操作
			isSensitive := false
			ext := strings.ToLower(filepath.Ext(event.Name))
			if ext == ".exe" || ext == ".bat" || ext == ".ps1" || ext == ".cmd" {
				isSensitive = true
			}

			evData := map[string]interface{}{
				"name":        event.Name,
				"op":          event.Op.String(),
				"isDir":       isDir,
				"isSensitive": isSensitive,
				"time":        modTime, // 真实毫秒级时间
			}

			if isSensitive || event.Op&fsnotify.Remove == fsnotify.Remove {
				a.securityEvents = append(a.securityEvents, evData)
				if len(a.securityEvents) > 100 {
					a.securityEvents = a.securityEvents[1:]
				}
			}

			runtime.EventsEmit(a.ctx, "file-event", evData)
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

// formatSize 格式化字节数为人类可读格式
func (a *App) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
