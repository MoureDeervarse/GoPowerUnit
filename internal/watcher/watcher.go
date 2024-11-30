package watcher

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/mouredeervarse/go-power-unit/internal/config"
)

type Watcher struct {
	cfg     *config.Config
	watcher *fsnotify.Watcher
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
				// TODO: Check if path should be ignored
				return w.watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := reloadFunc(); err != nil {
						log.Println("reload error:", err)
					}
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
