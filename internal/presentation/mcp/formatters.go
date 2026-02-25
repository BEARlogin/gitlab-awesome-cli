package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

func formatProject(p entity.Project) string {
	return fmt.Sprintf("- **%s** (ID: %d) — %s | pipelines: %d, active: %d",
		p.PathWithNS, p.ID, p.WebURL, p.PipelineCount, p.ActiveCount)
}

func formatProjects(projects []entity.Project) string {
	if len(projects) == 0 {
		return "No projects found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d project(s):\n\n", len(projects))
	for _, p := range projects {
		b.WriteString(formatProject(p))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatPipeline(p entity.Pipeline) string {
	age := time.Since(p.CreatedAt).Truncate(time.Second)
	return fmt.Sprintf("- %s #%d | %s | ref: %s | %s ago | %d jobs",
		p.Status.Symbol(), p.ID, p.ProjectPath, p.Ref, age, p.JobCount)
}

func formatPipelines(pipelines []entity.Pipeline) string {
	if len(pipelines) == 0 {
		return "No pipelines found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d pipeline(s):\n\n", len(pipelines))
	for _, p := range pipelines {
		b.WriteString(formatPipeline(p))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatJob(j entity.Job) string {
	dur := fmt.Sprintf("%.0fs", j.Duration)
	return fmt.Sprintf("- %s %s (ID: %d) | stage: %s | %s | %s",
		j.Status.Symbol(), j.Name, j.ID, j.Stage, dur, j.WebURL)
}

func formatJobs(jobs []entity.Job) string {
	if len(jobs) == 0 {
		return "No jobs found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d job(s):\n\n", len(jobs))
	for _, j := range jobs {
		b.WriteString(formatJob(j))
		b.WriteByte('\n')
	}
	return b.String()
}
