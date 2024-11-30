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
	// 이미지 빌드
	buildCmd := exec.Command("docker", "build", "-t", u.imageName, u.projectPath)
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// 레지스트리 태그 추가
	registryImage := fmt.Sprintf("%s/%s", u.registryURL, u.imageName)
	tagCmd := exec.Command("docker", "tag", u.imageName, registryImage)
	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("docker tag failed: %w", err)
	}

	// 레지스트리에 푸시
	pushCmd := exec.Command("docker", "push", registryImage)
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	// 배포 엔드포인트 호출
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
	// 배포된 서비스의 중지는 배포 시스템에서 관리
	return nil
}

func (u *DeployUnit) Reload() error {
	return u.Start()
}

func (u *DeployUnit) Deploy() error {
	return nil
}
