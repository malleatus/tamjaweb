package bookmarks

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	internalBookmarks "github.com/malleatus/tamjaweb/internal/bookmarks"
	"github.com/malleatus/tamjaweb/internal/browser"
)

func NewListCommand(opts *internalBookmarks.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all bookmarks",
		Run: func(cmd *cobra.Command, args []string) {
			allBookmarks, err := browser.GetAllBookmarks(opts.Profile)
			if err != nil {
				log.Error("Failed to get bookmarks", "error", err)
				return
			}

			formattedOutput, err := internalBookmarks.PrintBookmarks(allBookmarks)
			if err != nil {
				log.Error("Failed to format bookmarks", "error", err)
				return
			}
			fmt.Print(formattedOutput)
		},
	}

	return cmd
}
