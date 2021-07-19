package main

import (
	"fmt"
	"log"

	"github.com/MalayCreates/SpotifyCleanUp/pkg/wrapper"
)

var spotifyRest wrapper.SpotifyWrapper
var playlist wrapper.Playlists

func main(){
	fmt.Println("Starting the app")

	spotifyRest = wrapper.NewRest()
	client, err := spotifyRest.LoginAccount()
	if err != nil{
		log.Fatalf("Error logging in")
	}

	playlist = wrapper.NewPlaylists()
	playlist.GetAggregatePlaylist(client)
}