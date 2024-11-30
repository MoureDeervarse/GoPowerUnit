package unit

import (
	"context"
	"os"
	"os/exec"
	"sync"
)

type LocalUnit struct {
	projectPath string
	cmd         *exec.Cmd
	mutex       sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewLocalUnit(projectPath string) *LocalUnit {
	ctx, cancel := context.WithCancel(context.Background())
	return &LocalUnit{
		projectPath: projectPath,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (u *LocalUnit) Start() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.cmd != nil {
		return nil // 이미 실행 중
	}

	u.cmd = exec.CommandContext(u.ctx, "go", "run", ".")
	u.cmd.Dir = u.projectPath
	u.cmd.Stdout = os.Stdout
	u.cmd.Stderr = os.Stderr

	return u.cmd.Start()
}

func (u *LocalUnit) Stop() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.cmd == nil {
		return nil
	}

	u.cancel()
	err := u.cmd.Wait()
	u.cmd = nil
	return err
}

func (u *LocalUnit) Reload() error {
	if err := u.Stop(); err != nil {
		return err
	}
	return u.Start()
}
