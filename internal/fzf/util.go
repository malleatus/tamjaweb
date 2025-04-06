package fzf

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	fzflib "github.com/junegunn/fzf/src"
)

// FilterOptions defines configuration options for fzf filtering
type FilterOptions struct {
	Delimiter   string
	MatchFields string
}

// DefaultFilterOptions returns the default filtering options
func DefaultFilterOptions() FilterOptions {
	return FilterOptions{
		Delimiter:   "\t",
		MatchFields: "2..", // Only match against text after index + tab
	}
}

// FilterStrings runs fzf's filter functionality on a list of strings
// Returns the indices of matched strings
func FilterStrings(inputs []string, term string) ([]int, error) {
	log.Debug("FilterStrings called", "term", term, "input_count", len(inputs))

	if term == "" {
		// If term is empty, return all indices
		indices := make([]int, len(inputs))
		for i := range inputs {
			indices[i] = i
		}
		return indices, nil
	}

	inputChan := make(chan string)
	outputChan := make(chan string)

	go func() {
		defer close(inputChan)
		for _, input := range inputs {
			inputChan <- input
		}
	}()

	options, err := fzflib.ParseOptions(
		false, // don't load defaults
		[]string{
			"--filter", term,
			"--delimiter", "\t",
			"--with-nth", "2..", // Only match against text after index + tab
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build fzf options: %w", err)
	}

	options.Input = inputChan
	options.Output = outputChan

	var wg sync.WaitGroup
	wg.Add(1)

	var matchedIndices []int

	go func() {
		defer wg.Done()
		for match := range outputChan {
			parts := strings.SplitN(match, "\t", 2)
			if len(parts) >= 1 {
				if idx, err := strconv.Atoi(parts[0]); err == nil && idx >= 0 && idx < len(inputs) {
					matchedIndices = append(matchedIndices, idx)
				}
			}
		}
	}()

	_, err = fzflib.Run(options)
	close(outputChan)

	if err != nil {
		return nil, fmt.Errorf("failed to run fzf filter: %w", err)
	}

	wg.Wait()

	return matchedIndices, nil
}
