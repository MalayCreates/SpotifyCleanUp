package main

import (
	"log"

	"github.com/MalayCreates/SpotifyCleanUp/pkg/wrapper"
)

// var spotifyRest wrapper.SpotifyWrapper
// var playlist wrapper.Playlists

func main() {

	var spotifyRest = wrapper.NewRest()

	_, err := spotifyRest.LoginAccount()
	if err != nil {
		log.Fatalf("Error logging in, %+v", err)
	}
	var spotifyPlaylist = spotifyRest.NewPlaylist()
	// https://accounts.spotify.com/authorize?client_id=e1963af1e71e4fb18f341898aad96c33&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=user-read-private&state=Active
	err = spotifyPlaylist.GetAggregatePlaylist()
	if err != nil {
		log.Fatalf("Error in GetAggregatePlaylist, %+v", err)
	}

	err = spotifyPlaylist.CreateCategories()
	if err != nil {
		log.Fatalf("Error in TestCreateCategories, %+v", err)
	}
}
