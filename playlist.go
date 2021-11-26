package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

type Channel struct {
	Name    string
	Address string
}

type Playlist struct {
	Items []Channel
}

func parseInfo(line string) []string {
	isName := false
	isEscaped := false
	isQuoted := false

	fieldCheck := func(c rune) bool {
		if isName {
			return false
		}

		if isEscaped {
			isEscaped = false
			return false
		}

		if c == '\\' {
			isEscaped = true
			return false
		}

		if c == '"' {
			isQuoted = !isQuoted
		}

		if isQuoted {
			return false
		}

		if c == ',' {
			isName = true
			return true
		}

		return c == ' ' || c == '='
	}

	return strings.FieldsFunc(line, fieldCheck)
}

func parsePlaylist(source string) (*Playlist, error) {
	scanner := bufio.NewScanner(strings.NewReader(source))

	if !scanner.Scan() {
		return nil, fmt.Errorf("empty playlist")
	}

	if !strings.HasPrefix(scanner.Text(), "#EXTM3U") {
		return nil, fmt.Errorf("invalid playlist format")
	}

	p := new(Playlist)

	isInfo := false

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#EXTINF") {
			parts := parseInfo(line)

			channel := new(Channel)
			channel.Name = parts[len(parts)-1]

			p.Items = append(p.Items, *channel)
			isInfo = true

			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if isInfo {
			p.Items[len(p.Items)-1].Address = line
			isInfo = false
		}
	}

	return p, nil
}

func GetPlaylist(url string) (*Playlist, error) {
	statusCode, body, err := fasthttp.Get(nil, url)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("request failed: %d", statusCode)
	}

	playlist, err := parsePlaylist(string(body))

	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	return playlist, nil
}
