package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mauricejumelet/asana-cli/internal/api"
)

type UsersCmd struct {
	List UsersListCmd `cmd:"" help:"List users in the workspace"`
	Me   UsersMeCmd   `cmd:"" help:"Show current user"`
}

type UsersListCmd struct {
	JSON bool `short:"j" help:"Output as JSON"`
}

func (c *UsersListCmd) Run(client *api.Client) error {
	users, err := client.ListUsers()
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(users)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GID\tNAME\tEMAIL")
	fmt.Fprintln(w, "---\t----\t-----")

	for _, user := range users {
		email := "-"
		if user.Email != "" {
			email = user.Email
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", user.GID, user.Name, email)
	}

	w.Flush()
	return nil
}

type UsersMeCmd struct {
	JSON bool `short:"j" help:"Output as JSON"`
}

func (c *UsersMeCmd) Run(client *api.Client) error {
	user, err := client.GetMe()
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(user)
	}

	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("GID: %s\n", user.GID)
	if user.Email != "" {
		fmt.Printf("Email: %s\n", user.Email)
	}

	return nil
}
