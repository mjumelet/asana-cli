# asana-cli

A fast, lightweight command-line interface for [Asana](https://asana.com) built in Go.

## Overview

asana-cli provides a simple way to interact with Asana from your terminal. List tasks, search across your workspace, add comments, and manage projects—all without leaving the command line.

## Features

- **Task Management** - List, search, and view detailed task information
- **Comments** - Add plain text or rich HTML comments to tasks
- **Projects** - Browse and filter projects in your workspace
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
# Clone the repository
git clone https://github.com/mjumelet/asana-cli.git
cd asana-cli

# Build
go build -o asana .

# Optionally, move to your PATH
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

## Usage

### Tasks

```bash
# List my tasks (shortcut)
asana tasks list -m

# List tasks assigned to you
asana tasks list -a me

# List tasks in a specific project
asana tasks list -p PROJECT_GID

# Filter by tag
asana tasks list -t TAG_GID

# Filter by due date
asana tasks list -m -d today      # Due today
asana tasks list -m -d tomorrow   # Due tomorrow
asana tasks list -m -d week       # Due this week
asana tasks list -m -d overdue    # Overdue tasks
asana tasks list -m -d 2024-03-15 # Due on specific date

# Combine filters
asana tasks list -p PROJECT_GID -d week -s modified_at

# Sort options: due_date (default), created_at, modified_at
asana tasks list -m -s created_at

# Include completed tasks
asana tasks list -m --all

# Search for tasks
asana tasks search "bug fix"

# Get task details
asana tasks get TASK_GID

# Get task details with comments and activity
asana tasks get TASK_GID --comments

# Add a comment
asana tasks comment TASK_GID "This is done!"

# Add an HTML comment
asana tasks comment TASK_GID "<strong>Done!</strong> See <a href='https://example.com'>results</a>" --html
```

### Task Filtering Options

| Flag | Description |
|------|-------------|
| `-m, --mine` | Show only tasks assigned to me |
| `-p, --project` | Filter by project GID |
| `-a, --assignee` | Filter by assignee GID (or `me`) |
| `-t, --tag` | Filter by tag GID |
| `-d, --due` | Filter by due date: `today`, `tomorrow`, `week`, `overdue`, or `YYYY-MM-DD` |
| `-s, --sort` | Sort by: `due_date`, `created_at`, `modified_at` |
| `-l, --limit` | Maximum number of results (default: 25) |
| `--all` | Include completed tasks |

### Task Management

```bash
# Create a new task
asana tasks create "Fix login bug" -a me -d 2024-03-20 -p PROJECT_GID

# Create task with description
asana tasks create "Review PR" -n "Please review the authentication changes" -a me

# Mark task as complete
asana tasks complete TASK_GID

# Reopen a completed task
asana tasks reopen TASK_GID

# Update a task
asana tasks update TASK_GID -n "Updated title" -a USER_GID -d 2024-03-25

# Delete a task
asana tasks delete TASK_GID
asana tasks delete TASK_GID -f  # Skip confirmation
```

### Projects

```bash
# List all projects
asana projects list

# Include archived projects
asana projects list -a

# Limit results
asana projects list -l 10
```

### Users

```bash
# List all users in the workspace
asana users list

# Get current user info
asana users me
```

### Summary & Reporting

```bash
# Get task summary for the workspace
asana summary

# Get summary for a specific project
asana summary -p PROJECT_GID

# Output as JSON for processing
asana summary -j
```

### Output Formats

All list commands support JSON output for scripting:

```bash
# JSON output
asana tasks list -a me -j

# Pipe to jq
asana tasks list -a me -j | jq '.[].name'
```

### Global Flags

```bash
# Use a custom config file
asana -c /path/to/.env tasks list

# Show help
asana --help
asana tasks --help
asana tasks list --help
```

## Commands Reference

| Command | Description |
|---------|-------------|
| `tasks list` | List tasks with optional filters |
| `tasks get <GID>` | Get detailed task information |
| `tasks create <name>` | Create a new task |
| `tasks complete <GID>` | Mark a task as complete |
| `tasks reopen <GID>` | Reopen a completed task |
| `tasks update <GID>` | Update task details |
| `tasks delete <GID>` | Delete a task |
| `tasks search <query>` | Search for tasks in your workspace |
| `tasks comment <GID> <message>` | Add a comment to a task |
| `projects list` | List projects in the workspace |
| `users list` | List workspace users |
| `users me` | Show current user |
| `summary` | Show task statistics |
| `configure` | Show configuration help |
| `-v, --version` | Show version information |

## HTML Comments

When using `--html` with the comment command, the following tags are supported:

- `<strong>`, `<em>`, `<u>`, `<s>` - Text formatting
- `<code>`, `<pre>` - Code formatting
- `<ol>`, `<ul>`, `<li>` - Lists
- `<a href="...">` - Links
- `<blockquote>` - Quotes

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
