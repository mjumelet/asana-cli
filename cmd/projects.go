package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mauricejumelet/asana-cli/internal/api"
)

type ProjectsCmd struct {
	List ProjectsListCmd `cmd:"" help:"List projects in the workspace"`
}

type ProjectsListCmd struct {
	Archived bool `short:"a" help:"Include archived projects"`
	Limit    int  `short:"l" default:"50" help:"Maximum number of projects to return"`
	JSON     bool `short:"j" help:"Output as JSON"`
}

func (c *ProjectsListCmd) Run(client *api.Client) error {
	projects, err := client.ListProjects(c.Archived, c.Limit)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(projects)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GID\tNAME\tARCHIVED\tCREATED")
	fmt.Fprintln(w, "---\t----\t--------\t-------")

	for _, project := range projects {
		archived := "No"
		if project.Archived {
			archived = "Yes"
		}

		created := "-"
		if project.CreatedAt != "" {
			created = project.CreatedAt[:10] // Just the date part
		}

		name := truncate(project.Name, 40)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", project.GID, name, archived, created)
	}

	w.Flush()
	return nil
}
