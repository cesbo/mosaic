package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/valyala/fasthttp"

	"mosaic/internal/help"
	"mosaic/internal/playlist"
	"mosaic/internal/screenshot"
)

//go:embed assets
var assets embed.FS

//go:generate go run mosaic/cmd/version

var appInfo = help.AppInfo{
	AppName:       "Mosaic",
	VersionDate:   VersionDate,
	VersionCommit: VersionCommit,
	ExecPath:      os.Args[0],
}

type Image struct {
	Name string
	Data string
}

type Config struct {
	Listen    string   `json:"listen"`
	Threads   int      `json:"threads"`
	Images    int      `json:"images"`
	Refresh   int      `json:"refresh"`
	Playlists []string `json:"playlists"`
}

type App struct {
	Config

	mu     sync.Mutex
	Images []Image
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func screenshotTask(ch chan<- Image, wg *sync.WaitGroup, channels []playlist.Channel) {
	defer wg.Done()

	for _, channel := range channels {
		var image Image
		png, err := screenshot.TakeScreenshot(channel.Address)
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
		ch <- image
	}
}

func mainTaskStep(app *App) {
	playlist := new(playlist.Playlist)

	// Get all playlists
	for _, url := range app.Config.Playlists {
		if err := playlist.GetPlaylist(url); err != nil {
			fmt.Printf("playlist %s: %s\n", url, err)
			continue
		}
	}

	// Launch screenshot tasks
	for i := 0; i < len(playlist.Items); {
		var wg sync.WaitGroup
		ch := make(chan Image)

		for y := 0; y < app.Config.Threads && i < len(playlist.Items); y++ {
			z := min(i+app.Config.Images, len(playlist.Items))

			wg.Add(1)
			go screenshotTask(ch, &wg, playlist.Items[i:z])

			i = z
		}

		var images []Image
		go func() {
			for image := range ch {
				images = append(images, image)
			}
		}()

		wg.Wait()
		close(ch)

		app.mu.Lock()
		app.Images = images
		app.mu.Unlock()

		time.Sleep(time.Duration(app.Config.Refresh) * time.Second)
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

	app.mu.Lock()
	tmpl.Execute(ctx.Response.BodyWriter(), app)
	app.mu.Unlock()
}

func main() {
	if len(os.Args) == 1 || os.Args[1] == "help" {
		help.Usage(os.Stdout, &appInfo)
		os.Exit(0)
	}

	if os.Args[1] == "version" {
		help.Version(os.Stdout, &appInfo)
		os.Exit(0)
	}

	app := new(App)
	app.Config.Listen = ":8004"
	app.Config.Threads = 10
	app.Config.Images = 4
	app.Config.Refresh = 10

	fd, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %s", err)
		os.Exit(1)
	}

	defer fd.Close()

	decoder := json.NewDecoder(fd)
	if err := decoder.Decode(&app.Config); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config format: %s", err)
		os.Exit(1)
	}

	fd.Close()

	/* Screenshot task */

	go mainTask(app)

	/* HTTP Server */

	fasthttp.ListenAndServe(app.Config.Listen, app.handle)
}
