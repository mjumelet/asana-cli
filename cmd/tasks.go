package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mauricejumelet/asana-cli/internal/api"
)

type TasksCmd struct {
	List    TasksListCmd    `cmd:"" help:"List tasks"`
	Get     TasksGetCmd     `cmd:"" help:"Get a task by ID"`
	Comment TasksCommentCmd `cmd:"" help:"Add a comment to a task"`
	Search  TasksSearchCmd  `cmd:"" help:"Search for tasks"`
}

type TasksListCmd struct {
	Project  string `short:"p" help:"Filter by project GID"`
	Assignee string `short:"a" help:"Filter by assignee (use 'me' for yourself)"`
	All      bool   `help:"Include completed tasks"`
	Limit    int    `short:"l" default:"25" help:"Maximum number of tasks to return"`
	JSON     bool   `short:"j" help:"Output as JSON"`
}

func (c *TasksListCmd) Run(client *api.Client) error {
	tasks, err := client.ListTasks(c.Project, c.Assignee, c.All, c.Limit)
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
	return nil
}

type TasksGetCmd struct {
	TaskGID string `arg:"" help:"Task GID to retrieve"`
	JSON    bool   `short:"j" help:"Output as JSON"`
}

func (c *TasksGetCmd) Run(client *api.Client) error {
	task, err := client.GetTask(c.TaskGID)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(task)
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
	Limit int    `short:"l" default:"25" help:"Maximum number of tasks to return"`
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
