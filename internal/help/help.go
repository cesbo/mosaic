package help

import (
	"fmt"
	"io"
)

type AppInfo struct {
	AppName       string
	VersionDate   string
	VersionCommit string
	ExecPath      string
}

func Usage(wr io.Writer, info *AppInfo) {
	Version(wr, info)
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
`, info.ExecPath)
}

func Version(wr io.Writer, info *AppInfo) {
	fmt.Fprintf(wr, "%s v%s (commit:%s)\n", info.AppName, info.VersionDate, info.VersionCommit)
}
