package main

import (
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"sync"
	"text/template"

	"github.com/valyala/fasthttp"
)

var (
	addr    = flag.String("a", ":8001", "HTTP server address")
	threads = flag.Int("t", 10, "Number of threads")
	images  = flag.Int("i", 40, "Number of images on page")
)

//go:embed assets
var assets embed.FS

type Image struct {
	Name  string
	Data  string
	Error string
}

type App struct {
	Config struct {
		Playlists []string
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
		png, err := takeScreenshot(channel.Address)

		var image Image
		if err != nil {
			image = Image{
				Name:  channel.Name,
				Error: fmt.Sprintf("%s", err),
			}
		} else {
			image = Image{
				Name: channel.Name,
				Data: base64.StdEncoding.EncodeToString(png),
			}
		}

		limit := min(*images-1, len(app.Images))
		app.Images = append([]Image{image}, app.Images[:limit]...)
	}

	wait.Done()
}

func mainTaskStep(app *App) {
	for _, url := range app.Config.Playlists {
		playlist, err := GetPlaylist(url)

		if err != nil {
			fmt.Printf("playlist %s: %s\n", url, err)
			continue
		}

		var wait sync.WaitGroup

		step := (len(playlist.Items) + *threads - 1) / *threads

		for i := 0; i < len(playlist.Items); {
			z := min(i+step, len(playlist.Items))

			wait.Add(1)
			go screenshotTask(&wait, app, playlist.Items[i:z])

			i = z
		}

		wait.Wait()
	}
}

func mainTask(app *App) {
	mainTaskStep(app)
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

	go mainTask(app)

	/* HTTP Server */

	fasthttp.ListenAndServe(*addr, app.handle)
}
