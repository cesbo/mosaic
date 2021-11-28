package playlist_test

import (
	"mosaic/internal/playlist"
	"testing"
)

func TestParser(t *testing.T) {
	t1 := `#EXTM3U
#EXTINF:-1 group-title="-",Channel #1
http://example.com/channel-1/index.m3u8
#EXTINF:-1 group-title="-",Channel #2
http://example.com/channel-2/index.m3u8
`
	p := new(playlist.Playlist)
	if err := p.ParsePlaylist(t1); err != nil {
		t.Fatalf("ParsePlaylist(t1) = %v", err)
	}

	if len(p.Items) != 2 {
		t.Fatalf("Playlist size != 2")
	}

	c1 := p.Items[0]

	c1n := "Channel #1"
	if c1.Name != c1n {
		t.Fatalf("1: Channel name is %q should be %q", c1.Name, c1n)
	}

	c1a := "http://example.com/channel-1/index.m3u8"
	if c1.Address != c1a {
		t.Fatalf("1: Channel address is %q should be %q", c1.Address, c1a)
	}

	c2 := p.Items[0]

	c2n := "Channel #1"
	if c2.Name != c2n {
		t.Fatalf("1: Channel name is %q should be %q", c2.Name, c2n)
	}

	c2a := "http://example.com/channel-1/index.m3u8"
	if c2.Address != c2a {
		t.Fatalf("1: Channel address is %q should be %q", c2.Address, c2a)
	}
}
