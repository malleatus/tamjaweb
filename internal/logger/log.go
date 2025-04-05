package logger

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

var enabledNamespaces = map[string]bool{}

func init() {
	debugEnv := os.Getenv("DEBUG")
	for ns := range strings.SplitSeq(debugEnv, ",") {
		ns = strings.TrimSpace(ns)
		if ns == "*" {
			enabledNamespaces["*"] = true
			break
		}
		if ns != "" {
			enabledNamespaces[ns] = true
		}
	}
}

func New(ns string) *log.Logger {
	if enabledNamespaces["*"] || enabledNamespaces[ns] {
		l := log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller: false,
			Prefix:       ns,
		})
		l.SetLevel(log.DebugLevel)
		return l
	}

	// Disabled logger: send everything to a no-op function
	l := log.NewWithOptions(nil, log.Options{
		Prefix: ns,
	})
	l.SetLevel(log.ErrorLevel)
	return l
}
