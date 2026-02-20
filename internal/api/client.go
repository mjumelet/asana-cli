package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

type StoriesResponse struct {
	Data []Story `json:"data"`
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

// DeleteStory deletes a comment (story) from a task
func (c *Client) DeleteStory(storyGID string) error {
	endpoint := fmt.Sprintf("/stories/%s", storyGID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

// GetTaskStories returns all stories (comments and activity) for a task
func (c *Client) GetTaskStories(taskGID string) ([]Story, error) {
	params := url.Values{}
	params.Set("opt_fields", "gid,created_at,created_by,created_by.name,text,html_text,type,resource_subtype")

	endpoint := fmt.Sprintf("/tasks/%s/stories?%s", taskGID, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp StoriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return resp.Data, nil
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

// CreateTaskOptions contains options for creating a new task
type CreateTaskOptions struct {
	Name      string
	Notes     string
	Assignee  string
	DueOn     string
	Projects  []string
	Tags      []string
	Parent    string // For subtasks
}

// CreateTask creates a new task in the workspace
func (c *Client) CreateTask(opts CreateTaskOptions) (*Task, error) {
	data := map[string]interface{}{
		"name": opts.Name,
	}

	if opts.Notes != "" {
		data["notes"] = opts.Notes
	}
	if opts.Assignee != "" {
		data["assignee"] = opts.Assignee
	}
	if opts.DueOn != "" {
		data["due_on"] = opts.DueOn
	}
	if len(opts.Projects) > 0 {
		data["projects"] = opts.Projects
	}
	if len(opts.Tags) > 0 {
		data["tags"] = opts.Tags
	}
	if opts.Parent != "" {
		data["parent"] = opts.Parent
	}

	// If no project specified and not a subtask, we need workspace
	if len(opts.Projects) == 0 && opts.Parent == "" {
		data["workspace"] = c.workspace
	}

	payload := map[string]interface{}{"data": data}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	body, err := c.doRequest("POST", "/tasks", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var resp TaskResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// UpdateTaskOptions contains options for updating a task
type UpdateTaskOptions struct {
	Name      *string
	Notes     *string
	Assignee  *string
	DueOn     *string
	Completed *bool
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(taskGID string, opts UpdateTaskOptions) (*Task, error) {
	data := map[string]interface{}{}

	if opts.Name != nil {
		data["name"] = *opts.Name
	}
	if opts.Notes != nil {
		data["notes"] = *opts.Notes
	}
	if opts.Assignee != nil {
		data["assignee"] = *opts.Assignee
	}
	if opts.DueOn != nil {
		data["due_on"] = *opts.DueOn
	}
	if opts.Completed != nil {
		data["completed"] = *opts.Completed
	}

	payload := map[string]interface{}{"data": data}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	endpoint := fmt.Sprintf("/tasks/%s", taskGID)
	body, err := c.doRequest("PUT", endpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var resp TaskResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// CompleteTask marks a task as completed
func (c *Client) CompleteTask(taskGID string) (*Task, error) {
	completed := true
	return c.UpdateTask(taskGID, UpdateTaskOptions{Completed: &completed})
}

// ReopenTask marks a task as not completed
func (c *Client) ReopenTask(taskGID string) (*Task, error) {
	completed := false
	return c.UpdateTask(taskGID, UpdateTaskOptions{Completed: &completed})
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(taskGID string) error {
	endpoint := fmt.Sprintf("/tasks/%s", taskGID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

// UsersResponse represents the API response for users
type UsersResponse struct {
	Data []User `json:"data"`
}

// UserResponse represents the API response for a single user
type UserResponse struct {
	Data User `json:"data"`
}

// ListUsers returns all users in the workspace
func (c *Client) ListUsers() ([]User, error) {
	params := url.Values{}
	params.Set("opt_fields", "gid,name,email")

	endpoint := fmt.Sprintf("/workspaces/%s/users?%s", c.workspace, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp UsersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return resp.Data, nil
}

// GetMe returns the current authenticated user
func (c *Client) GetMe() (*User, error) {
	params := url.Values{}
	params.Set("opt_fields", "gid,name,email")

	endpoint := "/users/me?" + params.Encode()
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp UserResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// TaskSummary represents task counts for summary reporting
type TaskSummary struct {
	TotalTasks     int
	CompletedTasks int
	OpenTasks      int
	OverdueTasks   int
	ByAssignee     map[string]int
	Unassigned     int
}

// GetTaskSummary returns a summary of tasks in the workspace
func (c *Client) GetTaskSummary(projectGID string) (*TaskSummary, error) {
	params := url.Values{}
	if projectGID != "" {
		params.Set("projects.any", projectGID)
	} else {
		// Search API requires at least one filter - use modified in last year as broad filter
		params.Set("modified_on.after", time.Now().AddDate(-1, 0, 0).Format("2006-01-02"))
	}
	params.Set("limit", "100")
	params.Set("opt_fields", "gid,completed,due_on,assignee,assignee.name")

	endpoint := fmt.Sprintf("/workspaces/%s/tasks/search?%s", c.workspace, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp TasksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	summary := &TaskSummary{
		ByAssignee: make(map[string]int),
	}

	today := time.Now().Format("2006-01-02")

	for _, task := range resp.Data {
		summary.TotalTasks++

		if task.Completed {
			summary.CompletedTasks++
		} else {
			summary.OpenTasks++

			if task.DueOn != "" && task.DueOn < today {
				summary.OverdueTasks++
			}
		}

		if task.Assignee != nil {
			summary.ByAssignee[task.Assignee.Name]++
		} else {
			summary.Unassigned++
		}
	}

	return summary, nil
}

// Attachment represents an Asana attachment
type Attachment struct {
	GID             string  `json:"gid"`
	Name            string  `json:"name"`
	ResourceSubtype string  `json:"resource_subtype,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	DownloadURL     string  `json:"download_url,omitempty"`
	PermanentURL    string  `json:"permanent_url,omitempty"`
	ViewURL         string  `json:"view_url,omitempty"`
	Host            string  `json:"host,omitempty"`
	Size            int64   `json:"size,omitempty"`
	Parent          *Entity `json:"parent,omitempty"`
}

type AttachmentResponse struct {
	Data Attachment `json:"data"`
}

type AttachmentsResponse struct {
	Data []Attachment `json:"data"`
}

// ListAttachments returns attachments on a task
func (c *Client) ListAttachments(taskGID string) ([]Attachment, error) {
	params := url.Values{}
	params.Set("opt_fields", "gid,name,resource_subtype,created_at,host,size")

	endpoint := fmt.Sprintf("/tasks/%s/attachments?%s", taskGID, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp AttachmentsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return resp.Data, nil
}

// GetAttachment returns a single attachment by GID
func (c *Client) GetAttachment(attachmentGID string) (*Attachment, error) {
	params := url.Values{}
	params.Set("opt_fields", "gid,name,resource_subtype,created_at,download_url,permanent_url,view_url,host,size,parent,parent.name")

	endpoint := fmt.Sprintf("/attachments/%s?%s", attachmentGID, params.Encode())
	body, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp AttachmentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// doMultipartRequest sends a multipart/form-data request with a file upload
func (c *Client) doMultipartRequest(endpoint, filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copying file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	reqURL := baseURL + endpoint
	req, err := http.NewRequest("POST", reqURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

// UploadAttachment uploads a file to a task
func (c *Client) UploadAttachment(taskGID, filePath string) (*Attachment, error) {
	endpoint := fmt.Sprintf("/tasks/%s/attachments", taskGID)
	body, err := c.doMultipartRequest(endpoint, filePath)
	if err != nil {
		return nil, err
	}

	var resp AttachmentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Data, nil
}

// DownloadAttachment downloads an attachment file to disk
func (c *Client) DownloadAttachment(attachment *Attachment, destPath string) error {
	if attachment.DownloadURL == "" {
		return fmt.Errorf("attachment has no download URL")
	}

	resp, err := c.httpClient.Get(attachment.DownloadURL)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// DeleteAttachment deletes an attachment
func (c *Client) DeleteAttachment(attachmentGID string) error {
	endpoint := fmt.Sprintf("/attachments/%s", attachmentGID)
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}
