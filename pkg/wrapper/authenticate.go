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

// wrappercontext is the struct used for storing all API data, including the client which is heavily used
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

// SpotifyConf is the struct for the yaml values of credentials from spotify.YAML
type SpotifyConf struct {
	Spotify struct {
		Credentials struct {
			Key    string `yaml:"Key"`
			Secret string `yaml:"Secret"`
		} `yaml:"Credentials"`
	} `yaml:"Spotify"`
}

// playlist is the struct to be used by Playlists interface to store all playlist IDs, track IDs, track tags,
// and categories/genres to create.
type playlist struct {
	playlistIDs []*spotify.PlaylistTrackPage
	tracks     []spotify.ID
	trackTags  [][]string
	categories []string
}

// SpotifyWrapper is the interface for Logging in and completing authentication, crucial for getting a client.
type SpotifyWrapper interface {
	LoginAccount() (*spotify.Client, error)
	completeAuth(res http.ResponseWriter, req *http.Request)
}

// Playlists is the interface for aggregating, classfiying, and creating playlists
type Playlists interface {
	GetAggregatePlaylist(*spotify.Client)
}

// NewRest will initialize a new rest through the wrapper using the credentials provided in spotify.YAML
// This will only access local memory and will not directly use the api. This only stores data in the
// wrappercontext struct in an easy way for the wrapper to access.
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

// LoginAccount will take a while to work initially, but after the URL is presented the first time it can be used over and over.
// It is recommended to run program then click link rather than click link then run program. 
// LoginAccount will work in tandem with CompleteAuth as there is a call to completeAuth in LoginAccount.
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
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
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

// completeAuth is used to pass through http.HandleFunc and will use the state, Auth, and channel from the wrappercontext struct, w.
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

// NewPlaylist will create a new playlists struct with intialized type arrrays.
func NewPlaylist() Playlists {
	return &playlist{make([]*spotify.PlaylistTrackPage,0),make([]spotify.ID, 0), make([][]string, 0), make([]string, 0)}
}

// GetAggregatePlaylist gets all playlists for a user that they make visible on their profile, and will grab every trackID.
func (p *playlist) GetAggregatePlaylist(client *spotify.Client) {
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatalf("Error getting current user, %+v", err)
	}
	pl, err := client.GetPlaylistsForUser(user.ID)
	if err != nil{
		log.Fatalf("Error getting user playlists %+v",err)
	}
	for i := range(pl.Playlists){
		playlistID, err := client.GetPlaylistTracks(pl.Playlists[i].ID)
		if err != nil{
			log.Fatalf("Error gathering playlist IDs %+v",err)
		}
		p.playlistIDs = append(p.playlistIDs, playlistID)
		log.Println(pl.Playlists[i].Name)
		for page := 1; ; page++ {
			log.Printf("  Page %d has %d tracks", page, len(playlistID.Tracks))
			for pageRange := range(playlistID.Tracks){
				trackID := playlistID.Tracks[pageRange].Track.SimpleTrack.ID
				duplicate := false
				for ids := range(p.tracks){
					if p.tracks[ids] == trackID{
						duplicate = true
					}
				}
				if !duplicate{
					p.tracks = append(p.tracks, trackID)
				}
			}
			err = client.NextPage(playlistID)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}