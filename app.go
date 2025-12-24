package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	watcher *fsnotify.Watcher
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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

func (a *App) listenForEvents() {
	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}

			// 如果是创建了新文件夹，自动加入监听
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					a.addRecursive(event.Name)
				}
			}

			// 推送事件到前端
			runtime.EventsEmit(a.ctx, "file-event", map[string]string{
				"name": event.Name,
				"op":   event.Op.String(),
			})
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

// addRecursive adds a directory and all its sub-directories to the watcher
func (a *App) addRecursive(path string) error {
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return a.watcher.Add(walkPath)
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

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
