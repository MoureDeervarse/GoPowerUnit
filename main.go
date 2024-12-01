package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/mouredeervarse/go-power-unit/internal/config"
	"github.com/mouredeervarse/go-power-unit/internal/unit"
	"github.com/mouredeervarse/go-power-unit/internal/watcher"
)

func main() {
	fmt.Println("GoPowerUnit on running...")

	// create channel for signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var rootCmd = &cobra.Command{
		Use:   "live-reload",
		Short: "A live reload tool for Go projects",
	}

	var localCmd = &cobra.Command{
		Use:   "local",
		Short: "Run project locally with live reload",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			w := watcher.New(cfg)
			r := unit.NewLocalUnit(cfg.ProjectPath)

			// setup signal handler
			go func() {
				<-sigChan
				fmt.Println("\nShutting down...")
				if err := r.Cleanup(); err != nil {
					log.Printf("Cleanup error: %v", err)
				}
				os.Exit(0)
			}()

			// first start
			_ = r.Reload()
			w.Watch(r.Reload)
		},
	}

	var dockerCmd = &cobra.Command{
		Use:   "docker",
		Short: "Run project in Docker with live reload",
		Run: func(cmd *cobra.Command, args []string) {
			// cfg := config.Load()
			// w := watcher.New(cfg)
			// r := unit.NewDockerUnit(cfg.ProjectPath, cfg.DockerImage, cfg.DockerImage)

			// w.Watch(r.Reload)
		},
	}

	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy to remote environment",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			r := unit.NewDeployUnit(cfg.ProjectPath, cfg.DockerImage, cfg.DockerRegistry, cfg.DeployURL)
			r.Deploy()
		},
	}

	rootCmd.AddCommand(localCmd, dockerCmd, deployCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
