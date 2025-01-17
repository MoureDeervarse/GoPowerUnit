package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

type DeployUnit struct {
	projectPath    string
	imageName      string
	registryURL    string
	deployEndpoint string
}

func NewDeployUnit(projectPath, imageName, registryURL, deployEndpoint string) *DeployUnit {
	return &DeployUnit{
		projectPath:    projectPath,
		imageName:      imageName,
		registryURL:    registryURL,
		deployEndpoint: deployEndpoint,
	}
}

func (u *DeployUnit) Start() error {
	// Build image
	buildCmd := exec.Command("docker", "build", "-t", u.imageName, u.projectPath)
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// Add registry tag
	registryImage := fmt.Sprintf("%s/%s", u.registryURL, u.imageName)
	tagCmd := exec.Command("docker", "tag", u.imageName, registryImage)
	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("docker tag failed: %w", err)
	}

	// Push to registry
	pushCmd := exec.Command("docker", "push", registryImage)
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	// Call deploy endpoint
	payload := map[string]string{
		"image": registryImage,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(u.deployEndpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deploy request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (u *DeployUnit) Stop() error {
	// Stop service is managed by deployment system
	return nil
}

func (u *DeployUnit) Reload() error {
	return u.Start()
}

func (u *DeployUnit) Deploy() error {
	return nil
}
