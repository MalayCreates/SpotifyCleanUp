package main

import (
	"fmt"
	"log"

	"github.com/MalayCreates/SpotifyCleanUp/pkg/grabber"
	"github.com/MalayCreates/SpotifyCleanUp/pkg/wrapper"
)

var spotifyRest wrapper.SpotifyWrapper
var playlist grabber.Playlists

func main(){
	fmt.Println("Starting the app")

	spotifyRest = wrapper.NewRest()
	client, err := spotifyRest.LoginAccount()
	if err != nil{
		log.Fatalf("Error logging in")
	}

	playlist = grabber.NewPlaylists()
	playlist.GetAggregatePlaylist(client)
}