package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/malleatus/tamjaweb/internal/browser"
)

// the browser profile to use
var profile string

// bookmarksCmd represents the bookmarks command
var bookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "list all bookmarks",
	Run: func(cmd *cobra.Command, args []string) {
		allBookmarks, err := browser.GetAllBookmarks(profile)
		if err != nil {
			log.Error("Failed to get bookmarks", "error", err)
			return
		}

		// Initialize a tab writer for pretty formatting
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if _, err := fmt.Fprintln(w, "Browser\tTitle\tURL\tFolder\tDate Added"); err != nil {
			log.Error("Failed to write header", "error", err)
			return
		}
		if _, err := fmt.Fprintln(w, "-------\t-----\t---\t------\t----------"); err != nil {
			log.Error("Failed to write separator", "error", err)
			return
		}

		// Print bookmarks from each browser
		for browserName, bookmarks := range allBookmarks {
			for _, bookmark := range bookmarks {
				if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					browserName,
					bookmark.Title,
					bookmark.URL,
					bookmark.FolderPath,
					bookmark.DateAdded.Format("2006-01-02 15:04:05"),
				); err != nil {
					log.Error("Failed to write bookmark", "error", err)
					return
				}
			}
		}

		if err := w.Flush(); err != nil {
			log.Error("Failed to flush output", "error", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(bookmarksCmd)

	// Add the profile flag
	bookmarksCmd.Flags().StringVar(&profile, "profile", "Default", "Browser profile to use")
}
