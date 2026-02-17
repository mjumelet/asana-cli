package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mauricejumelet/asana-cli/internal/config"
)

const baseURL = "https://app.asana.com/api/1.0"

type Client struct {
	httpClient *http.Client
	token      string
	workspace  string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{},
		token:      cfg.Token,
		workspace:  cfg.Workspace,
	}
}

func (c *Client) Workspace() string {
	return c.workspace
}

func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	reqURL := baseURL + endpoint

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

type ErrorResponse struct {
	Errors []struct {
		Message string `json:"message"`
		Help    string `json:"help,omitempty"`
	} `json:"errors"`
}

// Task represents an Asana task
type Task struct {
	GID          string   `json:"gid"`
	Name         string   `json:"name"`
	Notes        string   `json:"notes,omitempty"`
	HTMLNotes    string   `json:"html_notes,omitempty"`
	Completed    bool     `json:"completed"`
	CompletedAt  string   `json:"completed_at,omitempty"`
	DueOn        string   `json:"due_on,omitempty"`
	DueAt        string   `json:"due_at,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	ModifiedAt   string   `json:"modified_at,omitempty"`
	Assignee     *User    `json:"assignee,omitempty"`
	Projects     []Entity `json:"projects,omitempty"`
	Tags         []Entity `json:"tags,omitempty"`
	Permalink    string   `json:"permalink_url,omitempty"`
}

type User struct {
	GID   string `json:"gid"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type Entity struct {
	GID  string `json:"gid"`
	Name string `json:"name,omitempty"`
}

type Project struct {
	GID       string `json:"gid"`
	Name      string `json:"name"`
	Archived  bool   `json:"archived,omitempty"`
	Color     string `json:"color,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Permalink string `json:"permalink_url,omitempty"`
}

type Story struct {
	GID       string `json:"gid"`
	CreatedAt string `json:"created_at"`
	CreatedBy *User  `json:"created_by,omitempty"`
	Text      string `json:"text,omitempty"`
	HTMLText  string `json:"html_text,omitempty"`
	Type      string `json:"type,omitempty"`
}

type TasksResponse struct {
	Data       []Task `json:"data"`
	NextPage   *Page  `json:"next_page,omitempty"`
}

type TaskResponse struct {
	Data Task `json:"data"`
}

type ProjectsResponse struct {
	Data     []Project `json:"data"`
	NextPage *Page     `json:"next_page,omitempty"`
}

type StoryResponse struct {
	Data Story `json:"data"`
}

type Page struct {
	Offset string `json:"offset"`
	Path   string `json:"path"`
	URI    string `json:"uri"`
}

// TaskListOptions contains all filtering options for listing tasks
type TaskListOptions struct {
	Project          string // Project GID
	Assignee         string // Assignee GID or "me"
	Tag              string // Tag GID
	Due              string // Due filter: today, tomorrow, week, overdue, or YYYY-MM-DD
	IncludeCompleted bool   // Include completed tasks
	Limit            int    // Maximum results
	SortBy           string // Sort field: due_date, created_at, modified_at
}

// ListTasks returns tasks filtered by the given options
func (c *Client) ListTasks(opts TaskListOptions) ([]Task, error) {
	// Use the search API for advanced filtering
	params := url.Values{}

	// Project filter
	if opts.Project != "" {
		params.Set("projects.any", opts.Project)
	}

	// Assignee filter
	if opts.Assignee != "" {
		params.Set("assignee.any", opts.Assignee)
	}

	// Tag filter
	if opts.Tag != "" {
		params.Set("tags.any", opts.Tag)
	}

	// Due date filter
	if opts.Due != "" {
		c.applyDueFilter(params, opts.Due)
	}

	// Completed filter
	if !opts.IncludeCompleted {
		params.Set("completed", "false")
	}

	// Limit
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("limit", "100")
	}

	// Sort
	if opts.SortBy != "" {
		params.Set("sort_by", opts.SortBy)
		params.Set("sort_ascending", "true")
	}

	// Exclude subtasks for cleaner output
	params.Set("is_subtask", "false")

	params.Set("opt_fields", "gid,name,completed,due_on,assignee,assignee.name,projects,projects.name,tags,tags.name,permalink_url")

	endpoint := fmt.Sprintf("/workspaces/%s/tasks/search?%s", c.workspace, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp TasksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return resp.Data, nil
}

// applyDueFilter adds due date parameters based on the filter string
func (c *Client) applyDueFilter(params url.Values, due string) {
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	weekEnd := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	switch due {
	case "today":
		params.Set("due_on", today)
	case "tomorrow":
		params.Set("due_on", tomorrow)
	case "week":
		params.Set("due_on.before", weekEnd)
		params.Set("due_on.after", today)
	case "overdue":
		params.Set("due_on.before", today)
	default:
		// Assume it's a date in YYYY-MM-DD format
		params.Set("due_on", due)
	}
}

// SearchTasks searches for tasks in the workspace
func (c *Client) SearchTasks(query string, limit int) ([]Task, error) {
	params := url.Values{}

	if query != "" {
		params.Set("text", query)
	}

	params.Set("completed", "false")

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "100")
	}

	params.Set("opt_fields", "gid,name,completed,due_on,assignee,assignee.name,projects,projects.name,permalink_url")

	endpoint := fmt.Sprintf("/workspaces/%s/tasks/search?%s", c.workspace, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp TasksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return resp.Data, nil
}

// GetTask returns a single task by GID
func (c *Client) GetTask(gid string) (*Task, error) {
	params := url.Values{}
	params.Set("opt_fields", "gid,name,notes,html_notes,completed,completed_at,due_on,due_at,created_at,modified_at,assignee,assignee.name,assignee.email,projects,projects.name,tags,tags.name,permalink_url")

	endpoint := fmt.Sprintf("/tasks/%s?%s", gid, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp TaskResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// AddComment adds a comment (story) to a task
// The comment can be plain text or HTML for rich text formatting
// For rich text, wrap content in <body> tags and use supported HTML:
// <strong>, <em>, <u>, <s>, <code>, <ol>, <ul>, <li>, <a>, <blockquote>, <pre>
func (c *Client) AddComment(taskGID, comment string, isHTML bool) (*Story, error) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{},
	}

	if isHTML {
		payload["data"].(map[string]interface{})["html_text"] = comment
	} else {
		payload["data"].(map[string]interface{})["text"] = comment
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	endpoint := fmt.Sprintf("/tasks/%s/stories", taskGID)
	body, err := c.doRequest("POST", endpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var resp StoryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// ListProjects returns projects in the workspace
func (c *Client) ListProjects(archived bool, limit int) ([]Project, error) {
	params := url.Values{}
	params.Set("archived", fmt.Sprintf("%t", archived))

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "100")
	}

	params.Set("opt_fields", "gid,name,archived,color,created_at,permalink_url")

	endpoint := fmt.Sprintf("/workspaces/%s/projects?%s", c.workspace, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp ProjectsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return resp.Data, nil
}
