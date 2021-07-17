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
	Client      chan *spotify.Client
}

type SpotifyConf struct {
	Spotify struct {
		Credentials struct {
			Key    string `yaml:"Key"`
			Secret string `yaml:"Secret"`
		} `yaml:"Credentials"`
	} `yaml:"Spotify"`
}
type SpotifyWrapper interface {
	LoginAccount()
	completeAuth(res http.ResponseWriter, req *http.Request)
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
	state := "active"
	ch := make(chan *spotify.Client)

	return &wrappercontext{
		redirectURI: redirectURL,
		Key:         config.Spotify.Credentials.Key,
		Secret:      config.Spotify.Credentials.Secret,
		Version:     "2",
		State:       state,
		AuthURL:     "None",
		Auth:        spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadPrivate),
		Client:      ch,
	}
}

func (w *wrappercontext) LoginAccount() {
	// first start an HTTP server
	http.HandleFunc("/callback", w.completeAuth)
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log.Println("Got request for:", req.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := w.Auth.AuthURL(w.State)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-w.Client

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
}

func (w *wrappercontext) completeAuth(res http.ResponseWriter, req *http.Request) {
	w.Auth.SetAuthInfo(w.Key, w.Secret)
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
	w.Client <- &client
}
