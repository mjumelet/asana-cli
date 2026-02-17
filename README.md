# asana-cli

A fast, lightweight command-line interface for [Asana](https://asana.com) built in Go.

## Overview

asana-cli provides a simple way to interact with Asana from your terminal. List tasks, search across your workspace, add comments, and manage projects—all without leaving the command line.

## Features

- **Task Management** - Create, update, complete, and delete tasks
- **Comments** - Add plain text or rich HTML comments to tasks
- **Projects** - Browse and filter projects in your workspace
- **Users** - List workspace members and get user info
- **Reporting** - Task summaries with statistics by assignee
- **Multiple Output Formats** - Human-readable tables or JSON for scripting
- **Flexible Configuration** - Environment variables, config files, or custom paths

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap mjumelet/tap
brew install asana-cli
```

### From Source

```bash
git clone https://github.com/mjumelet/asana-cli.git
cd asana-cli
go build -o asana .
mv asana /usr/local/bin/
```

### Requirements

- Go 1.21 or later (for building from source)

## Configuration

asana-cli requires two configuration values:

| Variable | Description |
|----------|-------------|
| `ASANA_TOKEN` | Your Asana Personal Access Token |
| `ASANA_WORKSPACE` | The GID of your Asana workspace |

### Getting Your Credentials

1. **Personal Access Token**: Generate one at [https://app.asana.com/0/my-apps](https://app.asana.com/0/my-apps)
2. **Workspace GID**: Find it in your Asana URL (`https://app.asana.com/0/<workspace_gid>/...`) or run `asana configure` for help

### Configuration Methods

Configuration is resolved in the following priority order:

1. **Environment variables** (highest priority)
2. **Config file** specified via `--config` flag
3. **`.env` file** in the current directory
4. **`~/.config/asana-cli/.env`** (XDG-style config directory)

### Example .env File

```bash
ASANA_TOKEN=1/1234567890:abcdefghijklmnop
ASANA_WORKSPACE=1234567890123456
```

Run `asana configure` to see all configuration options and setup instructions.

## Global Flags

These flags work with all commands:

| Flag | Description | Example |
|------|-------------|---------|
| `-c, --config` | Path to config file (.env format) | `asana -c ~/.my-asana.env tasks list` |
| `-v, --version` | Show version information | `asana -v` |
| `-h, --help` | Show help for any command | `asana tasks list --help` |

## Commands

### tasks list

List tasks with optional filters.

```bash
asana tasks list [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-m, --mine` | Show only tasks assigned to me | `asana tasks list -m` |
| `-p, --project` | Filter by project GID | `asana tasks list -p 1234567890` |
| `-a, --assignee` | Filter by assignee GID or `me` | `asana tasks list -a me` |
| `-t, --tag` | Filter by tag GID | `asana tasks list -t 9876543210` |
| `-d, --due` | Filter by due date | `asana tasks list -d today` |
| `-s, --sort` | Sort by: `due_date`, `created_at`, `modified_at` | `asana tasks list -s created_at` |
| `-l, --limit` | Maximum results (default: 100) | `asana tasks list -l 50` |
| `--all` | Include completed tasks | `asana tasks list -m --all` |
| `-j, --json` | Output as JSON | `asana tasks list -m -j` |

**Due date options:** `today`, `tomorrow`, `week`, `overdue`, or `YYYY-MM-DD`

**Examples:**

```bash
# List my tasks
asana tasks list -m

# List overdue tasks assigned to me
asana tasks list -m -d overdue

# List tasks in a project, sorted by modification date
asana tasks list -p 1234567890 -s modified_at

# List tasks due this week with a specific tag
asana tasks list -d week -t 9876543210

# List all tasks (including completed) as JSON
asana tasks list -m --all -j
```

### tasks get

Get detailed information about a task.

```bash
asana tasks get <task-gid> [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `--comments` | Include comments and activity | `asana tasks get 123 --comments` |
| `-j, --json` | Output as JSON | `asana tasks get 123 -j` |

**Examples:**

```bash
# Get task details
asana tasks get 1234567890123456

# Get task with all comments
asana tasks get 1234567890123456 --comments

# Get task as JSON (for scripting)
asana tasks get 1234567890123456 -j
```

### tasks create

Create a new task.

```bash
asana tasks create <name> [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-n, --notes` | Task description | `asana tasks create "Task" -n "Details here"` |
| `-a, --assignee` | Assignee GID or `me` | `asana tasks create "Task" -a me` |
| `-d, --due` | Due date (YYYY-MM-DD) | `asana tasks create "Task" -d 2024-03-20` |
| `-p, --project` | Project GID to add task to | `asana tasks create "Task" -p 123456` |
| `-j, --json` | Output as JSON | `asana tasks create "Task" -j` |

**Examples:**

```bash
# Create a simple task
asana tasks create "Fix login bug"

# Create task assigned to me with due date
asana tasks create "Review PR" -a me -d 2024-03-20

# Create task in a project with description
asana tasks create "Update documentation" -p 1234567890 -n "Update the API docs with new endpoints"

# Create task and get JSON response
asana tasks create "New feature" -a me -p 1234567890 -j
```

### tasks complete

Mark a task as complete.

```bash
asana tasks complete <task-gid>
```

**Example:**

```bash
asana tasks complete 1234567890123456
```

### tasks reopen

Reopen a completed task.

```bash
asana tasks reopen <task-gid>
```

**Example:**

```bash
asana tasks reopen 1234567890123456
```

### tasks update

Update an existing task.

```bash
asana tasks update <task-gid> [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-n, --name` | New task name | `asana tasks update 123 -n "New name"` |
| `--notes` | New task description | `asana tasks update 123 --notes "Updated desc"` |
| `-a, --assignee` | New assignee GID or `me` | `asana tasks update 123 -a me` |
| `-d, --due` | New due date (YYYY-MM-DD) | `asana tasks update 123 -d 2024-04-01` |
| `-j, --json` | Output as JSON | `asana tasks update 123 -n "New" -j` |

**Examples:**

```bash
# Change task name
asana tasks update 1234567890 -n "Updated task title"

# Reassign task and change due date
asana tasks update 1234567890 -a 9876543210 -d 2024-04-15

# Update description
asana tasks update 1234567890 --notes "New detailed description"
```

### tasks delete

Delete a task.

```bash
asana tasks delete <task-gid> [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-f, --force` | Skip confirmation prompt | `asana tasks delete 123 -f` |

**Examples:**

```bash
# Delete with confirmation
asana tasks delete 1234567890123456

# Delete without confirmation
asana tasks delete 1234567890123456 -f
```

### tasks comment

Add a comment to a task.

```bash
asana tasks comment <task-gid> <message> [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `--html` | Treat message as HTML rich text | `asana tasks comment 123 "<b>Done</b>" --html` |

**Examples:**

```bash
# Add plain text comment
asana tasks comment 1234567890 "This is done!"

# Add HTML formatted comment
asana tasks comment 1234567890 "<strong>Completed!</strong> See <a href='https://example.com'>results</a>" --html
```

**Supported HTML tags:** `<strong>`, `<em>`, `<u>`, `<s>`, `<code>`, `<pre>`, `<ol>`, `<ul>`, `<li>`, `<a>`, `<blockquote>`

### tasks search

Search for tasks in your workspace.

```bash
asana tasks search <query> [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-l, --limit` | Maximum results (default: 100) | `asana tasks search "bug" -l 50` |
| `-j, --json` | Output as JSON | `asana tasks search "bug" -j` |

**Examples:**

```bash
# Search for tasks
asana tasks search "bug fix"

# Search with limited results
asana tasks search "documentation" -l 20

# Search and output as JSON
asana tasks search "urgent" -j
```

### projects list

List projects in the workspace.

```bash
asana projects list [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-a, --archived` | Include archived projects | `asana projects list -a` |
| `-l, --limit` | Maximum results (default: 50) | `asana projects list -l 100` |
| `-j, --json` | Output as JSON | `asana projects list -j` |

**Examples:**

```bash
# List active projects
asana projects list

# List all projects including archived
asana projects list -a

# List projects as JSON
asana projects list -j

# List more projects
asana projects list -l 100
```

### users list

List all users in the workspace.

```bash
asana users list [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-j, --json` | Output as JSON | `asana users list -j` |

**Examples:**

```bash
# List all workspace users
asana users list

# List users as JSON
asana users list -j
```

### users me

Show the current authenticated user.

```bash
asana users me [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-j, --json` | Output as JSON | `asana users me -j` |

**Examples:**

```bash
# Show current user
asana users me

# Get current user as JSON
asana users me -j
```

### summary

Show task summary and statistics.

```bash
asana summary [flags]
```

**Flags:**

| Flag | Description | Example |
|------|-------------|---------|
| `-p, --project` | Filter by project GID | `asana summary -p 1234567890` |
| `-j, --json` | Output as JSON | `asana summary -j` |

**Examples:**

```bash
# Get workspace summary
asana summary

# Get summary for a specific project
asana summary -p 1234567890123456

# Get summary as JSON
asana summary -j
```

**Output includes:**
- Total, open, and completed task counts
- Overdue task count
- Unassigned task count
- Tasks per assignee (sorted by count)

### configure

Show configuration help and setup instructions.

```bash
asana configure
```

## JSON Output

All list and get commands support `-j` or `--json` for JSON output, useful for scripting:

```bash
# Get task names with jq
asana tasks list -m -j | jq '.[].name'

# Get project GIDs
asana projects list -j | jq '.[].gid'

# Count tasks per assignee
asana summary -j | jq '.ByAssignee'
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Kong](https://github.com/alecthomas/kong) for CLI parsing
- Uses [godotenv](https://github.com/joho/godotenv) for configuration file loading
