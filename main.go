package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr = flag.String("a", ":8001", "HTTP server address")
)

func getScreenshot(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: use /
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

func getPlaylist(url string) (*Playlist, error) {
	statusCode, body, err := fasthttp.Get(nil, url)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("request failed: %d", statusCode)
	}

	playlist, err := NewPlaylist(string(body))

	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	return playlist, nil
}

func makeScreenshot(idx int, url string) error {
	png, err := getScreenshot(url)
	if err != nil {
		return fmt.Errorf("make screenshot: %w", err)
	}

	fileName := fmt.Sprintf("/Users/and/Downloads/%d.png", idx)
	if err := os.WriteFile(fileName, png, 0644); err != nil {
		return fmt.Errorf("save png: %w", err)
	}

	return nil
}

func makeScreenshots(wait *sync.WaitGroup, playlist *Playlist, skip int, limit int) {
	if limit > len(playlist.Items) {
		limit = len(playlist.Items)
	}

	for i := skip; i < limit; i++ {
		addr := playlist.Items[i].Address
		if err := makeScreenshot(i, addr); err != nil {
			fmt.Printf("error on channel %s: %s\n", addr, err)
		}
	}

	wait.Done()
}

func taskStep(urlList []string) {
	for _, url := range urlList {
		playlist, err := getPlaylist(url)

		if err != nil {
			fmt.Printf("playlist %s: %s\n", url, err)
			continue
		}

		var wait sync.WaitGroup

		threads := 5
		step := (len(playlist.Items) + threads - 1) / threads

		for i := 0; i < len(playlist.Items); i += step {
			wait.Add(0)
			go makeScreenshots(&wait, playlist, i, i+step)
		}

		wait.Wait()
	}

	fmt.Println("DONE")
}

func task(urlList []string) {
	// for {
	taskStep(urlList)
	// }
}

func handle(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q", ctx.RequestURI())
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] URL ...\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	/* Init app */
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	urlList := flag.Args()

	/* Screenshot task */

	go task(urlList)

	/* HTTP Server */

	fasthttp.ListenAndServe(*addr, handle)
}
