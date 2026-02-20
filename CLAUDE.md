# Asana CLI - Development Notes

## Asana API Reference

Base URL: `https://app.asana.com/api/1.0`

API Documentation: https://developers.asana.com/docs
OpenAPI Spec: https://github.com/Asana/openapi

### Authentication

Uses Personal Access Token via `Authorization: Bearer <token>` header.

### Key Endpoints

#### Tasks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/tasks/{task_gid}` | Get task details |
| POST | `/tasks` | Create a new task |
| PUT | `/tasks/{task_gid}` | Update a task |
| DELETE | `/tasks/{task_gid}` | Delete a task |
| GET | `/workspaces/{workspace_gid}/tasks/search` | Search tasks |
| GET | `/tasks/{task_gid}/stories` | Get task comments/activity |
| POST | `/tasks/{task_gid}/stories` | Add comment to task |

#### Task Fields (for create/update)
- `name` - Task title
- `notes` - Task description (plain text)
- `html_notes` - Task description (HTML)
- `completed` - Boolean to mark complete
- `due_on` - Due date (YYYY-MM-DD)
- `due_at` - Due datetime (ISO 8601)
- `assignee` - User GID or "me"
- `projects` - Array of project GIDs
- `tags` - Array of tag GIDs
- `parent` - Parent task GID (for subtasks)

#### Projects
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/{project_gid}` | Get project details |
| GET | `/workspaces/{workspace_gid}/projects` | List projects |
| POST | `/projects` | Create a project |
| PUT | `/projects/{project_gid}` | Update a project |
| DELETE | `/projects/{project_gid}` | Delete a project |
| GET | `/projects/{project_gid}/tasks` | Get tasks in project |
| GET | `/projects/{project_gid}/sections` | Get project sections |

#### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/me` | Get current user |
| GET | `/users/{user_gid}` | Get user details |
| GET | `/workspaces/{workspace_gid}/users` | List workspace users |

#### Workspaces
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces` | List workspaces |
| GET | `/workspaces/{workspace_gid}` | Get workspace details |

#### Tags
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/{workspace_gid}/tags` | List tags |
| GET | `/tags/{tag_gid}` | Get tag details |
| POST | `/workspaces/{workspace_gid}/tags` | Create a tag |

#### Attachments
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/tasks/{task_gid}/attachments` | List attachments on a task |
| GET | `/attachments/{attachment_gid}` | Get attachment details |
| POST | `/tasks/{task_gid}/attachments` | Upload a file (multipart/form-data) |
| DELETE | `/attachments/{attachment_gid}` | Delete an attachment |

#### Sections
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/{project_gid}/sections` | List project sections |
| POST | `/projects/{project_gid}/sections` | Create a section |
| GET | `/sections/{section_gid}/tasks` | Get tasks in section |

### Search Parameters

The `/workspaces/{workspace_gid}/tasks/search` endpoint supports:
- `text` - Full-text search
- `assignee.any` - Filter by assignee(s)
- `projects.any` - Filter by project(s)
- `tags.any` - Filter by tag(s)
- `completed` - true/false
- `is_subtask` - true/false
- `due_on`, `due_on.before`, `due_on.after` - Due date filters
- `sort_by` - due_date, created_at, modified_at, likes
- `sort_ascending` - true/false

### Rate Limiting

- Standard limit: 1500 requests per minute
- Free tier: 150 requests per minute
- Responses include `X-Asana-*` headers for rate limit info

### Pagination

- Use `limit` parameter (max 100)
- Response includes `next_page.offset` for pagination
- Pass `offset` parameter for next page

## LLM Agent Use Cases

This CLI is designed to be used by LLM agents for task automation:

1. **Task Management**
   - `asana tasks create` - Create tasks with natural language
   - `asana tasks complete` - Mark tasks done
   - `asana tasks assign` - Reassign tasks
   - `asana tasks update` - Modify task details

2. **Reporting**
   - `asana summary` - Task counts by assignee/status
   - `asana tasks list -m -d overdue` - Find overdue tasks

3. **Project Overview**
   - `asana projects list` - Browse projects
   - `asana projects get` - Project details with task counts

4. **Attachments**
   - `asana attachments list <task_gid>` - List attachments on a task
   - `asana attachments upload <task_gid> <file>` - Upload a file to a task
   - `asana attachments download <attachment_gid>` - Download an attachment
   - `asana attachments delete <attachment_gid>` - Delete an attachment

5. **User Management**
   - `asana users list` - List team members
   - `asana users me` - Current user info
