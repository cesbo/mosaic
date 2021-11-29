# Mosaic

## Build

```
GOOS=linux \
GOARCH=amd64 \
VERSION=$(date "+%y.%m") \
COMMIT=$(git rev-parse --short=8 HEAD) \
go build -ldflags "-s -w -X main.versionDate=$VERSION -X main.versionCommit=$COMMIT" mosaic/cmd/mosaic
```

## Usage

```
mosaic help
```
