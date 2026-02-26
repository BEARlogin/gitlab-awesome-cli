# glcli

```
  ██████╗ ██╗      ██████╗██╗     ██╗
 ██╔════╝ ██║     ██╔════╝██║     ██║
 ██║  ███╗██║     ██║     ██║     ██║
 ██║   ██║██║     ██║     ██║     ██║
 ╚██████╔╝███████╗╚██████╗███████╗██║
  ╚═════╝ ╚══════╝ ╚═════╝╚══════╝╚═╝
```

**An interactive terminal UI for GitLab — like k9s, but for pipelines.**

Monitor pipelines across all your projects, stream job logs, retry failures, and manage CI/CD without ever opening a browser tab.

---

## Preview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  glcli  [1] Projects  [2] Pipelines  [3] Jobs  [4] Log          ↻ 5s  l:50 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  PIPELINES  (47 total)  /filter                                             │
│                                                                             │
│  PROJECT                  BRANCH          STATUS     DURATION   STARTED     │
│  ──────────────────────── ─────────────── ────────── ────────── ─────────── │
│  group/backend            main            ● running  2m 14s     just now    │
│  group/frontend           feat/auth       ✓ success  4m 33s     3 min ago   │
│  group/api-gateway        main            ✓ success  1m 58s     7 min ago   │
│  group/worker             fix/memory-leak ✗ failed   3m 01s     12 min ago  │
│  group/infra              main            ● running  0m 44s     just now    │
│  group/backend            feat/payments   ⏸ manual   —          15 min ago  │
│                                                                             │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  j/k navigate  enter select  r retry  c cancel  / filter  q quit           │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Features

- **Multi-view TUI** — Projects, Pipelines, Jobs, and Log tabs in a single terminal window
- **All pipelines at a glance** — aggregates pipelines from all configured projects on one screen
- **Live auto-refresh** — configurable polling interval with a visible countdown
- **Job log streaming** — tail logs in real time with viewport scrolling
- **Pipeline actions** — run manual jobs, retry failed, cancel running — all with a confirmation dialog
- **Fuzzy filter** — press `/` to filter pipelines by project name, branch, or status
- **Pipeline limit control** — press `l` to cycle the fetch limit: 20 → 50 → 100 → 200
- **Add/remove projects** — interactive autocomplete search against the GitLab API
- **Vim-style navigation** — `j`/`k`, `g`/`G`, `Ctrl+u`/`Ctrl+d`
- **Russian keyboard layout support** — keys work regardless of active layout
- **Clean config** — single YAML file at `~/.glcli.yaml`
- **MCP Server** — let AI assistants (Claude Code, etc.) interact with your GitLab via Model Context Protocol

---

## Install

### go install

```bash
go install github.com/bearlogin/gitlab-awesome-cli/cmd/glcli@latest
```

### Homebrew (coming soon)

```bash
brew install bearlogin/tap/glcli
```

### Pre-built binaries

