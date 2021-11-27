package playlist

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

// Parse m3u/m3u8 file
func (playlist *Playlist) parsePlaylist(source string) error {
	scanner := bufio.NewScanner(strings.NewReader(source))

	if !scanner.Scan() {
		return fmt.Errorf("empty playlist")
	}

	if !strings.HasPrefix(scanner.Text(), "#EXTM3U") {
		return fmt.Errorf("invalid playlist format")
	}

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

			playlist.Items = append(playlist.Items, *channel)
			isInfo = true

			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if isInfo {
			playlist.Items[len(playlist.Items)-1].Address = line
			isInfo = false
		}
	}

	return nil
}

// Populate playlist object from the m3u8 link
func (playlist *Playlist) GetPlaylist(url string) error {
	statusCode, body, err := fasthttp.Get(nil, url)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("request failed: %d", statusCode)
	}

	if err := playlist.parsePlaylist(string(body)); err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	return nil
}
