package github

import (
	"errors"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/suite"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

type GitHubTestSuite struct {
	suite.Suite

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
}

// TearDownTest runs after each test in the suite.
func (s *GitHubTestSuite) TearDownTest() {
	DefaultExecRunner = s.originalExecRunner
	BuildGitHubClient = s.originalBuildGitHubClient
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
	opts := []recorder.Option{
		recorder.WithMode(recorder.ModeRecordOnce),
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

func TestGitHubTestSuite(t *testing.T) {
	suite.Run(t, new(GitHubTestSuite))
}
