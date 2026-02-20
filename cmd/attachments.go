package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mauricejumelet/asana-cli/internal/api"
)

type AttachmentsCmd struct {
	List     AttachmentsListCmd     `cmd:"" help:"List attachments on a task"`
	Get      AttachmentsGetCmd      `cmd:"" help:"Get attachment details"`
	Upload   AttachmentsUploadCmd   `cmd:"" help:"Upload a file to a task"`
	Download AttachmentsDownloadCmd `cmd:"" help:"Download an attachment"`
	Delete   AttachmentsDeleteCmd   `cmd:"" help:"Delete an attachment"`
}

type AttachmentsListCmd struct {
	TaskGID string `arg:"" help:"Task GID to list attachments for"`
	JSON    bool   `short:"j" help:"Output as JSON"`
}

func (c *AttachmentsListCmd) Run(client *api.Client) error {
	attachments, err := client.ListAttachments(c.TaskGID)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(attachments)
	}

	if len(attachments) == 0 {
		fmt.Println("No attachments found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GID\tNAME\tSIZE\tCREATED\tHOST")
	fmt.Fprintln(w, "---\t----\t----\t-------\t----")

	for _, a := range attachments {
		created := "-"
		if a.CreatedAt != "" && len(a.CreatedAt) >= 10 {
			created = a.CreatedAt[:10]
		}

		size := "-"
		if a.Size > 0 {
			size = formatSize(a.Size)
		}

		host := "-"
		if a.Host != "" {
			host = a.Host
		}

		name := truncate(a.Name, 50)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", a.GID, name, size, created, host)
	}

	w.Flush()
	return nil
}

type AttachmentsGetCmd struct {
	AttachmentGID string `arg:"" help:"Attachment GID to retrieve"`
	JSON          bool   `short:"j" help:"Output as JSON"`
}

func (c *AttachmentsGetCmd) Run(client *api.Client) error {
	attachment, err := client.GetAttachment(c.AttachmentGID)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(attachment)
	}

	fmt.Printf("Name: %s\n", attachment.Name)
	fmt.Printf("GID: %s\n", attachment.GID)

	if attachment.ResourceSubtype != "" {
		fmt.Printf("Type: %s\n", attachment.ResourceSubtype)
	}
	if attachment.Host != "" {
		fmt.Printf("Host: %s\n", attachment.Host)
	}
	if attachment.Size > 0 {
		fmt.Printf("Size: %s\n", formatSize(attachment.Size))
	}
	if attachment.CreatedAt != "" {
		fmt.Printf("Created: %s\n", attachment.CreatedAt)
	}
	if attachment.Parent != nil {
		fmt.Printf("Parent: %s (%s)\n", attachment.Parent.Name, attachment.Parent.GID)
	}
	if attachment.DownloadURL != "" {
		fmt.Printf("Download URL: %s\n", attachment.DownloadURL)
	}
	if attachment.PermanentURL != "" {
		fmt.Printf("Permanent URL: %s\n", attachment.PermanentURL)
	}
	if attachment.ViewURL != "" {
		fmt.Printf("View URL: %s\n", attachment.ViewURL)
	}

	return nil
}

type AttachmentsUploadCmd struct {
	TaskGID  string `arg:"" help:"Task GID to attach file to"`
	FilePath string `arg:"" help:"Path to file to upload" type:"path"`
	JSON     bool   `short:"j" help:"Output as JSON"`
}

func (c *AttachmentsUploadCmd) Run(client *api.Client) error {
	attachment, err := client.UploadAttachment(c.TaskGID, c.FilePath)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(attachment)
	}

	fmt.Printf("File uploaded successfully!\n")
	fmt.Printf("GID: %s\n", attachment.GID)
	fmt.Printf("Name: %s\n", attachment.Name)

	return nil
}

type AttachmentsDownloadCmd struct {
	AttachmentGID string `arg:"" help:"Attachment GID to download"`
	Output        string `short:"o" help:"Output file path (defaults to current directory with attachment name)"`
}

func (c *AttachmentsDownloadCmd) Run(client *api.Client) error {
	attachment, err := client.GetAttachment(c.AttachmentGID)
	if err != nil {
		return err
	}

	destPath := c.Output
	if destPath == "" {
		destPath = filepath.Join(".", attachment.Name)
	}

	if err := client.DownloadAttachment(attachment, destPath); err != nil {
		return err
	}

	fmt.Printf("Downloaded: %s\n", destPath)
	return nil
}

type AttachmentsDeleteCmd struct {
	AttachmentGID string `arg:"" help:"Attachment GID to delete"`
	Force         bool   `short:"f" help:"Skip confirmation"`
}

func (c *AttachmentsDeleteCmd) Run(client *api.Client) error {
	if !c.Force {
		fmt.Printf("Are you sure you want to delete attachment %s? [y/N] ", c.AttachmentGID)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := client.DeleteAttachment(c.AttachmentGID); err != nil {
		return err
	}

	fmt.Printf("Attachment %s deleted.\n", c.AttachmentGID)
	return nil
}

func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
