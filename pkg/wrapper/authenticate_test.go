package wrapper

import (
	"log"
	"testing"

	"github.com/zmb3/spotify"
)

// Testing used a sample playlist containing a limited amoount of songs
// This testing is no good, need to redo it


var spotifyRest = NewRest()


func TestGetAggregatePlaylist(t *testing.T) {
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

	var shit []spotify.ID
	err = spotifyPlaylist.CreateCategories(shit)
	if err != nil {
		t.Errorf("Error in TestCreateCategories, %+v", err)
	}
}

// func TestCreateCategories(t *testing.T) {
// 	var shit []spotify.ID
// 	shit = append(shit, spotify.ID("0eGsygTp906u18L0Oimnem"))
// 	shit = append(shit, spotify.ID("0OgGn1ofaj55l2PcihQQGV"))
// 	shit = append(shit, spotify.ID("2MLHyLy5z5l5YRp7momlgw"))
// 	err := spotifyPlaylist.CreateCategories(shit)
// 	if err != nil {
// 		t.Errorf("Error in TestCreateCategories, %+v", err)
// 	}
// }
