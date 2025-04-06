package github

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	github "github.com/malleatus/tamjaweb/internal/github"
)

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

	return cmd
}
