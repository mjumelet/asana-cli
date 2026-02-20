package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mauricejumelet/asana-cli/internal/api"
)

type TasksCmd struct {
	List     TasksListCmd     `cmd:"" help:"List tasks"`
	Get      TasksGetCmd      `cmd:"" help:"Get a task by ID"`
	Create   TasksCreateCmd   `cmd:"" help:"Create a new task"`
	Complete TasksCompleteCmd `cmd:"" help:"Mark a task as complete"`
	Reopen   TasksReopenCmd   `cmd:"" help:"Reopen a completed task"`
	Update   TasksUpdateCmd   `cmd:"" help:"Update a task"`
	Delete   TasksDeleteCmd   `cmd:"" help:"Delete a task"`
	Comment  TasksCommentCmd  `cmd:"" help:"Add a comment to a task"`
	Search   TasksSearchCmd   `cmd:"" help:"Search for tasks"`
}

type TasksListCmd struct {
	// Shortcut flags
	Mine bool `short:"m" help:"Show only tasks assigned to me (shortcut for -a me)"`

	// Filter flags
	Project  string `short:"p" help:"Filter by project GID or name"`
	Assignee string `short:"a" help:"Filter by assignee GID (use 'me' for yourself)"`
	Tag      string `short:"t" help:"Filter by tag GID"`
	Due      string `short:"d" help:"Filter by due date: today, tomorrow, week, overdue, or YYYY-MM-DD"`

	// Display flags
	All   bool `help:"Include completed tasks"`
	Limit int  `short:"l" default:"100" help:"Maximum number of tasks to return"`
	Sort  string `short:"s" default:"due_date" help:"Sort by: due_date, created_at, modified_at"`
	JSON  bool `short:"j" help:"Output as JSON"`
}

func (c *TasksListCmd) Run(client *api.Client) error {
	// Handle --mine shortcut
	assignee := c.Assignee
	if c.Mine {
		assignee = "me"
	}

	opts := api.TaskListOptions{
		Project:       c.Project,
		Assignee:      assignee,
		Tag:           c.Tag,
		Due:           c.Due,
		IncludeCompleted: c.All,
		Limit:         c.Limit,
		SortBy:        c.Sort,
	}

	tasks, err := client.ListTasks(opts)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(tasks)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GID\tNAME\tDUE\tASSIGNEE\tPROJECT")
	fmt.Fprintln(w, "---\t----\t---\t--------\t-------")

	for _, task := range tasks {
		assignee := "-"
		if task.Assignee != nil {
			assignee = task.Assignee.Name
		}

		due := "-"
		if task.DueOn != "" {
			due = task.DueOn
		}

		project := "-"
		if len(task.Projects) > 0 {
			project = task.Projects[0].Name
		}

		name := truncate(task.Name, 50)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", task.GID, name, due, assignee, project)
	}

	w.Flush()

	if len(tasks) >= c.Limit {
		fmt.Printf("\n(Showing %d tasks, use -l to increase limit)\n", c.Limit)
	}

	return nil
}

type TasksGetCmd struct {
	TaskGID  string `arg:"" help:"Task GID to retrieve"`
	Comments bool   `help:"Include comments and activity"`
	JSON     bool   `short:"j" help:"Output as JSON"`
}

