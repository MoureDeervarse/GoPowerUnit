package unit

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
)

type DockerUnit struct {
	projectPath   string
	imageName     string
	containerName string
	mutex         sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewDockerUnit(projectPath, imageName, containerName string) *DockerUnit {
	ctx, cancel := context.WithCancel(context.Background())
	return &DockerUnit{
		projectPath:   projectPath,
		imageName:     imageName,
		containerName: containerName,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (u *DockerUnit) Start() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// 이미지 빌드
	buildCmd := exec.CommandContext(u.ctx, "docker", "build", "-t", u.imageName, u.projectPath)
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// 기존 컨테이너 제거
	stopCmd := exec.CommandContext(u.ctx, "docker", "rm", "-f", u.containerName)
	stopCmd.Run() // 에러 무시 (컨테이너가 없을 수 있음)

	// 새 컨테이너 실행
	runCmd := exec.CommandContext(u.ctx, "docker", "run",
		"-d",
		"--name", u.containerName,
		u.imageName,
	)
	return runCmd.Run()
}

func (u *DockerUnit) Stop() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	stopCmd := exec.CommandContext(u.ctx, "docker", "rm", "-f", u.containerName)
	return stopCmd.Run()
}

func (u *DockerUnit) Reload() error {
	if err := u.Stop(); err != nil {
		return err
	}
	return u.Start()
}
