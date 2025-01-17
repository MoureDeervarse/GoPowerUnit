package config

type Config struct {
	ProjectPath    string
	WatchPaths     []string
	IgnorePaths    []string
	DockerImage    string
	DockerRegistry string
	DeployURL      string
	DockerfilePath string
	Ports          []string
}

func Load() *Config {
	return &Config{
		ProjectPath: ".",
		WatchPaths:  []string{"."},
		IgnorePaths: []string{
			"vendor/",
			".git/",
			"node_modules/",
		},
		DockerfilePath: ".",
	}
}
