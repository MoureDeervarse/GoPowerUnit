package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mouredeervarse/go-power-unit/internal/config"
	"github.com/mouredeervarse/go-power-unit/internal/unit"
)

type Watcher struct {
	cfg     *config.Config
	watcher *fsnotify.Watcher
	timer   *time.Timer
	units   []*unit.DockerUnit
}

func New(cfg *config.Config) (*Watcher, error) {
	// fsnotify watcher 초기화
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	w := &Watcher{
		cfg:     cfg,
		watcher: fsWatcher,
	}

	// Docker unit initialization
	if cfg.DockerImage != "" {
		dockerUnit, err := unit.NewDockerUnit(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize docker unit: %v", err)
		}
		w.units = append(w.units, dockerUnit)
	}

	return w, nil
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
