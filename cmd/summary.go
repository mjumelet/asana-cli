package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/mauricejumelet/asana-cli/internal/api"
)

type SummaryCmd struct {
	Project string `short:"p" help:"Filter by project GID"`
	JSON    bool   `short:"j" help:"Output as JSON"`
}

func (c *SummaryCmd) Run(client *api.Client) error {
	summary, err := client.GetTaskSummary(c.Project)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(summary)
	}

	fmt.Println("Task Summary")
	fmt.Println("============")
	fmt.Printf("Total Tasks:     %d\n", summary.TotalTasks)
	fmt.Printf("Open Tasks:      %d\n", summary.OpenTasks)
	fmt.Printf("Completed Tasks: %d\n", summary.CompletedTasks)
	fmt.Printf("Overdue Tasks:   %d\n", summary.OverdueTasks)
	fmt.Printf("Unassigned:      %d\n", summary.Unassigned)

	if len(summary.ByAssignee) > 0 {
		fmt.Println("\nTasks by Assignee")
		fmt.Println("-----------------")

		// Sort assignees by task count (descending)
		type assigneeCount struct {
			Name  string
			Count int
		}
		var sorted []assigneeCount
		for name, count := range summary.ByAssignee {
			sorted = append(sorted, assigneeCount{name, count})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Count > sorted[j].Count
		})

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ASSIGNEE\tTASKS")
		fmt.Fprintln(w, "--------\t-----")
		for _, ac := range sorted {
			fmt.Fprintf(w, "%s\t%d\n", ac.Name, ac.Count)
		}
		w.Flush()
	}

	return nil
}
