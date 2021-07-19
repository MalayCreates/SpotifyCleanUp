package grabber

import (
	"log"

	"github.com/zmb3/spotify"
)

type Playlists interface {
	GetAggregatePlaylist(*spotify.Client)
}

type playlist struct {
	tracks     []string
	trackTags  [][]string
	categories []string
}

func NewPlaylists() Playlists {
	return &playlist{make([]string, 0), make([][]string, 0), make([]string, 0)}
}

func (p *playlist) GetAggregatePlaylist(client *spotify.Client) {
	user, err := client.CurrentUser()
	if err != nil{
		log.Fatalf("Error getting current user, %+v",err)
	}
	pl, _ := client.GetPlaylistsForUser(user.ID)
	playlistID := string(pl.Playlists[0].ID)
	x, _ := (client.GetPlaylistTracks(spotify.ID(playlistID)))
	log.Println(len(x.Tracks))
	// fmt.Println(x.Tracks[0].Track.SimpleTrack.ID)
}
