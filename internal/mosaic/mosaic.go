package mosaic

import (
	"embed"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sync"
	"text/template"
	"time"

	"mosaic/internal/config"
	"mosaic/internal/help"
	"mosaic/internal/playlist"
	"mosaic/internal/screenshot"
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
		delay := 30 * time.Second
		if delay > elapsed {
			time.Sleep(delay - elapsed)
		}
	}
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(assets, "assets/index.html")
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Server", help.UserAgent)

	app.mu.Lock()
	tmpl.Execute(w, app)
	app.mu.Unlock()
}

func (app *App) Start() {
	go app.mainTask()

	router := http.NewServeMux()
	router.HandleFunc("/", app.indexHandler)

	server := &http.Server{
		Handler:      router,
		Addr:         app.Config.Listen,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %s", err)
		os.Exit(1)
	}
}
