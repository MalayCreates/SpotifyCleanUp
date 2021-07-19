package wrapper

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify"
	"gopkg.in/yaml.v3"
)

type wrappercontext struct {
	redirectURI string
	Key         string
	Secret      string
	Version     string
	State       string
	AuthURL     string
	Auth        spotify.Authenticator
	Channel     chan *spotify.Client
	Client      *spotify.Client
	UserID      string
}

type SpotifyConf struct {
	Spotify struct {
		Credentials struct {
			Key    string `yaml:"Key"`
			Secret string `yaml:"Secret"`
		} `yaml:"Credentials"`
	} `yaml:"Spotify"`
}

type playlist struct {
	tracks     []string
	trackTags  [][]string
	categories []string
}

type SpotifyWrapper interface {
	LoginAccount() (*spotify.Client, error)
	completeAuth(res http.ResponseWriter, req *http.Request)
}

type Playlists interface {
	GetAggregatePlaylist(*spotify.Client)
}

func NewRest() SpotifyWrapper {
	configPath := os.Getenv("CFGPATH")
	if configPath == "" {
		configPath = "."
	}

	var yamlFile []byte
	yamlFile, err := ioutil.ReadFile(configPath + "/configs/spotify.yaml")
	if err != nil {
		log.Fatalf("Unable to load config file %+v", err)
		log.Fatal("Run spotify.yaml.in")
		os.Exit(1)
	}

	config := SpotifyConf{}
	err = yaml.Unmarshal([]byte(yamlFile), &config)
	if err != nil {
		log.Fatalf("Error in unmarshalling %+v", err)
	}
	redirectURL := "http://localhost:8080/callback"
	ch := make(chan *spotify.Client)

	return &wrappercontext{
		redirectURI: redirectURL,
		Key:         config.Spotify.Credentials.Key,
		Secret:      config.Spotify.Credentials.Secret,
		Version:     "2",
		State:       "Active",
		AuthURL:     "None",
		Auth:        spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadPrivate),
		Channel:     ch,
		Client:      nil,
		UserID:      "None",
	}
}

func (w *wrappercontext) LoginAccount() (*spotify.Client, error) {
	// first start an HTTP server
	w.Auth.SetAuthInfo(w.Key, w.Secret)
	http.HandleFunc("/callback", w.completeAuth)
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log.Println("Got request for:", req.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := w.Auth.AuthURL(w.State)
	w.AuthURL = url
	// fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	// wait for auth to complete
	client := <-w.Channel

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatalf("error in user client %+v", err)
	}
	w.UserID = user.ID
	fmt.Println("You are logged in as:", w.UserID)
	w.Client = client

	return w.Client, nil
}

func (w *wrappercontext) completeAuth(res http.ResponseWriter, req *http.Request) {
	tok, err := w.Auth.Token(w.State, req)
	if err != nil {
		http.Error(res, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := req.FormValue("state"); st != w.State {
		http.NotFound(res, req)
		log.Fatalf("State mismatch: %s != %s\n", st, w.State)
	}
	// use the token to get an authenticated client
	client := w.Auth.NewClient(tok)
	fmt.Fprintf(res, "Login Completed!")
	w.Channel <- &client
}

func NewPlaylists() Playlists {
	return &playlist{make([]string, 0), make([][]string, 0), make([]string, 0)}
}

func (p *playlist) GetAggregatePlaylist(client *spotify.Client) {
	user, err := client.CurrentUser()
	// var playlistIDs []spotify.ID
	if err != nil {
		log.Fatalf("Error getting current user, %+v", err)
	}
	pl, err := client.GetPlaylistsForUser(user.ID)
	if err != nil{
		log.Fatalf("Error getting user playlists %+v",err)
	}
	for i := range(pl.Playlists){
		// playlistIDs = append(playlistIDs, pl.Playlists[i].ID)
		tracks, _ := client.GetPlaylistTracks(pl.Playlists[i].ID)
		for j := range(tracks.Tracks){
			log.Println(tracks.Tracks[j].Track.SimpleTrack.Name)
		}
	}
}