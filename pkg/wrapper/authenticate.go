package wrapper

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify"
	"gopkg.in/yaml.v3"
)
type wrappercontext struct{
	redirectURI string
	Key string
	Secret string
	Version string
	BaseURL string
	Client *http.Client
}

type SpotifyConf struct{
	Spotify struct {
		Credentials struct {
			Key        string `yaml:"Key"`
			Secret     string `yaml:"Secret"`
		} `yaml:"Credentials"`
	} `yaml:"Spotify"`
}
 type SpotifyWrapper interface{
	PlaceHolder()
}

func NewRest() wrappercontext{
	configPath := os.Getenv("CFGPATH")
	if configPath == "" {
		configPath = "."
	}

	var yamlFile []byte
	yamlFile, err := ioutil.ReadFile(configPath + "/configs/spotify.yaml")
	if err != nil{
		log.Fatalf("Unable to load config file %+v", err)
		log.Fatal("Run spotify.yaml.in")
		os.Exit(1)
	}

	config := SpotifyConf{}
	err = yaml.Unmarshal([]byte(yamlFile),&config)
	if err != nil{
		log.Fatalf("Error in unmarshalling %+v", err)
	}
	redirectURI := "playlistparse://returnafterlogin"
	auth := spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
	auth.SetAuthInfo(config.Spotify.Credentials.Key, config.Spotify.Credentials.Secret)
	
}