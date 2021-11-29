# Mosaic

## Build

```
go generate mosaic/cmd/mosaic
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" mosaic/cmd/mosaic
```

## Usage

```
mosaic help
```
