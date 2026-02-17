package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/mauricejumelet/asana-cli/cmd"
	"github.com/mauricejumelet/asana-cli/internal/api"
	"github.com/mauricejumelet/asana-cli/internal/config"
)

var version = "1.0.0"

var CLI struct {
	// Global flags
	Config string `short:"c" help:"Path to config file (.env format)" type:"path"`

	// Commands
	Tasks     cmd.TasksCmd    `cmd:"" help:"Manage tasks"`
	Projects  cmd.ProjectsCmd `cmd:"" help:"Manage projects"`
	Users     cmd.UsersCmd    `cmd:"" help:"Manage users"`
	Summary   cmd.SummaryCmd  `cmd:"" help:"Show task summary and statistics"`
	Configure ConfigureCmd    `cmd:"" help:"Show configuration help"`
}

type ConfigureCmd struct{}

func (c *ConfigureCmd) Run() error {
	config.PrintConfigHelp()
	return nil
}

func main() {
	// Handle version flag early
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--version" {
			fmt.Printf("asana-cli v%s\n", version)
			return
		}
	}

	ctx := kong.Parse(&CLI,
		kong.Name("asana"),
		kong.Description("A command-line interface for Asana (v"+version+")"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Commands that don't need the API client
	switch ctx.Command() {
	case "configure":
		err := ctx.Run()
		ctx.FatalIfErrorf(err)
		return
	}

	// Load configuration
	cfg, err := config.Load(CLI.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create API client
	client := api.NewClient(cfg)

	// Run the command with the client
	err = ctx.Run(client)
	ctx.FatalIfErrorf(err)
}
