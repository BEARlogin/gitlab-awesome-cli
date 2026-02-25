// internal/infrastructure/config/config.go
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitLabURL       string        `yaml:"gitlab_url"`
	Token           string        `yaml:"token"`
	Projects        []string      `yaml:"projects"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".glcli.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.RefreshInterval == 0 {
		cfg.RefreshInterval = 5 * time.Second
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func RunSetupWizard() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	cfg := &Config{
		RefreshInterval: 5 * time.Second,
	}

	fmt.Print("GitLab URL (e.g. https://gitlab.example.com): ")
	url, _ := reader.ReadString('\n')
	cfg.GitLabURL = strings.TrimSpace(url)

	fmt.Print("Personal Access Token: ")
	token, _ := reader.ReadString('\n')
	cfg.Token = strings.TrimSpace(token)

	fmt.Print("Projects (comma-separated, e.g. group/project1,group/project2): ")
	projects, _ := reader.ReadString('\n')
	for _, p := range strings.Split(strings.TrimSpace(projects), ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			cfg.Projects = append(cfg.Projects, p)
		}
	}

	path := DefaultPath()
	if err := cfg.Save(path); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}
	fmt.Printf("Config saved to %s\n", path)
	return cfg, nil
}
