package watcher

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mouredeervarse/go-power-unit/internal/config"
)

type Watcher struct {
	cfg     *config.Config
	watcher *fsnotify.Watcher
	timer   *time.Timer
}

func New(cfg *config.Config) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &Watcher{
		cfg:     cfg,
		watcher: w,
	}
}

func (w *Watcher) Watch(reloadFunc func() error) {
	defer w.watcher.Close()

	for _, path := range w.cfg.WatchPaths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return w.watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	const cooldown = 500 * time.Millisecond
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && filepath.Ext(event.Name) == ".go" {
					if w.timer != nil {
						w.timer.Stop()
					}
					w.timer = time.AfterFunc(cooldown, func() {
						if err := reloadFunc(); err != nil {
							log.Println("reload error:", err)
						}
					})
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	<-done
}
