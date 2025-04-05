package bookmarks

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	internalBookmarks "github.com/malleatus/tamjaweb/internal/bookmarks"
	"github.com/malleatus/tamjaweb/internal/browser"
)

var searchTerm string

func NewSearchCommand(opts *internalBookmarks.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for bookmarks",
		Run: func(cmd *cobra.Command, args []string) {
			if searchTerm == "" && len(args) == 0 {
				log.Error("Search term is required")
				return
			}

			// Use args as search term if not provided via flag
			if searchTerm == "" && len(args) > 0 {
				searchTerm = strings.Join(args, " ")
			}

			allBookmarks, err := browser.GetAllBookmarks(opts.Profile)
			if err != nil {
				log.Error("Failed to get bookmarks", "error", err)
				return
			}

			filteredBookmarks := internalBookmarks.FilterBookmarksByTerm(allBookmarks, searchTerm)
			formattedOutput, err := internalBookmarks.PrintBookmarks(filteredBookmarks)
			if err != nil {
				log.Error("Failed to format bookmarks", "error", err)
				return
			}
			fmt.Print(formattedOutput)
		},
	}
	cmd.Flags().StringVar(&searchTerm, "term", "", "Term to search for in bookmarks")

	return cmd
}
