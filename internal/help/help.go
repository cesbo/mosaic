package help

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const (
	AppName string = "Mosaic"
)

var (
	Version   string
	UserAgent string
)

func init() {
	versionDate := "-"
	versionCommit := "-"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				versionCommit = setting.Value[:8]
			case "vcs.time":
				if build, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					versionDate = build.Format("06.01")
				}
			}
		}
	}

	Version = fmt.Sprintf("%s (commit:%s)", versionDate, versionCommit)
	UserAgent = fmt.Sprintf("%s/%s", AppName, versionDate)
}

func PrintHelp() {
	fmt.Printf(`
Usage:
    %s command|config

command:

    help        print help
    version     print version

config:         path to configuration file

config format:
{
    "listen": ":8004",
    "threads": 10,
    "images": 4,
    "refresh": 10,
    "playlists": [
        "http://example.com/playlist.m3u8"
    ]
}

config options:
    listen      - HTTP server address. default: ":8004"
                  example: "127.0.0.1:8004", ":8004"
    threads     - Number of threads. default: 10
    images      - Number of images per threads. default: 4
    refresh     - Refresh interval in seconds. default: 10
    playlists   - List of links to playlists (m3u or m3u8). required
`, os.Args[0])
}

func PrintVersion() {
	fmt.Printf("%s v%s\n", AppName, Version)
}
