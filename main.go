package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr = flag.String("a", ":8001", "HTTP server address")
)

func taskStep(urlList []string) {
	for _, url := range urlList {
		statusCode, body, err := fasthttp.Get(nil, url)

		if err != nil {
			fmt.Printf("failed to get playlist %s [%s]\n", url, err)
			continue
		}

		if statusCode != fasthttp.StatusOK {
			fmt.Printf("failed to get playlist %s [%d]\n", url, statusCode)
			continue
		}

		playlist, err := NewPlaylist(string(body))

		if err != nil {
			fmt.Printf("failed to parse playlist %s [%s]\n", url, err)
			continue
		}

		// TODO: make screenshots

		fmt.Printf("%+v\n", playlist)
	}

	time.Sleep(5 * time.Second)
}

func task(urlList []string) {
	for {
		taskStep(urlList)
	}
}

func handle(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q", ctx.RequestURI())
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

	urlList := flag.Args()

	/* Screenshot task */

	go task(urlList)

	/* HTTP Server */

	fasthttp.ListenAndServe(*addr, handle)
}