Download the latest binary for your platform from the [Releases](https://github.com/bearlogin/gitlab-awesome-cli/releases) page.

```bash
# Linux amd64
curl -Lo glcli https://github.com/bearlogin/gitlab-awesome-cli/releases/latest/download/glcli_linux_amd64
chmod +x glcli
sudo mv glcli /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -Lo glcli https://github.com/bearlogin/gitlab-awesome-cli/releases/latest/download/glcli_darwin_arm64
chmod +x glcli
sudo mv glcli /usr/local/bin/
```

---

## Quick Start

Run `glcli` for the first time and the setup wizard will guide you through initial configuration:

```
$ glcli

GitLab URL (e.g. https://gitlab.example.com): https://gitlab.mycompany.com
Personal Access Token: glpat-xxxxxxxxxxxxxxxxxxxx
Projects (comma-separated, e.g. group/project1,group/project2): mygroup/api,mygroup/frontend

Config saved to ~/.glcli.yaml
```

You need a GitLab **Personal Access Token** with the following scopes:

| Scope | Required | What it's used for |
|-------|----------|--------------------|
| `read_api` | yes | List projects, pipelines, jobs, read logs |
| `api` | for actions | Play/retry/cancel jobs, search projects |

---

## Configuration

Config lives at `~/.glcli.yaml`:

```yaml
gitlab_url: https://gitlab.example.com
token: glpat-xxxxxxxxxxxxxxxxxxxx
projects:
  - group/backend
  - group/frontend
  - group/infra
refresh_interval: 5s
pipeline_limit: 50
```

| Field              | Type     | Default | Description                                      |
|--------------------|----------|---------|--------------------------------------------------|
| `gitlab_url`       | string   | —       | Base URL of your GitLab instance                 |
| `token`            | string   | —       | Personal Access Token                            |
| `projects`         | []string | —       | List of `namespace/project` slugs to monitor     |
| `refresh_interval` | duration | `5s`    | How often to poll GitLab for updates             |
| `pipeline_limit`   | int      | `50`    | Maximum pipelines fetched per project            |

---

## Keybindings

### Global

| Key              | Action                             |
|------------------|------------------------------------|
| `1`              | Go to Projects view                |
| `2`              | Go to Pipelines view               |
| `3`              | Go to Jobs view                    |
| `4`              | Go to Log view                     |
| `Tab`            | Next view                          |
| `Shift+Tab`      | Previous view                      |
| `q` / `Ctrl+C`   | Quit                               |

### Navigation

| Key          | Action                          |
|--------------|---------------------------------|
| `j` / `↓`    | Move cursor down                |
| `k` / `↑`    | Move cursor up                  |
| `g`          | Jump to top                     |
| `G`          | Jump to bottom                  |
| `Ctrl+d`     | Scroll half-page down           |
| `Ctrl+u`     | Scroll half-page up             |
| `Enter`      | Select / drill into             |
| `Esc`        | Go back / close dialog          |

### Pipelines view

| Key | Action                                           |
|-----|--------------------------------------------------|
| `/` | Open filter prompt                               |
| `l` | Cycle pipeline limit (20 → 50 → 100 → 200)      |
| `r` | Retry selected pipeline                          |
| `c` | Cancel selected pipeline                         |

### Jobs view

| Key | Action                          |
|-----|---------------------------------|
| `r` | Retry selected job              |
| `p` | Play / trigger manual job       |
| `c` | Cancel selected job             |

### Projects view

| Key | Action                          |
|-----|---------------------------------|
| `a` | Add project (with search)       |
| `d` | Remove project                  |

### Log view

| Key          | Action                          |
|--------------|---------------------------------|
| `j` / `↓`    | Scroll down                     |
| `k` / `↑`    | Scroll up                       |
| `Ctrl+d`     | Scroll half-page down           |
| `Ctrl+u`     | Scroll half-page up             |
| `g`          | Jump to top                     |
| `G`          | Jump to bottom (follow mode)    |

---

## MCP Server (AI Integration)

glcli ships with a built-in [MCP](https://modelcontextprotocol.io/) server — a separate binary that lets AI assistants work with your GitLab directly from the terminal.

### Install

```bash
make build-mcp
# or
go install github.com/bearlogin/gitlab-awesome-cli/cmd/glcli-mcp@latest
```

### Register in Claude Code

```bash
claude mcp add glcli glcli-mcp
```

Or with debug logging:

```bash
claude mcp add glcli -- env GLCLI_MCP_LOG=/tmp/glcli-mcp.log glcli-mcp
```

Or manually add to `.claude/settings.json` (project-level) or `~/.claude/settings.json` (global):

```json
{
  "mcpServers": {
    "glcli": {
      "command": "glcli-mcp"
    }
  }
}
```

### Available Tools

| Tool | Description |
|------|-------------|
| `list_projects` | List configured projects with pipeline counts |
| `list_pipelines` | List pipelines with optional filters (project, status, ref, limit) |
| `list_jobs` | List jobs for a specific pipeline |
| `get_job_log` | Get the log output of a job |
| `play_job` | Start a manual job |
| `retry_job` | Retry a failed job |
| `cancel_job` | Cancel a running/pending job |
| `search_projects` | Search GitLab projects by name or path |

### Resources

| URI | Description |
|-----|-------------|
| `gitlab://config` | Current configuration (token excluded) |

### Usage Examples

Once registered, just ask your AI assistant in natural language:

```
> Show me failed pipelines
> What's the log of the latest failed job in mygroup/api?
> Retry that job
> List all running pipelines on main branch
```

---

## Building from Source

Requirements: Go 1.21+

```bash
git clone https://github.com/bearlogin/gitlab-awesome-cli.git
cd gitlab-awesome-cli
make build        # TUI binary → dist/glcli
make build-mcp    # MCP server binary → dist/glcli-mcp
make build-all    # both
```

Other Makefile targets:

```bash
make run          # go run (TUI)
make test         # run tests
make lint         # run golangci-lint
make vet          # go vet
make install      # install glcli to $GOPATH/bin or /usr/local/bin
make install-mcp  # install glcli-mcp
make clean        # remove dist/
```

---

## Architecture

```
cmd/
  glcli/                — TUI entry point
  glcli-mcp/            — MCP server entry point
internal/
  domain/               — entities, value objects, repository interfaces
  application/service/  — use-case orchestration
  infrastructure/
    gitlab/             — GitLab API client (go-gitlab)
    config/             — YAML config loading + setup wizard
  presentation/
    tui/                — terminal UI
      views/            — Projects, Pipelines, Jobs, Log screens
      components/       — shared widgets (statusbar, breadcrumb, confirm dialog)
      styles/           — lipgloss theme
      keymap/           — key normalization incl. Russian layout
    mcp/                — MCP server (tools, resources, formatters)
```

---

## License

MIT — Copyright (c) 2026 bearlogin
