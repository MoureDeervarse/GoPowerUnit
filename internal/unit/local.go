package unit

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type LocalUnit struct {
	projectPath string
	mainPath    string
	cmd         *exec.Cmd
}

func NewLocalUnit(projectPath string) *LocalUnit {
	mainPath, err := findMainGo(projectPath)
	if err != nil {
		log.Fatalf("Failed to find main.go: %v", err)
	}
	return &LocalUnit{
		projectPath: projectPath,
		mainPath:    mainPath,
	}
}

func findMainGo(root string) (string, error) {
	var mainPath string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// stop walking directory when find main.go
		if !info.IsDir() && info.Name() == "main.go" {
			mainPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking directory: %v", err)
	}
	if mainPath == "" {
		return "", fmt.Errorf("main.go not found in %s", root)
	}
	return mainPath, nil
}

func (u *LocalUnit) Reload() error {
	// kill previous process if exists
	if u.cmd != nil && u.cmd.Process != nil {
		// kill process group with minus pgid
		pgid, err := syscall.Getpgid(u.cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		}

		// wait for port to be released
		time.Sleep(100 * time.Millisecond)

		// kill process if still running
		if err := u.cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill process: %v", err)
		}

		// wait for process to be terminated
		u.cmd.Wait()
	}

	// run main.go in the directory of main.go
	dir := filepath.Dir(u.mainPath)
	u.cmd = exec.Command("go", "run", u.mainPath)
	u.cmd.Dir = dir
	u.cmd.Stdout = os.Stdout
	u.cmd.Stderr = os.Stderr

	// create new process group
	u.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := u.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	fmt.Printf("Running %s...\n", u.mainPath)
	return nil
}

func (u *LocalUnit) Cleanup() error {
	if u.cmd == nil || u.cmd.Process == nil {
		return nil
	}

	// terminate process group
	pgid, err := syscall.Getpgid(u.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// wait for a moment
	time.Sleep(100 * time.Millisecond)

	// force kill if still running
	if err := u.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process during cleanup: %v", err)
	}

	return u.cmd.Wait()
}
