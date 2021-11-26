package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

func takeScreenshot(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"/usr/bin/env",
		"ffmpeg",
		"-i", url,
		"-v", "quiet",
		"-ss", "2",
		"-y",
		"-t", "1",
		"-vframes", "1",
		"-vf", "scale=160:100",
		"-f", "image2",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg execute: %w", err)
	}

	result, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdout read: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w", err)
	}

	return result, nil
}
