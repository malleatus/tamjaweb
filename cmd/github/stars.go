package github

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	github "github.com/malleatus/tamjaweb/internal/github"
)

func NewStarsSearchCommand(opts *github.Options) *cobra.Command {
	var searchTerm string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for stars",
		Run: func(cmd *cobra.Command, args []string) {
			if searchTerm == "" && len(args) == 0 {
				log.Error("Search term is required")
				return
			}

			// Use args as search term if not provided via flag
			if searchTerm == "" && len(args) > 0 {
				searchTerm = strings.Join(args, " ")
			}

			allStars, err := github.GetAllStars(opts.User)
			if err != nil {
				log.Error("Failed to get stars", "error", err)
				return
			}

			filteredStars := github.FilterStarsByTerm(allStars, searchTerm)

			formattedOutput, err := github.PrintStars(filteredStars)
			if err != nil {
				log.Error("Failed to format stars", "error", err)
				return
			}
			fmt.Print(formattedOutput)
		},
	}
	cmd.Flags().StringVar(&searchTerm, "term", "", "Term to search for in bookmarks")

	return cmd
}

func NewStarsListCommand(opts *github.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all stars",
		Run: func(cmd *cobra.Command, args []string) {
			allStars, err := github.GetAllStars(opts.User)
			if err != nil {
				log.Error("Failed to get stars", "error", err)
				return
			}

			formattedOutput, err := github.PrintStars(allStars)
			if err != nil {
				log.Error("Failed to format stars", "error", err)
				return
			}
			fmt.Print(formattedOutput)
		},
	}

	return cmd
}

func NewStarsCommand(opts *github.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stars",
		Short: "Work with GitHub stars",
	}

	cmd.AddCommand(NewStarsListCommand(opts))
	cmd.AddCommand(NewStarsSearchCommand(opts))

	return cmd
}
