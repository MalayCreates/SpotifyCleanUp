# Current working directory
CWD = $(shell pwd)

# Go and environmental paths
export GOPATH=$(CWD)/.go
export CFGPATH=$(CWD)/

deps:
	mkdir -p $(CWD)/.go
	go get -d ./...

# # Run go tests
# tests: deps
# 	go test -v ./...

# Build all executables
all: deps
	go build -o spotifyplaylistcreator.out $(CWD)/cmd/SpotifyPlaylistCreator