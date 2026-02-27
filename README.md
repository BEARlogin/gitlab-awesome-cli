# glcli

```
  ██████╗ ██╗      ██████╗██╗     ██╗
 ██╔════╝ ██║     ██╔════╝██║     ██║
 ██║  ███╗██║     ██║     ██║     ██║
 ██║   ██║██║     ██║     ██║     ██║
 ╚██████╔╝███████╗╚██████╗███████╗██║
  ╚═════╝ ╚══════╝ ╚═════╝╚══════╝╚═╝
```

**An interactive terminal UI for GitLab — like k9s, but for pipelines and merge requests.**

You're deep in the terminal — 16 tabs open, Claude Code running, deploys in flight. The last thing you want is to alt-tab into a browser, navigate to GitLab, click through three projects, and wait for the UI to load just to check if a pipeline passed.

glcli keeps you in the terminal. Monitor pipelines across all your projects, stream job logs, retry failures, review merge request diffs, create and merge MRs — without ever opening a browser tab. The built-in MCP server lets your AI assistant do the same.

---

## Preview

```
 1:Projects   2:Pipelines   3:Jobs   4:Log   5:MRs
 infrasamurai/app

  PROJECT          #ID     BRANCH           STATUS     AGE
  ──────────────── ─────── ──────────────── ────────── ────────
▸ app              #7401   main             ✓ success  2h ago
  app              #7399   feat/payments    ● running  3h ago
  dashboard        #7388   main             ✓ success  5h ago
  backend          #7380   fix/memory-leak  ✗ failed   8h ago
  instabot         #7372   main             ⏸ manual   12h ago

  1/47
 ↑↓ navigate  fn↑↓ page  Enter jobs  c commits  / filter  l limit  Tab next tab  q quit
```

### Merge Request Detail with Colored Diffs

```
 1:Projects   2:Pipelines   3:Jobs   4:Log   5:MRs
 mygroup/api  !42

 MR !42: Add user authentication

 ◉ !42: Add user authentication
 Author: @bearlogin  |  feat/auth → main  |  opened  |  can_be_merged

 [Diffs]  |   Comments
 ────────────────────────────────────────────────────────────

 src/auth/handler.go (new)
 @@ -0,0 +1,25 @@
 +package auth
 +
 +func NewHandler(svc *Service) *Handler {
 +    return &Handler{svc: svc}
 +}

 src/main.go
 @@ -10,6 +10,8 @@
  func main() {
      router := mux.NewRouter()
 +    authSvc := auth.NewService(db)
 +    router.Handle("/login", auth.NewHandler(authSvc))
      router.Handle("/api", apiHandler)
  }

 ↑↓ scroll  Tab diff/comments  r refresh  a approve  m merge  Esc back  q quit
```

### Create Merge Request with Branch Autocomplete

```
 1:Projects   2:Pipelines   3:Jobs   4:Log   5:MRs
 mygroup/api  New MR

  Create Merge Request

▸ Source Branch   feat█
   ▸ feat/auth
     feat/payments
     feat/notifications
  Target Branch   main
  Title
  Description
  Draft           [ ]

  Tab/↑↓ navigate  Enter select/next  Ctrl+S submit  Esc cancel
```

---

## Features

### Pipelines & Jobs
- **All pipelines at a glance** — aggregates pipelines from all configured projects on one screen
- **Live auto-refresh** — configurable polling interval
- **Job log streaming** — tail logs with viewport scrolling
- **Pipeline actions** — run manual jobs, retry failed, cancel running — with confirmation dialogs
- **Fuzzy filter** — press `/` to filter pipelines by project name, branch, or status
- **Pipeline limit control** — press `l` to cycle the fetch limit: 20 → 50 → 100 → 200
- **Commit history** — press `c` on a pipeline to view commits for that ref

### Merge Requests
- **MR list** — view open merge requests across all configured projects
- **MR detail** — diffs with syntax-aware coloring (green additions, red deletions, cyan hunk headers) and comments
- **Create MR** — interactive form with branch autocomplete from GitLab API, source/target validation
- **Approve & Merge** — one-key actions with confirmation dialogs
- **Force refresh** — press `r` to reload MR data

### General
- **Multi-view TUI** — Projects, Pipelines, Jobs, Log, and MRs tabs
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
| `read_api` | yes | List projects, pipelines, jobs, MRs, read logs |
| `api` | for actions | Play/retry/cancel jobs, create/approve/merge MRs |

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
| `5`              | Go to MRs view                     |
| `Tab`            | Next view                          |
| `Shift+Tab`      | Previous view                      |
| `Esc`            | Go back                            |
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

### Projects view

| Key | Action                          |
|-----|---------------------------------|
| `a` | Add project (with search)       |
| `d` | Remove project                  |
| `m` | Go to MRs view                  |

### Pipelines view

| Key | Action                                           |
|-----|--------------------------------------------------|
| `/` | Open filter prompt                               |
| `l` | Cycle pipeline limit (20 → 50 → 100 → 200)      |
| `c` | View commits for selected pipeline's ref         |

### Jobs view

| Key | Action                          |
|-----|---------------------------------|
| `r` | Run manual / retry failed job   |
| `c` | Cancel running job              |

### MRs view

| Key | Action                          |
|-----|---------------------------------|
| `/` | Open filter prompt              |
| `n` | Create new merge request        |

### MR Detail view

| Key   | Action                          |
|-------|---------------------------------|
| `Tab` | Switch between Diffs/Comments   |
| `r`   | Force refresh                   |
| `a`   | Approve merge request           |
| `m`   | Merge merge request             |

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

If `glcli-mcp` is in your `$PATH`:

```bash
claude mcp add glcli -- glcli-mcp
```

If it's **not** in your `$PATH` (e.g. installed via `go install`), use the full path:

```bash
claude mcp add glcli -- ~/go/bin/glcli-mcp
```

To enable debug logging, add the server manually to `~/.claude/settings.json` (global) or `.claude/settings.json` (project-level):

```json
{
  "mcpServers": {
    "glcli": {
      "command": "glcli-mcp",
      "env": {
        "GLCLI_MCP_LOG": "/tmp/glcli-mcp.log"
      }
    }
  }
}
```

> **Note:** avoid wrapping the binary with `env` in the `command` field — use the `env` object instead, otherwise stdin may not be forwarded correctly.

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
| `list_merge_requests` | List merge requests for a project |
| `get_merge_request` | Get details of a specific merge request |
| `list_mr_notes` | List comments/notes on a merge request |
| `get_mr_diffs` | Get diffs of a merge request |
| `approve_mr` | Approve a merge request |
| `merge_mr` | Merge a merge request |
| `create_merge_request` | Create a new merge request |
| `list_pipeline_commits` | List commits for a pipeline ref |

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
> Show open merge requests
> Create a merge request from feat/auth to main with title "Add authentication"
> Approve MR !42
> What are the diffs in MR !42?
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
      views/            — Projects, Pipelines, Jobs, Log, MRs, MR Detail, MR Create, Commits
      components/       — shared widgets (statusbar, breadcrumb, confirm dialog)
      styles/           — lipgloss theme (incl. diff coloring)
      keymap/           — key normalization incl. Russian layout
    mcp/                — MCP server (tools, resources, formatters)
```

---

## License

MIT — Copyright (c) 2026 bearlogin
