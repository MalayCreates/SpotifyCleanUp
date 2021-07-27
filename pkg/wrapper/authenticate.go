package wrapper

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/zmb3/spotify"
	"gopkg.in/yaml.v3"
)

// wrappercontext is the struct used for storing all API data, including the client which is heavily used
type wrappercontext struct {
	RedirectURI string
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
	playlistIDs []spotify.ID
	tracks      []spotify.ID
	artistIDs   []spotify.ID
	trackTags   map[spotify.ID][]string
	categories  []string
	Client      *spotify.Client
}

// SpotifyWrapper is the interface for Logging in and completing authentication, crucial for getting a client.
type SpotifyWrapper interface {
	LoginAccount() (spotify.ID, error)
	completeAuth(res http.ResponseWriter, req *http.Request)
	NewPlaylist() Playlists
}

// Playlists is the interface for aggregating, classfiying, and creating playlists
type Playlists interface {
	GetAggregatePlaylist() error
	CreateCategories() error
	CreatePlaylists(userID spotify.ID) error
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
		RedirectURI: redirectURL,
		Key:         config.Spotify.Credentials.Key,
		Secret:      config.Spotify.Credentials.Secret,
		Version:     "2",
		State:       "Active",
		AuthURL:     "None",
		Auth:        spotify.NewAuthenticator(redirectURL, spotify.ScopePlaylistModifyPrivate,),
		Channel:     ch,
		Client:      nil,
		UserID:      "None",
	}
}

// LoginAccount will take a while to work initially, but after the URL is presented the first time it can be used over and over.
// It is recommended to run program then click link rather than click link then run program.
// LoginAccount will work in tandem with CompleteAuth as there is a call to completeAuth in LoginAccount.
func (w *wrappercontext) LoginAccount() (spotify.ID, error) {
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

	return spotify.ID(w.UserID), nil
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
func (w *wrappercontext) NewPlaylist() Playlists {
	return &playlist{make([]spotify.ID, 0), make([]spotify.ID, 0), make([]spotify.ID, 0), make(map[spotify.ID][]string), make([]string, 0), w.Client}
}

// GetAggregatePlaylist gets all playlists for a user that they make visible on their profile, and will grab every trackID.
func (p *playlist) GetAggregatePlaylist() error {
	log.Println(p)
	user, err := p.Client.CurrentUser()
	if err != nil {
		log.Fatalf("Error getting current user, %+v", err)
	}
	pl, err := p.Client.GetPlaylistsForUser(user.ID)
	if err != nil {
		log.Fatalf("Error getting user playlists %+v", err)
	}
	for i := range pl.Playlists {
		p.playlistIDs = append(p.playlistIDs, pl.Playlists[i].ID)
		playlistID, err := p.Client.GetPlaylistTracks(pl.Playlists[i].ID)
		if err != nil {
			log.Fatalf("Error gathering playlist IDs %+v", err)
		}
		playlistName := pl.Playlists[i].Name
		log.Printf("Collecting %s", playlistName)
		for page := 1; ; page++ {
			log.Printf("Collecting page %d of %d tracks", page, len(playlistID.Tracks))
			for pageRange := range playlistID.Tracks {
				trackID := playlistID.Tracks[pageRange].Track.SimpleTrack.ID
				duplicate := false
				for ids := range p.tracks {
					if p.tracks[ids] == trackID {
						duplicate = true
					}
				}
				if !duplicate {
					p.trackTags[trackID] = []string{}
					p.tracks = append(p.tracks, trackID)
				}
			}
			err = p.Client.NextPage(playlistID)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("Added %s", playlistName)
	}
	log.Println("Completed collecting all tracks")
	return nil
}

func (p *playlist) CreateCategories() error {

	pages := float64(len(p.tracks) / 50)
	lastPage := len(p.tracks) % 50
	for i := 1; i <= int(math.Ceil(pages)); i++ {
		if i != int(math.Ceil(pages)) || lastPage == 0 {
			trackWindow, err := p.Client.GetTracks((p.tracks[((i - 1) * 50):(i * 50)])...)
			if err != nil {
				log.Fatalf("Error getting track details %+v", err)
			}
			for j := range trackWindow {
				p.artistIDs = append(p.artistIDs, (trackWindow[j].SimpleTrack.Artists[0].ID))
			}
			artistWindow, err := p.Client.GetArtists((p.artistIDs[((i - 1) * 50):(i * 50)])...)
			// artistWindow, err := p.Client.GetArtists((p.artistIDs[0:3 ])...)
			if err != nil {
				log.Fatalf("Error getting artist details")
			}
			for j := range artistWindow {
				genres := artistWindow[j].Genres
				for g := range genres {
					p.categories = append(p.categories, genres[g])
					p.trackTags[p.tracks[((i-1)*50)+j]] = append(p.trackTags[p.tracks[((i-1)*50)+j]], genres[g])
				}
			}
		} else {
			trackWindow, err := p.Client.GetTracks((p.tracks[(len(p.tracks) - 1 - lastPage):(len(p.tracks))])...)
			if err != nil {
				log.Fatalf("Error getting track details %+v", err)
			}
			for j := range trackWindow {
				p.artistIDs = append(p.artistIDs, (trackWindow[j].SimpleTrack.Artists[0].ID))
			}
			artistWindow, err := p.Client.GetArtists((p.artistIDs[(len(p.tracks) - 1 - lastPage):(len(p.tracks))])...)
			// artistWindow, err := p.Client.GetArtists((p.artistIDs[0:3 ])...)
			if err != nil {
				log.Fatalf("Error getting artist details")
			}
			for j := range artistWindow {
				genres := artistWindow[j].Genres
				for g := range genres {
					p.categories = append(p.categories, genres[g])
					p.trackTags[p.tracks[((i-1)*50)+j]] = append(p.trackTags[p.tracks[((i-1)*50)+j]], genres[g])
				}
			}
		}
	}
	log.Printf("%d Tracks have been tagged", len(p.trackTags))
	return nil
}

func (p *playlist) CreatePlaylists(userID spotify.ID) error {
	keys := make(map[string]bool)
    list := []string{}
    for _, entry := range p.categories {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }
	p.categories = list

	for _, ele := range p.categories{
		playlistName := strings.Title(strings.ToLower(ele))
		playlistName = strings.ReplaceAll(playlistName," ","")
		log.Printf("_%s\n",playlistName)
		f, err := p.Client.CreatePlaylistForUser(string(userID),playlistName,playlistName,false)
		log.Println(f)
		if err != nil{
			log.Fatalf("error creating playlist %+v", err)
		}
	}
	
	return nil
}
