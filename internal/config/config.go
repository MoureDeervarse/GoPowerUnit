package config

type Config struct {
	ProjectPath    string
	WatchPaths     []string
	IgnorePaths    []string
	DockerImage    string
	DockerRegistry string
	DeployURL      string
}

func Load() *Config {
	// TODO: Implement configuration loading from file or environment variables
	return &Config{
		ProjectPath: ".",
		WatchPaths:  []string{"."},
		IgnorePaths: []string{
			"vendor/",
			".git/",
			"node_modules/",
		},
	}
}