func (c *TasksGetCmd) Run(client *api.Client) error {
	task, err := client.GetTask(c.TaskGID)
	if err != nil {
		return err
	}

	// Fetch comments if requested
	var stories []api.Story
	if c.Comments {
		stories, err = client.GetTaskStories(c.TaskGID)
		if err != nil {
			return err
		}
	}

	// Fetch attachments
	attachments, err := client.ListAttachments(c.TaskGID)
	if err != nil {
		return err
	}

	if c.JSON {
		if c.Comments {
			return printJSON(map[string]interface{}{
				"task":        task,
				"comments":    stories,
				"attachments": attachments,
			})
		}
		return printJSON(map[string]interface{}{
			"task":        task,
			"attachments": attachments,
		})
	}

	fmt.Printf("Task: %s\n", task.Name)
	fmt.Printf("GID: %s\n", task.GID)
	fmt.Printf("Status: %s\n", statusString(task.Completed))

	if task.Assignee != nil {
		fmt.Printf("Assignee: %s", task.Assignee.Name)
		if task.Assignee.Email != "" {
			fmt.Printf(" <%s>", task.Assignee.Email)
		}
		fmt.Println()
	}

	if task.DueOn != "" {
		fmt.Printf("Due: %s\n", task.DueOn)
	}

	if len(task.Projects) > 0 {
		projects := make([]string, len(task.Projects))
		for i, p := range task.Projects {
			projects[i] = p.Name
		}
		fmt.Printf("Projects: %s\n", strings.Join(projects, ", "))
	}

	if len(task.Tags) > 0 {
		tags := make([]string, len(task.Tags))
		for i, t := range task.Tags {
			tags[i] = t.Name
		}
		fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}

	fmt.Printf("Created: %s\n", task.CreatedAt)
	fmt.Printf("Modified: %s\n", task.ModifiedAt)

	if task.Permalink != "" {
		fmt.Printf("URL: %s\n", task.Permalink)
	}

	if task.Notes != "" {
		fmt.Printf("\nDescription:\n%s\n", task.Notes)
	}

	// Display attachments
	if len(attachments) > 0 {
		fmt.Printf("\nAttachments (%d):\n", len(attachments))
		for _, a := range attachments {
			size := ""
			if a.Size > 0 {
				size = fmt.Sprintf(" (%s)", formatSize(a.Size))
			}
			fmt.Printf("  - %s%s [%s]\n", a.Name, size, a.GID)
		}
	}

	// Display comments/activity
	if c.Comments && len(stories) > 0 {
		fmt.Printf("\nComments & Activity (%d):\n", len(stories))
		fmt.Println(strings.Repeat("-", 40))
		for _, story := range stories {
			author := "Unknown"
			if story.CreatedBy != nil {
				author = story.CreatedBy.Name
			}
			timestamp := story.CreatedAt
			if len(timestamp) > 10 {
				timestamp = timestamp[:10]
			}
			fmt.Printf("[%s] %s\n", timestamp, author)
			if story.Text != "" {
				fmt.Printf("  %s\n", story.Text)
			}
			fmt.Println()
		}
	}

	return nil
}

type TasksCommentCmd struct {
	TaskGID string `arg:"" help:"Task GID to comment on"`
	Message string `arg:"" help:"Comment message (use --html for rich text)"`
	HTML    bool   `help:"Treat message as HTML rich text"`
}

func (c *TasksCommentCmd) Run(client *api.Client) error {
	message := c.Message

	// If HTML flag is set but message doesn't have body tags, wrap it
	if c.HTML && !strings.Contains(message, "<body>") {
		message = "<body>" + message + "</body>"
	}

	story, err := client.AddComment(c.TaskGID, message, c.HTML)
	if err != nil {
		return err
	}

	fmt.Printf("Comment added successfully (ID: %s)\n", story.GID)
	fmt.Printf("Created at: %s\n", story.CreatedAt)

	return nil
}

type TasksSearchCmd struct {
	Query string `arg:"" help:"Search query"`
	Limit int    `short:"l" default:"100" help:"Maximum number of tasks to return"`
	JSON  bool   `short:"j" help:"Output as JSON"`
}

func (c *TasksSearchCmd) Run(client *api.Client) error {
	tasks, err := client.SearchTasks(c.Query, c.Limit)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(tasks)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GID\tNAME\tDUE\tASSIGNEE\tPROJECT")
	fmt.Fprintln(w, "---\t----\t---\t--------\t-------")

	for _, task := range tasks {
		assignee := "-"
		if task.Assignee != nil {
			assignee = task.Assignee.Name
		}

		due := "-"
		if task.DueOn != "" {
			due = task.DueOn
		}

		project := "-"
		if len(task.Projects) > 0 {
			project = task.Projects[0].Name
		}

		name := truncate(task.Name, 50)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", task.GID, name, due, assignee, project)
	}

	w.Flush()

	if len(tasks) >= c.Limit {
		fmt.Printf("\n(Showing %d tasks, use -l to increase limit)\n", c.Limit)
	}

	return nil
}

