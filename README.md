## SpotifyCleanUp
This is a little application to manage my playlists. I have music that I already enjoy and toss them into my library and one or two playlists at most. It does cause some trouble when I have heavy metal, R&B, Rap, and Indie all mixed together. I do like all the songs in the playlist or my library, but if I want to show people a playlist of a certain mood or genre it gets distracting. If you too have the issue of not separating your playlists or library, or just want to split up your current playlists into genre/mood this is perfect.

### Configuration
Configuration for SpotifyCleanUp is managed through the `.in` file. inside of `config` directory. Executing `spotify.yaml.in` will create `spotify.yaml` which will ask for credentials such as the `Key` and `Secret` . All `spotify.yaml` files MUST be created using the `spotify.yaml.in` file.

### Config Path
SpotifyCleanUp relies on using enviornment variable `CFGPATH`. `CFGPATH` is the
location to look for configuation files. Prefixing go commands like so
```
CFGPATH=$(pwd) go ...
```
makes sure the enviornment variable is set and config files are able to be found. If `CFGPATH` is not found it will default to `.` which is unknown and undesirable.

### Running and Testing
Before compiling be sure to run tests as not testing could cause issues with Spotify playlists such as false deletion, creation, or modification. 
All test can be ran with
```
make tests
```
if not all tests pass, the application is buggy and cannot be run without definite errors in Spotify account or connection.

Once all tests are ran, and the final build is ready, a binary can be built and ran with
```
make all
```
then running
```
./spotifycleanup.out
```