package watch

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yaacov/kubectl-mtv/pkg/util/tui"
)

// DefaultInterval is the default watch interval for all watch operations
const DefaultInterval = 5 * time.Second

// RenderFunc is a function that renders output and returns an error if any
type RenderFunc func() error

// captureOutput wraps a RenderFunc into a DataFetcher by capturing its stdout output.
func captureOutput(renderFunc RenderFunc) tui.DataFetcher {
	return func() (string, error) {
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			return "", fmt.Errorf("failed to create pipe: %w", err)
		}
		os.Stdout = w

		outputChan := make(chan string)
		go func() {
			var buf strings.Builder
			_, _ = io.Copy(&buf, r)
			outputChan <- buf.String()
		}()

		renderErr := renderFunc()

		w.Close()
		os.Stdout = oldStdout
		output := <-outputChan

		return output, renderErr
	}
}

// Watch uses TUI mode for watching with smooth updates and interactive features.
// It exits when user presses q or Ctrl+C.
func Watch(renderFunc RenderFunc, interval time.Duration) error {
	return tui.Run(captureOutput(renderFunc), interval)
}

// WatchWithQuery is like Watch but additionally enables the interactive query (`:`) feature.
func WatchWithQuery(renderFunc RenderFunc, interval time.Duration, queryUpdater tui.QueryUpdater, currentQuery string) error {
	opts := []tui.Option{
		tui.WithQueryUpdater(queryUpdater),
	}
	if currentQuery != "" {
		opts = append(opts, tui.WithInitialQuery(currentQuery))
	}
	return tui.RunWithOptions(captureOutput(renderFunc), interval, opts...)
}
