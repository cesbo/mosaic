package mosaic

import (
	"embed"
	"encoding/base64"
	"fmt"
	"mosaic/internal/config"
	"mosaic/internal/playlist"
	"mosaic/internal/screenshot"
	"sync"
	"text/template"
	"time"

	"github.com/valyala/fasthttp"
)

//go:embed assets
var assets embed.FS

type Image struct {
	Name string
	Data string
}

type App struct {
	Config config.Config

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

func (app *App) mainTaskStep() {
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

func (app *App) mainTask() {
	for {
		start := time.Now()
		app.mainTaskStep()
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

func (app *App) Start() {
	go app.mainTask()
	fasthttp.ListenAndServe(app.Config.Listen, app.handle)
}
