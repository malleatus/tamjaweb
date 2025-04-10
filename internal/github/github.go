package github

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v70/github"
	"github.com/olekukonko/tablewriter"
)

type Options struct {
	User string
}

// Star represents a starred repository on GitHub
type Star struct {
	Stargazer   string
	Repo        string
	Description string
	URL         string
	StarredAt   string
}

var BuildGitHubClient = func() *github.Client {
	return github.NewClient(nil)
}

func GetAllStars(user string) ([]Star, error) {
	stars, err := GetCachedStars()
	if err != nil {
		return nil, fmt.Errorf("error fetching cached stars: %v", err)
	}

	// FIXME: need to filter these to the specified user
	if len(stars) == 0 {
		// no cached stars, do the lookup blocking
		stars, err = fetchStars(user)
		if err != nil {
			return nil, fmt.Errorf("error fetching stars from GitHub: %v", err)
		}

		err = WriteCachedStars(user, stars)
		if err != nil {
			return nil, fmt.Errorf("error writing stars to cache: %v", err)
		}
	}

	return stars, nil
}

func fetchStars(user string) ([]Star, error) {
	var stars []Star

	ctx := context.Background()
	client := BuildGitHubClient()

	starredRepos, _, err := client.Activity.ListStarred(ctx, user, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching starred repositories: %v", err)
	}

	for _, starred := range starredRepos {
		if starred.Repository != nil {
			var repo, description, repoURL, starredAt string

			repo = *starred.Repository.FullName
			repoURL = *starred.Repository.HTMLURL
			starredAt = starred.StarredAt.Format(time.DateOnly)

			if starred.Repository.Description != nil {
				description = *starred.Repository.Description
			}

			stars = append(stars, Star{
				Stargazer:   user,
				Repo:        repo,
				Description: description,
				URL:         repoURL,
				StarredAt:   starredAt,
			})
		}
	}

	return stars, nil
}

func PrintStars(stars []Star) (string, error) {
	if len(stars) == 0 {
		return "No stars found", nil
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	table.SetHeader([]string{"Stargazer", "Repository", "Description", "URL"})
	table.SetAutoWrapText(true)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
	})
	table.SetColWidth(50)

	// Add data rows
	for _, star := range stars {
		table.Append([]string{
			star.Stargazer,
			star.Repo,
			star.Description,
			star.URL,
		})
	}

	table.Render()

	return buf.String(), nil
}

// ExecRunner is an interface for executing commands, so we can inject a mock
// during tests.
type ExecRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// RealExecRunner uses the actual `exec.Command`.
type RealExecRunner struct{}

var DefaultExecRunner ExecRunner = &RealExecRunner{}

func (r *RealExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// GetGitHubToken runs `gh auth token` and returns the trimmed output.
func GetGitHubToken() (string, error) {
	runner := DefaultExecRunner
	out, err := runner.Run("gh", "auth", "token")
	if err != nil {
		return "", fmt.Errorf("failed to run gh auth token: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
