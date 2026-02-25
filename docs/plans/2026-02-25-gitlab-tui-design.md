# GitLab Awesome CLI — TUI Design

## Overview

Interactive terminal UI for GitLab (like k9s for Kubernetes). Multi-view navigation between projects, pipelines, jobs, and logs. Built with Go + Bubble Tea.

## Tech Stack

- **Language**: Go
- **TUI**: charmbracelet/bubbletea + bubbles + lipgloss
- **GitLab API**: go-gitlab (xanzy/go-gitlab)
- **Config**: YAML (~/.glcli.yaml)

## Navigation

4 views, switchable by number keys or drill-down:

```
[1] Projects → [2] Pipelines → [3] Jobs → [4] Log
```

- `1-4` — direct view switch
- `Enter` — drill down
- `Esc` — back
- Breadcrumb at top shows current path

## Views

### Projects
List of configured projects with pipeline counts and active status.

### Pipelines
Pipelines for selected project: ID, branch, status, age, job count.

### Jobs
Jobs for selected pipeline: name, status, duration. Actions: run manual, retry failed, cancel running.

### Log
Streaming log output for selected job. Live-updates while job is running.

## Hotkeys

| Key     | Action              |
|---------|---------------------|
| `1-4`   | Switch view         |
| `Enter` | Drill down          |
| `Esc`   | Back                |
| `r`     | Run manual / Retry  |
| `c`     | Cancel running      |
| `/`     | Filter              |
| `q`     | Quit                |
| `?`     | Help                |

## Config

`~/.glcli.yaml`:
```yaml
gitlab_url: https://gitlab.example.com
token: glpat-xxxxx
projects:
  - group/awesome-project
  - group/backend-api
refresh_interval: 5s
```

First run without config triggers interactive setup wizard.

## Architecture

```
cmd/glcli/main.go
internal/
  config/config.go          — config loading, setup wizard
  gitlab/client.go          — GitLab API client wrapper
  tui/
    app.go                  — main Bubble Tea model, view routing
    views/
      projects.go           — projects list view
      pipelines.go          — pipelines list view
      jobs.go               — jobs list view
      log.go                — job log streaming view
    components/
      breadcrumb.go         — navigation breadcrumb
      statusbar.go          — hotkey hints at bottom
      confirm.go            — confirmation dialog
    styles/
      styles.go             — lipgloss styles
```

## Live Updates

- Ticker every N seconds (configurable) polls GitLab API
- Statuses update in-place
- Job log streams via GitLab API (chunked)
- Failed jobs update status immediately
