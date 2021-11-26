package main

import (
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	version = flag.Bool("v", false, "Show version info")
	addr    = flag.String("a", ":8001", "HTTP server address")
	threads = flag.Int("t", 10, "Number of threads")
	images  = flag.Int("i", 4, "Number of images per thread")
	refresh = flag.Int("r", 10, "Refresh interval in seconds")
)

//go:embed assets
var assets embed.FS

//go:generate go run cmd/version/main.go

type Image struct {
	Name string
	Data string
}

type App struct {
	Config struct {
		Refresh     int
		ImagesLimit int
		Playlists   []string
	}

	Images []Image
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func screenshotTask(wait *sync.WaitGroup, app *App, channels []Channel) {
	for _, channel := range channels {

		var image Image
		png, err := TakeScreenshot(channel.Address)
		if err != nil {
			image = Image{
				Name: channel.Name,
			}
		} else {
			image = Image{
				Name: channel.Name,
				Data: base64.StdEncoding.EncodeToString(png),
			}
		}

		limit := min(app.Config.ImagesLimit-1, len(app.Images))
		app.Images = append([]Image{image}, app.Images[:limit]...)
	}

	wait.Done()
}

func mainTaskStep(app *App) {
	playlist := new(Playlist)

	// Get all playlists
	for _, url := range app.Config.Playlists {
		if err := playlist.GetPlaylist(url); err != nil {
			fmt.Printf("playlist %s: %s\n", url, err)
			continue
		}
	}

	// Launch screenshot tasks
	for i := 0; i < len(playlist.Items); {
		var wait sync.WaitGroup

		for y := 0; y < (*threads) && i < len(playlist.Items); y++ {
			z := min(i+(*images), len(playlist.Items))

			wait.Add(1)
			go screenshotTask(&wait, app, playlist.Items[i:z])

			i = z
		}

		wait.Wait()
		time.Sleep(time.Duration(*refresh) * time.Second)
	}
}

func mainTask(app *App) {
	for {
		start := time.Now()
		mainTaskStep(app)
		elapsed := time.Since(start)

		// Limit requests rate
		if 30 > elapsed.Seconds() {
			time.Sleep((30 - elapsed) * time.Second)
		}
	}
}

func (app *App) handle(ctx *fasthttp.RequestCtx) {
	tmpl, err := template.ParseFS(assets, "assets/index.html")
	if err != nil {
		ctx.Error(fmt.Sprintf("%s", err), 500)
		return
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.Execute(ctx.Response.BodyWriter(), app)
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] URL ...\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	/* Init app */
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Printf("Mosaic %s commit:%s\n", BuildDate, BuildCommit)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	app := new(App)
	app.Config.Refresh = *refresh
	app.Config.ImagesLimit = (*threads) * (*images)
	app.Config.Playlists = flag.Args()

	/* Screenshot task */

	go mainTask(app)

	/* HTTP Server */

	fasthttp.ListenAndServe(*addr, app.handle)
}
