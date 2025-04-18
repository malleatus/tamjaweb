package github

import (
	"errors"
	"os"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/suite"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

type GitHubTestSuite struct {
	suite.Suite

	originalHomeDir           string
	tempHomeDir               string
	originalMaxPages          int
	originalExecRunner        ExecRunner
	originalBuildGitHubClient func() *github.Client
	mockRunner                *mockRunner
}

type mockRunner struct {
	Command      string
	Args         []string
	Output       []byte
	Err          error
	TimesInvoked int
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	m.TimesInvoked++
	m.Command = name
	m.Args = args
	return m.Output, m.Err
}

// SetupTest runs before each test in the suite.
func (s *GitHubTestSuite) SetupTest() {
	s.originalExecRunner = DefaultExecRunner
	s.originalBuildGitHubClient = BuildGitHubClient
	s.mockRunner = &mockRunner{}

	DefaultExecRunner = s.mockRunner

	homeDir := os.Getenv("HOME")

	s.originalHomeDir = homeDir
	tempHomeDir, err := os.MkdirTemp("", "tamjaweb-test-cache")
	s.NoError(err)

	err = os.Setenv("HOME", tempHomeDir)
	s.NoError(err)
	s.tempHomeDir = tempHomeDir

	s.originalMaxPages = MaxPages
	MaxPages = 2
}

// TearDownTest runs after each test in the suite.
func (s *GitHubTestSuite) TearDownTest() {
	MaxPages = s.originalMaxPages
	DefaultExecRunner = s.originalExecRunner
	BuildGitHubClient = s.originalBuildGitHubClient

	err := os.Setenv("HOME", s.originalHomeDir)
	s.NoError(err)
	defer func() {
		err := os.RemoveAll(s.tempHomeDir)
		s.NoError(err)
	}()
}

// TestGetGitHubToken_Success: an example test
func (s *GitHubTestSuite) TestGetGitHubToken_Success() {
	// Mock a successful output from `gh auth token`.
	s.mockRunner.Output = []byte("FAKE_TOKEN\n")

	token, err := GetGitHubToken()
	s.Require().NoError(err)
	s.Equal("FAKE_TOKEN", token)
}

// TestGetGitHubToken_Error: another example test
func (s *GitHubTestSuite) TestGetGitHubToken_Error() {
	s.mockRunner.Err = errors.New("execution failed")

	token, err := GetGitHubToken()
	s.Empty(token)
	s.Error(err)
	s.Contains(err.Error(), "execution failed")
}

func (s *GitHubTestSuite) TestPrintStarsNoStars() {
	output, err := PrintStars([]Star{})
	s.NoError(err)
	s.Equal("No stars found", output)
}

func (s *GitHubTestSuite) TestPrintStarsWithStars() {
	stars := []Star{
		{
			Repo:        "octocat/Hello-World",
			Description: "A test repository",
			URL:         "https://github.com/octocat/Hello-World",
			StarredAt:   "2021-01-01",
		},
		{
			Repo:        "another/repo",
			Description: "Another test repository",
			URL:         "https://github.com/another/repo",
			StarredAt:   "2021-02-02",
		},
	}

	output, err := PrintStars(stars)
	s.NoError(err)

	cupaloy.SnapshotT(s.T(), output)
}

func (s *GitHubTestSuite) TestGetAllStarsWithVCR() {
	mode := recorder.ModeRecordOnce
	if os.Getenv("CI") == "true" {
		mode = recorder.ModeReplayOnly
	}
	opts := []recorder.Option{
		recorder.WithMode(mode),
		// NOTE: without this flag set, vcr will take as long as the original
		// request that was recorded did
		recorder.WithSkipRequestLatency(true),
	}
	r, err := recorder.New("fixtures/get_all_stars", opts...)

	s.NoError(err)
	s.T().Cleanup(func() {
		// Make sure recorder is stopped once done with it.
		if err := r.Stop(); err != nil {
			s.Error(err)
		}
	})

	// NOTE: not using any auth here, so there is nothing to sanitize from the response (in this case)
	client := github.NewClient(r.GetDefaultClient())

	BuildGitHubClient = func() *github.Client {
		return client
	}

	stars, err := GetAllStars("rwjblue")
	s.NoError(err, "Failed to get stars from GitHub API")

	cupaloy.SnapshotT(s.T(), stars)
}

func (s *GitHubTestSuite) TestGetAllStarsFiltersToSpecifiedUser() {
	stars := []Star{
		{
			Stargazer:   "rwjblue",
			Repo:        "malleatus/tamjaweb",
			Description: "A web app",
			URL:         "https://github.com/malleatus/tamjaweb",
			StarredAt:   "2023-01-01",
		},
		{
			Stargazer:   "otheruser",
			Repo:        "otheruser/repo",
			Description: "Another repo",
			URL:         "https://github.com/otheruser/repo",
			StarredAt:   "2023-02-02",
		},
		{
			Stargazer:   "rwjblue",
			Repo:        "emberjs/ember.js",
			Description: "Ember.js framework",
			URL:         "https://github.com/emberjs/ember.js",
			StarredAt:   "2023-03-03",
		},
	}

	cache, err := getStarsCache()
	s.NoError(err)

	err = cache.Write(stars)
	s.NoError(err)

	stars, err = GetAllStars("rwjblue")
	s.NoError(err)
	s.Equal(2, len(stars), "Should only return stars for rwjblue")

	s.Equal(stars, []Star{
		{
			Stargazer:   "rwjblue",
			Repo:        "malleatus/tamjaweb",
			Description: "A web app",
			URL:         "https://github.com/malleatus/tamjaweb",
			StarredAt:   "2023-01-01",
		},
		{
			Stargazer:   "rwjblue",
			Repo:        "emberjs/ember.js",
			Description: "Ember.js framework",
			URL:         "https://github.com/emberjs/ember.js",
			StarredAt:   "2023-03-03",
		},
	})
}

func (s *GitHubTestSuite) TestFilterStarsByTerm() {
	stars := []Star{
		{
			Stargazer:   "user1",
			Repo:        "owner1/repo1",
			Description: "A test repository",
			URL:         "https://github.com/owner1/repo1",
			StarredAt:   "2023-01-01",
		},
		{
			Stargazer:   "user1",
			Repo:        "owner2/another-repo",
			Description: "Another test repository",
			URL:         "https://github.com/owner2/another-repo",
			StarredAt:   "2023-02-01",
		},
		{
			Stargazer:   "user2",
			Repo:        "owner3/golang-project",
			Description: "A Go programming project",
			URL:         "https://github.com/owner3/golang-project",
			StarredAt:   "2023-03-01",
		},
	}

	testCases := []struct {
		name          string
		searchTerm    string
		expectedStars []struct {
			repo        string
			description string
			url         string
			starredAt   string
			stargazer   string
		}
	}{
		{
			name:       "Filter by repo exact match",
			searchTerm: "owner1/repo1",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{
				{
					repo:        "owner1/repo1",
					description: "A test repository",
					url:         "https://github.com/owner1/repo1",
					starredAt:   "2023-01-01",
					stargazer:   "user1",
				},
			},
		},
		{
			name:       "Filter by repo partial match",
			searchTerm: "another",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{
				{
					repo:        "owner2/another-repo",
					description: "Another test repository",
					url:         "https://github.com/owner2/another-repo",
					starredAt:   "2023-02-01",
					stargazer:   "user1",
				},
			},
		},
		{
			name:       "Filter by description",
			searchTerm: "Go programming",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{
				{
					repo:        "owner3/golang-project",
					description: "A Go programming project",
					url:         "https://github.com/owner3/golang-project",
					starredAt:   "2023-03-01",
					stargazer:   "user2",
				},
			},
		},
		{
			name:       "Case insensitive search",
			searchTerm: "test",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{
				{
					repo:        "owner1/repo1",
					description: "A test repository",
					url:         "https://github.com/owner1/repo1",
					starredAt:   "2023-01-01",
					stargazer:   "user1",
				},
				{
					repo:        "owner2/another-repo",
					description: "Another test repository",
					url:         "https://github.com/owner2/another-repo",
					starredAt:   "2023-02-01",
					stargazer:   "user1",
				},
			},
		},
		{
			name:       "Filter with fzf style match",
			searchTerm: "go",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{
				{
					repo:        "owner3/golang-project",
					description: "A Go programming project",
					url:         "https://github.com/owner3/golang-project",
					starredAt:   "2023-03-01",
					stargazer:   "user2",
				},
			},
		},
		{
			name:       "No matches",
			searchTerm: "nonexistent",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{},
		},
		{
			name:       "Empty search term matches all",
			searchTerm: "",
			expectedStars: []struct {
				repo        string
				description string
				url         string
				starredAt   string
				stargazer   string
			}{
				{
					repo:        "owner1/repo1",
					description: "A test repository",
					url:         "https://github.com/owner1/repo1",
					starredAt:   "2023-01-01",
					stargazer:   "user1",
				},
				{
					repo:        "owner2/another-repo",
					description: "Another test repository",
					url:         "https://github.com/owner2/another-repo",
					starredAt:   "2023-02-01",
					stargazer:   "user1",
				},
				{
					repo:        "owner3/golang-project",
					description: "A Go programming project",
					url:         "https://github.com/owner3/golang-project",
					starredAt:   "2023-03-01",
					stargazer:   "user2",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := FilterStarsByTerm(stars, tc.searchTerm)

			s.Equal(len(tc.expectedStars), len(result), "Number of stars with matches")

			if len(tc.expectedStars) == 0 {
				s.Empty(result, "Result should be empty when no matches")
				return
			}

			for i, expectedStar := range tc.expectedStars {
				if i < len(result) {
					s.Equal(expectedStar.repo, result[i].Repo, "Star repo at index %d", i)
					s.Equal(expectedStar.description, result[i].Description, "Star description at index %d", i)
					s.Equal(expectedStar.url, result[i].URL, "Star URL at index %d", i)
					s.Equal(expectedStar.starredAt, result[i].StarredAt, "Star starredAt at index %d", i)
					s.Equal(expectedStar.stargazer, result[i].Stargazer, "Star stargazer at index %d", i)
				}
			}
		})
	}
}

func TestGitHubTestSuite(t *testing.T) {
	suite.Run(t, new(GitHubTestSuite))
}
