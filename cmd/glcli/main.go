package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	gitlabinfra "github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/gitlab"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui"
)

func main() {
	// Log to file if GLCLI_LOG is set, otherwise discard
	if logPath := os.Getenv("GLCLI_LOG"); logPath != "" {
		_ = os.MkdirAll(filepath.Dir(logPath), 0755)
		if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			log.SetOutput(f)
			log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
			defer f.Close()
		}
	} else {
		log.SetOutput(io.Discard)
	}

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Println("No config found. Let's set up glcli!")
		cfg, err = config.RunSetupWizard()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	client, err := gitlabinfra.NewClient(cfg.GitLabURL, cfg.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GitLab client error: %v\n", err)
		os.Exit(1)
	}

	projectRepo := gitlabinfra.NewProjectRepo(client)
	pipelineRepo := gitlabinfra.NewPipelineRepo(client)
	jobRepo := gitlabinfra.NewJobRepo(client)

	pipelineSvc := service.NewPipelineService(projectRepo, pipelineRepo)
	jobSvc := service.NewJobService(jobRepo)

	app := tui.NewApp(cfg, pipelineSvc, jobSvc)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
