package watch

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/yaacov/kubectl-mtv/pkg/util/tui"
)

// RenderFunc is a function that renders output and returns an error if any
type RenderFunc func() error

// Watch uses TUI mode for watching with smooth updates and interactive features
// It exits when user presses q or Ctrl+C
func Watch(renderFunc RenderFunc, interval time.Duration) error {
	// Create a data fetcher that captures output from the renderFunc
	dataFetcher := func() (string, error) {
		// Create a pipe to capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Create a channel to collect output
		outputChan := make(chan string)
		go func() {
			var buf strings.Builder
			io.Copy(&buf, r)
			outputChan <- buf.String()
		}()

		// Call renderFunc which will print to our captured stdout
		err := renderFunc()

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout
		output := <-outputChan

		return output, err
	}

	return tui.Run(dataFetcher, interval)
}
