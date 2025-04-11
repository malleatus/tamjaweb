package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CacheCommandTestSuite struct {
	suite.Suite

	originalHomeDir string
	tempHomeDir     string
	stdout          *bytes.Buffer
	stderr          *bytes.Buffer
	originalStdout  io.Writer
	originalStderr  io.Writer
}

// SetupSuite is run once before any tests in the suite
func (s *CacheCommandTestSuite) SetupSuite() {
	s.originalHomeDir = os.Getenv("HOME")

	s.originalStdout = os.Stdout
	s.originalStderr = os.Stderr
}

// TearDownSuite is run once after all tests in the suite
func (s *CacheCommandTestSuite) TearDownSuite() {
	os.Stdout = s.originalStdout.(*os.File)
	os.Stderr = s.originalStderr.(*os.File)
}

// SetupTest is run before each test
func (s *CacheCommandTestSuite) SetupTest() {
	// Create a temporary home directory
	tempDir, err := os.MkdirTemp("", "tamjaweb-cache-test")
	s.Require().NoError(err)
	s.tempHomeDir = tempDir
	err = os.Setenv("HOME", tempDir)
	s.Require().NoError(err)

	// Create cache directory and sample file
	cacheDir := filepath.Join(tempDir, ".cache", "tamjaweb")
	err = os.MkdirAll(cacheDir, 0755)
	s.Require().NoError(err)

	// Create a test cache file
	testCacheFile := filepath.Join(cacheDir, "test.json")
	err = os.WriteFile(testCacheFile, []byte(`{"test":"data"}`), 0644)
	s.Require().NoError(err)

	s.stdout = new(bytes.Buffer)
	s.stderr = new(bytes.Buffer)
}

// TearDownTest is run after each test
func (s *CacheCommandTestSuite) TearDownTest() {
	err := os.Setenv("HOME", s.originalHomeDir)
	s.Require().NoError(err)

	defer func() {
		err := os.RemoveAll(s.tempHomeDir)
		s.NoError(err)
	}()
}

// TestCacheCommandShowPath tests that the command shows the cache path
func (s *CacheCommandTestSuite) TestCacheCommandShowPath() {
	cmd := newCacheCommand()
	cmd.SetOut(s.stdout)
	cmd.SetErr(s.stderr)

	// Redirect output to our buffer
	cmd.SetOut(s.stdout)
	cmd.SetErr(s.stderr)

	// Run the command
	err := cmd.Execute()
	s.Require().NoError(err)

	// Verify output contains cache path
	expectedPath := filepath.Join(s.tempHomeDir, ".cache", "tamjaweb")
	s.Contains(s.stdout.String(), expectedPath)
}

// TestCacheCommandClear tests the --clear flag
func (s *CacheCommandTestSuite) TestCacheCommandClear() {
	cmd := newCacheCommand()
	cmd.SetOut(s.stdout)
	cmd.SetErr(s.stderr)

	// Add the --clear flag
	cmd.SetArgs([]string{"--clear"})

	// Check file exists before running command
	cacheFile := filepath.Join(s.tempHomeDir, ".cache", "tamjaweb", "test.json")
	_, err := os.Stat(cacheFile)
	s.Require().NoError(err, "Cache file should exist before clearing")

	// Run the command
	err = cmd.Execute()
	s.Require().NoError(err)

	// Verify file was deleted
	_, err = os.Stat(cacheFile)
	s.Require().True(os.IsNotExist(err), "Cache file should be deleted after clearing")

	// Verify success message
	s.Contains(s.stdout.String(), "Cache cleared successfully")
}

// TestCacheCommandInSubTests runs the tests using Go's subtests pattern
func TestCacheCommand(t *testing.T) {
	suite.Run(t, new(CacheCommandTestSuite))
}