func statusString(completed bool) string {
	if completed {
		return "Completed"
	}
	return "Open"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// TasksCreateCmd creates a new task
type TasksCreateCmd struct {
	Name     string   `arg:"" help:"Task name"`
	Notes    string   `short:"n" help:"Task description"`
	Assignee string   `short:"a" help:"Assignee GID or 'me'"`
	Due      string   `short:"d" help:"Due date (YYYY-MM-DD)"`
	Project  string   `short:"p" help:"Project GID to add task to"`
	JSON     bool     `short:"j" help:"Output as JSON"`
}

func (c *TasksCreateCmd) Run(client *api.Client) error {
	opts := api.CreateTaskOptions{
		Name:     c.Name,
		Notes:    c.Notes,
		Assignee: c.Assignee,
		DueOn:    c.Due,
	}

	if c.Project != "" {
		opts.Projects = []string{c.Project}
	}

	task, err := client.CreateTask(opts)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(task)
	}

	fmt.Printf("Task created successfully!\n")
	fmt.Printf("GID: %s\n", task.GID)
	fmt.Printf("Name: %s\n", task.Name)
	if task.Permalink != "" {
		fmt.Printf("URL: %s\n", task.Permalink)
	}

	return nil
}

// TasksCompleteCmd marks a task as complete
type TasksCompleteCmd struct {
	TaskGID string `arg:"" help:"Task GID to complete"`
}

func (c *TasksCompleteCmd) Run(client *api.Client) error {
	task, err := client.CompleteTask(c.TaskGID)
	if err != nil {
		return err
	}

	fmt.Printf("Task completed: %s\n", task.Name)
	return nil
}

// TasksReopenCmd reopens a completed task
type TasksReopenCmd struct {
	TaskGID string `arg:"" help:"Task GID to reopen"`
}

func (c *TasksReopenCmd) Run(client *api.Client) error {
	task, err := client.ReopenTask(c.TaskGID)
	if err != nil {
		return err
	}

	fmt.Printf("Task reopened: %s\n", task.Name)
	return nil
}

// TasksUpdateCmd updates an existing task
type TasksUpdateCmd struct {
	TaskGID  string `arg:"" help:"Task GID to update"`
	Name     string `short:"n" help:"New task name"`
	Notes    string `help:"New task description"`
	Assignee string `short:"a" help:"New assignee GID or 'me'"`
	Due      string `short:"d" help:"New due date (YYYY-MM-DD)"`
	JSON     bool   `short:"j" help:"Output as JSON"`
}

func (c *TasksUpdateCmd) Run(client *api.Client) error {
	opts := api.UpdateTaskOptions{}

	if c.Name != "" {
		opts.Name = &c.Name
	}
	if c.Notes != "" {
		opts.Notes = &c.Notes
	}
	if c.Assignee != "" {
		opts.Assignee = &c.Assignee
	}
	if c.Due != "" {
		opts.DueOn = &c.Due
	}

	task, err := client.UpdateTask(c.TaskGID, opts)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(task)
	}

	fmt.Printf("Task updated: %s\n", task.Name)
	return nil
}

// TasksDeleteCmd deletes a task
type TasksDeleteCmd struct {
	TaskGID string `arg:"" help:"Task GID to delete"`
	Force   bool   `short:"f" help:"Skip confirmation"`
}

func (c *TasksDeleteCmd) Run(client *api.Client) error {
	if !c.Force {
		fmt.Printf("Are you sure you want to delete task %s? [y/N] ", c.TaskGID)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	err := client.DeleteTask(c.TaskGID)
	if err != nil {
		return err
	}

	fmt.Printf("Task %s deleted.\n", c.TaskGID)
	return nil
}
