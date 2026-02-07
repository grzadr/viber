package sub

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	ffprobeCmd         = "ffprobe"
	ffprobeShowEntries = "format=size,duration:format_tags=ARTIST,TITLE,ALBUM,track,disc,TRACKTOTAL,DISCTOTAL:stream=sample_rate,bits_per_raw_sample"
)

func RunFFprobe(filePath string) ([]byte, []byte, error) {
	cmd := exec.Command(ffprobeCmd,
		"-v", "error",
		"-hide_banner",
		"-select_streams", "a:0",
		"-show_entries", ffprobeShowEntries,
		"-print_format", "json",
		filePath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("ffprobe failed: %w", err)
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}