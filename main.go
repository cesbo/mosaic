package main

import (
	"context"
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"text/template"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr    = flag.String("a", ":8001", "HTTP server address")
	threads = flag.Int("t", 10, "Number of threads")
)

//go:embed assets
var assets embed.FS

type Image struct {
	Name string
	Data string
}

type App struct {
	Config struct {
		Playlists []string
	}

	Images []Image
}

func getScreenshot(url string) ([]byte, error) {
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

func makeScreenshot(app *App, channel *Channel) error {
	png, err := getScreenshot(channel.Address)
	if err != nil {
		png, err = assets.ReadFile("assets/error.png")
		if err != nil {
			panic(err)
		}
	}

	image := Image{
		Name: channel.Name,
		Data: base64.StdEncoding.EncodeToString(png),
	}

	app.Images = append([]Image{image}, app.Images...)

	return nil
}

func makeScreenshots(wait *sync.WaitGroup, app *App, channels []Channel) {
	for _, item := range channels {
		if err := makeScreenshot(app, &item); err != nil {
			fmt.Printf("error on channel %s: %s\n", item.Address, err)
		}
	}

	wait.Done()
}

func taskStep(app *App) {
	for _, url := range app.Config.Playlists {
		playlist, err := getPlaylist(url)

		if err != nil {
			fmt.Printf("playlist %s: %s\n", url, err)
			continue
		}

		var wait sync.WaitGroup

		step := (len(playlist.Items) + *threads - 1) / *threads

		for i := 0; i < len(playlist.Items); {
			z := i + step
			if z > len(playlist.Items) {
				z = len(playlist.Items)
			}

			wait.Add(1)
			go makeScreenshots(&wait, app, playlist.Items[i:z])
			i = z
		}

		wait.Wait()
	}

	fmt.Println("DONE")
}

func task(app *App) {
	// for {
	taskStep(app)
	// }
}

func (app *App) handle(ctx *fasthttp.RequestCtx) {
	tmpl, err := template.ParseFS(assets, "assets/index.html")
	if err != nil {
		ctx.Error(fmt.Sprintf("%s", err), 500)
		return
	}

	type IndexData struct {
		Images []Image
	}

	data := IndexData{
		Images: app.Images,
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.Execute(ctx.Response.BodyWriter(), data)
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

	app := new(App)
	app.Config.Playlists = flag.Args()

	/* Screenshot task */

	go task(app)

	/* HTTP Server */

	fasthttp.ListenAndServe(*addr, app.handle)
}
