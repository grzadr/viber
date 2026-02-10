package sub

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// const (
// 	ffprobeCmd         = "ffprobe"
// 	ffprobeShowEntries =
// "format=size,duration:format_tags=ARTIST,TITLE,ALBUM,track,disc,TRACKTOTAL,DISCTOTAL:stream=sample_rate,bits_per_raw_sample"
// )

// func RunFFprobe(filePath string) ([]byte, []byte, error) {
// 	cmd := exec.Command(ffprobeCmd,
// 		"-v", "error",
// 		"-hide_banner",
// 		"-select_streams", "a:0",
// 		"-show_entries", ffprobeShowEntries,
// 		"-print_format", "json",
// 		filePath,
// 	)

// 	var stdout, stderr bytes.Buffer
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr

// 	err := cmd.Run()
// 	if err != nil {
// 		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf(
// 			"ffprobe failed: %w",
// 			err,
// 		)
// 	}

// 	return stdout.Bytes(), stderr.Bytes(), nil
// }

type CommandResult struct {
	ExitCode int
	Stderr   string // Captured separately for debugging
	Error    error  // The Go error (e.g., context canceled, binary not found)
}

// StreamCommand runs a system command, streaming stdout line-by-line.
// Stderr is buffered and returned in the final result.
func StreamCommand(
	ctx context.Context,
	command string,
	args ...string,
) (<-chan string, <-chan CommandResult, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	// 1. Pipe for Stdout (Streaming)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pipe stdout: %w", err)
	}

	// 2. Buffer for Stderr (Captured for debugging)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// 3. Start the process
	if cmdErr := cmd.Start(); cmdErr != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", cmdErr)
	}

	stdoutChan := make(chan string)
	resultChan := make(
		chan CommandResult,
		1,
	) // Buffered to prevent leaking the monitor goroutine

	// var wg sync.WaitGroup
	// wg.Add(1)

	done := make(chan struct{})

	// 4. Goroutine: Consume Stdout Stream
	go func() {
		defer close(done)
		// defer wg.Done()

		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			stdoutChan <- scanner.Text()
		}
		// Note: We don't close stdoutChan here. We close it in the monitor to
		// ensure
		// strict ordering (all output processed -> then result sent).
	}()

	// 5. Goroutine: Monitor Process
	go func() {
		// Wait for stdout stream to finish reading
		<-done
		// wg.Wait()
		close(stdoutChan)

		// Wait for process exit and pipe closure
		waitErr := cmd.Wait()

		res := CommandResult{
			ExitCode: 0,
			Stderr:   stderrBuf.String(),
			Error:    waitErr,
		}

		// Determine specific Exit Code
		if waitErr != nil {
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				res.ExitCode = exitErr.ExitCode()
			} else {
				// E.g., Process killed by signal or failed to wait
				res.ExitCode = -1
			}
		}

		resultChan <- res
		close(resultChan)
	}()

	return stdoutChan, resultChan, nil
}
