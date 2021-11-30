package playlist

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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
func (playlist *Playlist) ParsePlaylist(source string) error {
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
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	response, err := client.Get(url)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %d", response.StatusCode)
	}

	buffer := new(strings.Builder)
	if _, err := io.Copy(buffer, response.Body); err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	if err := playlist.ParsePlaylist(buffer.String()); err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	return nil
}
