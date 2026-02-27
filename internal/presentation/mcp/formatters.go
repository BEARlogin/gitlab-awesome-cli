package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
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

func formatMergeRequest(mr entity.MergeRequest) string {
	state := valueobject.MRState(mr.State)
	age := time.Since(mr.UpdatedAt).Truncate(time.Second)
	draft := ""
	if mr.Draft {
		draft = " [Draft]"
	}
	return fmt.Sprintf("- %s !%d%s | %s → %s | %s | @%s | %s ago | %s",
		state.Symbol(), mr.IID, draft, mr.SourceBranch, mr.TargetBranch, mr.Title, mr.Author, age, mr.WebURL)
}

func formatMergeRequests(mrs []entity.MergeRequest) string {
	if len(mrs) == 0 {
		return "No merge requests found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d merge request(s):\n\n", len(mrs))
	for _, mr := range mrs {
		b.WriteString(formatMergeRequest(mr))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatMergeRequestDetail(mr *entity.MergeRequest) string {
	state := valueobject.MRState(mr.State)
	draft := ""
	if mr.Draft {
		draft = " [Draft]"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## %s !%d: %s%s\n\n", state.Symbol(), mr.IID, mr.Title, draft)
	fmt.Fprintf(&b, "- **State:** %s\n", mr.State)
	fmt.Fprintf(&b, "- **Author:** @%s\n", mr.Author)
	fmt.Fprintf(&b, "- **Branch:** %s → %s\n", mr.SourceBranch, mr.TargetBranch)
	fmt.Fprintf(&b, "- **Merge status:** %s\n", mr.MergeStatus)
	fmt.Fprintf(&b, "- **URL:** %s\n", mr.WebURL)
	if mr.Description != "" {
		fmt.Fprintf(&b, "\n### Description\n\n%s\n", mr.Description)
	}
	return b.String()
}

func formatMRNotes(notes []entity.MRNote) string {
	if len(notes) == 0 {
		return "No notes found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d note(s):\n\n", len(notes))
	for _, n := range notes {
		age := time.Since(n.CreatedAt).Truncate(time.Second)
		prefix := ""
		if n.System {
			prefix = "[system] "
		}
		fmt.Fprintf(&b, "- %s@%s (%s ago): %s\n", prefix, n.Author, age, n.Body)
	}
	return b.String()
}

func formatMRDiffs(diffs []entity.MRDiff) string {
	if len(diffs) == 0 {
		return "No diffs found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d changed file(s):\n\n", len(diffs))
	for _, d := range diffs {
		label := d.NewPath
		if d.NewFile {
			label += " (new)"
		} else if d.DeletedFile {
			label += " (deleted)"
		} else if d.RenamedFile {
			label = fmt.Sprintf("%s → %s (renamed)", d.OldPath, d.NewPath)
		}
		fmt.Fprintf(&b, "### %s\n```diff\n%s\n```\n\n", label, d.Diff)
	}
	return b.String()
}

func formatCommit(c entity.Commit) string {
	age := time.Since(c.CreatedAt).Truncate(time.Second)
	return fmt.Sprintf("- `%s` %s — %s (%s ago)", c.ShortID, c.Title, c.AuthorName, age)
}

func formatCommits(commits []entity.Commit) string {
	if len(commits) == 0 {
		return "No commits found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d commit(s):\n\n", len(commits))
	for _, c := range commits {
		b.WriteString(formatCommit(c))
		b.WriteByte('\n')
	}
	return b.String()
}
