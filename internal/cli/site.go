package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/site"
)

// newSite builds the phone checklist web page. It is a maintainer command
// (the page is published by CI), so it is hidden from the help.
func newSite() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:    "site",
		Short:  "Build the phone checklist as a static web page",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			html, err := site.Build(checks.FS)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(out, 0o755); err != nil {
				return err
			}
			file := filepath.Join(out, "index.html")
			if err := os.WriteFile(file, html, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%d KB)\n", file, len(html)/1024)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "_site", "folder to write index.html into")
	return cmd
}
