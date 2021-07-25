package wrapper

import (
	"log"
	"testing"
)

var spotifyRest = NewRest()
var spotifyPlaylist = NewPlaylist()

func TestGetAggregatePlaylist(t *testing.T) {
	// https://accounts.spotify.com/authorize?client_id=e1963af1e71e4fb18f341898aad96c33&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=user-read-private&state=Active
	client, err := spotifyRest.LoginAccount()
	if err != nil {
		log.Fatalf("Error logging in, %+v", err)
	}
	spotifyPlaylist.GetAggregatePlaylist(client)
}