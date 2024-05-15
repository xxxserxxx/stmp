// Copyright 2023 The STMPS Authors
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/spezifisch/stmps/mpvplayer"
	"github.com/spezifisch/stmps/subsonic"
)

func (ui *Ui) handlePageInput(event *tcell.EventKey) *tcell.EventKey {
	// we don't want any of these firing if we're trying to add a new playlist
	focused := ui.app.GetFocus()
	if ui.playlistPage.IsNewPlaylistInputFocused(focused) || ui.browserPage.IsSearchFocused(focused) {
		return event
	}

	switch event.Rune() {
	case '1':
		ui.ShowPage(PageBrowser)

	case '2':
		ui.ShowPage(PageQueue)

	case '3':
		ui.ShowPage(PagePlaylists)

	case '4':
		ui.ShowPage(PageLog)

	case '?':
		ui.ShowHelp()

	case 'Q':
		ui.Quit()

	case 'r':
		// add random songs to queue
		ui.handleAddRandomSongs()

	case 'D':
		// clear queue and stop playing
		ui.player.ClearQueue()
		ui.queuePage.UpdateQueue()

	case 'p':
		// toggle playing/pause
		err := ui.player.Pause()
		if err != nil {
			ui.logger.PrintError("handlePageInput: Pause", err)
		}
		return nil

	case 'P':
		// stop playing without changes to queue
		ui.logger.Print("key stop")
		err := ui.player.Stop()
		if err != nil {
			ui.logger.PrintError("handlePageInput: Stop", err)
		}
		return nil

	case 'X':
		// debug stuff
		ui.logger.Print("test")
		//ui.player.Test()
		ui.showMessageBox("foo bar")
		return nil

	case '-':
		// volume-
		if err := ui.player.AdjustVolume(-5); err != nil {
			ui.logger.PrintError("handlePageInput: AdjustVolume-", err)
		}
		return nil

	// TODO (A) volume up with '+'; trivial, but needs to be a different patch so adding note
	case '=':
		// volume+
		if err := ui.player.AdjustVolume(5); err != nil {
			ui.logger.PrintError("handlePageInput: AdjustVolume+", err)
		}
		return nil

	case '.':
		// <<
		if err := ui.player.Seek(10); err != nil {
			ui.logger.PrintError("handlePageInput: Seek+", err)
		}
		return nil

	case ',':
		// >>
		if err := ui.player.Seek(-10); err != nil {
			ui.logger.PrintError("handlePageInput: Seek-", err)
		}
		return nil

	case '>':
		// skip to next track
		if err := ui.player.PlayNextTrack(); err != nil {
			ui.logger.PrintError("handlePageInput: Next", err)
		}
		ui.queuePage.UpdateQueue()
	}

	return event
}

func (ui *Ui) ShowPage(name string) {
	ui.pages.SwitchToPage(name)
	ui.menuWidget.SetActivePage(name)
}

func (ui *Ui) Quit() {
	ui.player.Quit()
	ui.app.Stop()
}

func (ui *Ui) handleAddRandomSongs() {
	ui.addRandomSongsToQueue()
	ui.queuePage.UpdateQueue()
}

func (ui *Ui) addRandomSongsToQueue() {
	response, err := ui.connection.GetRandomSongs()
	if err != nil {
		ui.logger.Printf("addRandomSongsToQueue %s", err.Error())
	}
	for _, e := range response.RandomSongs.Song {
		ui.addSongToQueue(&e)
	}
}

// make sure to call ui.QueuePage.UpdateQueue() after this
func (ui *Ui) addSongToQueue(entity *subsonic.SubsonicEntity) {
	uri := ui.connection.GetPlayUrl(entity)

	queueItem := &mpvplayer.QueueItem{
		Id:       entity.Id,
		Uri:      uri,
		Title:    entity.GetSongTitle(),
		Artist:   entity.Artist,
		Duration: entity.Duration,
	}
	ui.player.AddToQueue(queueItem)
}

func makeSongHandler(entity *subsonic.SubsonicEntity, ui *Ui, fallbackArtist string) func() {
	// make copy of values so this function can be used inside a loop iterating over entities
	id := entity.Id
	uri := ui.connection.GetPlayUrl(entity)
	title := entity.Title
	artist := stringOr(entity.Artist, fallbackArtist)
	duration := entity.Duration

	return func() {
		if err := ui.player.PlayUri(id, uri, title, artist, duration); err != nil {
			ui.logger.PrintError("SongHandler Play", err)
			return
		}
		ui.queuePage.UpdateQueue()
	}
}
