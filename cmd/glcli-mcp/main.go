package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	gitlabinfra "github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/gitlab"
	mcpserver "github.com/bearlogin/gitlab-awesome-cli/internal/presentation/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func setupLog() *os.File {
	logPath := os.Getenv("GLCLI_MCP_LOG")
	if logPath == "" {
		log.SetOutput(os.Stderr)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log dir: %v\n", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	return f
}

func main() {
	logFile := setupLog()
	if logFile != nil {
		defer logFile.Close()
	}
	log.Printf("glcli-mcp %s starting", version)

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("config loaded: url=%s projects=%v", cfg.GitLabURL, cfg.Projects)

	client, err := gitlabinfra.NewClient(cfg.GitLabURL, cfg.Token)
	if err != nil {
		log.Fatalf("GitLab client error: %v", err)
	}
	log.Print("gitlab client created")

	projectRepo := gitlabinfra.NewProjectRepo(client)
	pipelineRepo := gitlabinfra.NewPipelineRepo(client)
	jobRepo := gitlabinfra.NewJobRepo(client)

	pipelineSvc := service.NewPipelineService(projectRepo, pipelineRepo)
	jobSvc := service.NewJobService(jobRepo)

	server := mcpserver.NewServer(cfg, pipelineSvc, jobSvc, version)
	log.Print("mcp server created, starting stdio transport")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
	log.Print("mcp server stopped")
}
