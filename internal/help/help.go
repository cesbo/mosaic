package help

import (
	"fmt"
	"io"
	"os"
)

const (
	AppName   string = "Mosaic"
	UserAgent string = AppName + "/" + VersionDate
)

func Usage(wr io.Writer) {
	Version(wr)

	fmt.Fprintf(wr, `
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

func Version(wr io.Writer) {
	fmt.Fprintf(wr, "%s v%s (commit:%s)\n", AppName, VersionDate, VersionCommit)
}
