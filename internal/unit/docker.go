package unit

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mouredeervarse/go-power-unit/internal/config"
)

type DockerUnit struct {
	config         *config.Config
	dockerfilePath string
	containerName  string
	stopChan       chan os.Signal
}

func NewDockerUnit(cfg *config.Config) (*DockerUnit, error) {
	absPath, err := filepath.Abs(cfg.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to absolute path: %v", err)
	}
	cfg.ProjectPath = absPath

	projectName := filepath.Base(absPath)

	d := &DockerUnit{
		config:        cfg,
		containerName: projectName,
		stopChan:      make(chan os.Signal, 1),
	}

	dockerfilePath, err := d.findDockerfile()
	if err != nil {
		return nil, err
	}
	d.dockerfilePath = dockerfilePath
	d.config.DockerfilePath = dockerfilePath

	if d.config.DockerImage == "" {
		d.config.DockerImage = projectName + ":latest"
	}

	signal.Notify(d.stopChan, os.Interrupt, syscall.SIGTERM)
	go d.handleSignals()

	return d, nil
}

func (d *DockerUnit) findDockerfile() (string, error) {
	dockerfilePath := filepath.Join(d.config.ProjectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err != nil {
		return "", fmt.Errorf("dockerfile not found")
	}

	if err := d.extractExposedPorts(dockerfilePath); err != nil {
		return "", fmt.Errorf("failed to extract ports: %v", err)
	}

	return dockerfilePath, nil
}

func (d *DockerUnit) extractExposedPorts(dockerfilePath string) error {
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ToUpper(line))
		if strings.HasPrefix(trimmed, "EXPOSE") {
			ports := strings.Fields(trimmed)[1:]
			for _, port := range ports {
				port = strings.Split(port, "/")[0]
				portMapping := fmt.Sprintf("%s:%s", port, port)
				d.config.Ports = append(d.config.Ports, portMapping)
			}
		}
	}

	return nil
}

func (d *DockerUnit) Build() error {
	if d.dockerfilePath == "" {
		return fmt.Errorf("dockerfile path not set")
	}

	imageTag := d.config.DockerImage
	if imageTag == "" {
		imageTag = "app:latest"
	}

	cmd := exec.Command("docker", "build", "-t", imageTag, "-f", d.dockerfilePath, d.config.ProjectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	return nil
}

func (d *DockerUnit) removeContainer() error {
	cmd := exec.Command("docker", "rm", "-f", d.containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *DockerUnit) runContainer() error {
	imageTag := d.config.DockerImage
	if imageTag == "" {
		imageTag = "app:latest"
	}

	args := []string{"run",
		"-d",
		"--name", d.containerName,
		"--rm"}

	for _, port := range d.config.Ports {
		args = append(args, "-p", port)
	}

	args = append(args, imageTag)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	return nil
}

func (d *DockerUnit) OnChange() error {
	// 1. remove existing container
	if err := d.removeContainer(); err != nil {
		log.Printf("warning: failed to remove container: %v", err)
		// remove container failure is not critical, so continue
	}

	// 2. build new image
	if err := d.Build(); err != nil {
		return err
	}

	// 3. run new container
	if err := d.runContainer(); err != nil {
		return err
	}

	return nil
}

func (d *DockerUnit) handleSignals() {
	<-d.stopChan
	log.Println("received stop signal. cleaning up container...")

	if err := d.removeContainer(); err != nil {
		log.Printf("remove container failed: %v", err)
	}

	os.Exit(0)
}
