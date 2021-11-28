# Mosaic

## Build

```
go generate mosaic/cmd/mosaic
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" mosaic/cmd/mosaic
```

## Usage

```
mosaic [OPTIONS] URL ...
```

Options:

- `-a string` - HTTP server address (default ":8001")
- `-t int` - Number of threads (default 10)
- `URL` - one or more link to m3u/m3u8 playlists
