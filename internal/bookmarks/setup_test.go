package bookmarks

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)

	code := m.Run()

	os.Exit(code)
}
