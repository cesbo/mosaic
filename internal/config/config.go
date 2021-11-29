package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Listen    string   `json:"listen"`
	Threads   int      `json:"threads"`
	Images    int      `json:"images"`
	Refresh   int      `json:"refresh"`
	Playlists []string `json:"playlists"`
}

func (config *Config) Load(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open: %s", err)
	}

	defer fd.Close()

	decoder := json.NewDecoder(fd)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	return nil
}
