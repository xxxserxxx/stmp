// Copyright 2023 The STMPS Authors
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/spezifisch/stmps/logger"
	"github.com/spezifisch/stmps/mpvplayer"
	"github.com/spezifisch/stmps/remote"
	"github.com/spezifisch/stmps/subsonic"
	"github.com/spf13/viper"
)

func readConfig() {
	required_properties := []string{"auth.username", "auth.password", "server.host"}

	viper.SetConfigName("stmp")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/stmp")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("Config file error: %s \n", err)
		os.Exit(1)
	}

	for _, prop := range required_properties {
		if !viper.IsSet(prop) {
			fmt.Printf("Config property %s is required\n", prop)
		}
	}
}

func main() {
	help := flag.Bool("help", false, "Print usage")
	enableMpris := flag.Bool("mpris", false, "Enable MPRIS2")
	list := flag.Bool("list", false, "list server data")
	flag.Parse()
	if *help {
		fmt.Printf("USAGE: %s <args>\n", os.Args[0])
		flag.Usage()
		os.Exit(0)
	}

	readConfig()

	logger := logger.Init()

	connection := subsonic.Init(logger)
	connection.SetClientInfo(clientName, clientVersion)
	connection.Username = viper.GetString("auth.username")
	connection.Password = viper.GetString("auth.password")
	connection.Host = viper.GetString("server.host")
	connection.PlaintextAuth = viper.GetBool("auth.plaintext")
	connection.Scrobble = viper.GetBool("server.scrobble")

	indexResponse, err := connection.GetIndexes()
	if err != nil {
		fmt.Printf("Error fetching indexes from server: %s\n", err)
		os.Exit(1)
	}
	// TODO (B) loading playlists can take a long time on e.g. gonic if there are a lot of them; can it be done in the background?
	playlistResponse, err := connection.GetPlaylists()
	if err != nil {
		fmt.Printf("Error fetching indexes from server: %s\n", err)
		os.Exit(1)
	}

	if *list {
		fmt.Printf("Index response:\n")
		fmt.Printf("  Directory: %s\n", indexResponse.Directory.Name)
		fmt.Printf("  Status: %s\n", indexResponse.Status)
		fmt.Printf("  Error: %s\n", indexResponse.Error.Message)
		fmt.Printf("  Playlist: %s\n", indexResponse.Playlist.Name)
		fmt.Printf("  Playlists: (%d)\n", len(indexResponse.Playlists.Playlists))
		for _, pl := range indexResponse.Playlists.Playlists {
			fmt.Printf("    [%d] %s\n", pl.Entries.Len(), pl.Name)
		}
		fmt.Printf("  Indexes:\n")
		for _, pl := range indexResponse.Indexes.Index {
			fmt.Printf("    %s\n", pl.Name)
		}
		fmt.Printf("Playlist response:\n")
		fmt.Printf("  Directory: %s\n", playlistResponse.Directory.Name)
		fmt.Printf("  Status: %s\n", playlistResponse.Status)
		fmt.Printf("  Error: %s\n", playlistResponse.Error.Message)
		fmt.Printf("  Playlist: %s\n", playlistResponse.Playlist.Name)
		fmt.Printf("  Playlists: (%d)\n", len(indexResponse.Playlists.Playlists))
		for _, pl := range playlistResponse.Playlists.Playlists {
			fmt.Printf("    [%d] %s\n", pl.Entries.Len(), pl.Name)
		}
		fmt.Printf("  Indexes:\n")
		for _, pl := range playlistResponse.Indexes.Index {
			fmt.Printf("    %s\n", pl.Name)
		}

		os.Exit(0)
	}

	// init mpv engine
	player, err := mpvplayer.NewPlayer(logger)
	if err != nil {
		fmt.Println("Unable to initialize mpv. Is mpv installed?")
		os.Exit(1)
	}

	// init mpris2 player control (linux only but fails gracefully on other systems)
	if *enableMpris {
		mpris, err := remote.RegisterMprisPlayer(player, logger)
		if err != nil {
			fmt.Printf("Unable to register MPRIS with DBUS: %s\n", err)
			fmt.Println("Try running without MPRIS")
			os.Exit(1)
		}
		defer mpris.Close()
	}

	// init macos mediaplayer control
	if runtime.GOOS == "darwin" {
		if err = remote.RegisterMPMediaHandler(player, logger); err != nil {
			fmt.Printf("Unable to initialize MediaPlayer bindings: %s\n", err)
			os.Exit(1)
		} else {
			logger.Print("MacOS MediaPlayer registered")
		}
	}

	ui := InitGui(&indexResponse.Indexes.Index,
		&playlistResponse.Playlists.Playlists,
		connection,
		player,
		logger)

	// run main loop
	if err := ui.Run(); err != nil {
		panic(err)
	}
}
