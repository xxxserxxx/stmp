package main

import (
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wildeyedskies/stmp/subsonic"
)

func (ui *Ui) handlePageInput(event *tcell.EventKey) *tcell.EventKey {
	// we don't want any of these firing if we're trying to add a new playlist
	focused := ui.app.GetFocus()
	if focused == ui.newPlaylistInput || focused == ui.searchField {
		return event
	}

	switch event.Rune() {
	case '1':
		ui.pages.SwitchToPage("browser")
		ui.currentPage.SetText("Browser")
	case '2':
		ui.pages.SwitchToPage("queue")
		ui.currentPage.SetText("Queue")
	case '3':
		ui.pages.SwitchToPage("playlists")
		ui.currentPage.SetText("Playlists")
	case '4':
		ui.pages.SwitchToPage("log")
		ui.currentPage.SetText("Log")
	case 'Q':
		ui.player.EventChannel <- nil
		ui.player.Instance.TerminateDestroy()
		ui.app.Stop()
	case 's':
		ui.handleAddRandomSongs()
	case 'D':
		ui.player.Queue = make([]QueueItem, 0)
		err := ui.player.Stop()
		if err != nil {
			ui.connection.Logger.Printf("handlePageInput: Stop -- %s", err.Error())
		}
		updateQueueList(ui.player, ui.queueList, ui.starIdList)
	case 'p':
		status, err := ui.player.Pause()
		if err != nil {
			ui.connection.Logger.Printf("handlePageInput: Pause -- %s", err.Error())
			ui.startStopStatus.SetText("[::b]stmp: [red]error")
			return nil
		}
		if status == PlayerStopped {
			ui.startStopStatus.SetText("[::b]stmp: [red]stopped")
		} else if status == PlayerPlaying {
			ui.startStopStatus.SetText("[::b]stmp: [green]playing " + ui.player.Queue[0].Title)
		} else if status == PlayerPaused {
			ui.startStopStatus.SetText("[::b]stmp: [yellow]paused")
		}
		return nil
	case '-':
		if err := ui.player.AdjustVolume(-5); err != nil {
			ui.connection.Logger.Printf("handlePageInput: AdjustVolume %d -- %s", -5, err.Error())
		}
		return nil

	case '=':
		if err := ui.player.AdjustVolume(5); err != nil {
			ui.connection.Logger.Printf("handlePageInput: AdjustVolume %d -- %s", 5, err.Error())
		}
		return nil

	case '.':
		if err := ui.player.Seek(10); err != nil {
			ui.connection.Logger.Printf("handlePageInput: Seek %d -- %s", 10, err.Error())
		}
		return nil
	case ',':
		if err := ui.player.Seek(-10); err != nil {
			ui.connection.Logger.Printf("handlePageInput: Seek %d -- %s", -10, err.Error())
		}
		return nil
	}

	return event
}

func (ui *Ui) handleEntitySelected(directoryId string) {
	response, err := ui.connection.GetMusicDirectory(directoryId)
	sort.Sort(response.Directory.Entities)
	if err != nil {
		ui.connection.Logger.Printf("handleEntitySelected: GetMusicDirectory %s -- %s", directoryId, err.Error())
	}

	ui.currentDirectory = &response.Directory
	ui.entityList.Clear()
	if response.Directory.Parent != "" {
		ui.entityList.AddItem(tview.Escape("[..]"), "", 0,
			ui.makeEntityHandler(response.Directory.Parent))
	}

	for _, entity := range response.Directory.Entities {
		var title string
		var id = entity.Id
		var handler func()
		if entity.IsDirectory {
			title = tview.Escape("[" + entity.Title + "]")
			handler = ui.makeEntityHandler(entity.Id)
		} else {
			title = entityListTextFormat(entity, ui.starIdList)
			handler = makeSongHandler(id, ui.connection.GetPlayUrl(&entity),
				title, stringOr(entity.Artist, response.Directory.Name),
				entity.Duration, ui.player, ui.queueList, ui.starIdList)
		}

		ui.entityList.AddItem(title, "", 0, handler)
	}
}

func (ui *Ui) handlePlaylistSelected(playlist subsonic.SubsonicPlaylist) {
	ui.selectedPlaylist.Clear()

	for _, entity := range playlist.Entries {
		var title string
		var handler func()

		var id = entity.Id

		title = entity.GetSongTitle()
		handler = makeSongHandler(id, ui.connection.GetPlayUrl(&entity), title, entity.Artist, entity.Duration, ui.player, ui.queueList, ui.starIdList)

		ui.selectedPlaylist.AddItem(title, "", 0, handler)
	}
}

func (ui *Ui) handleDeleteFromQueue() {
	currentIndex := ui.queueList.GetCurrentItem()
	queue := ui.player.Queue

	if currentIndex == -1 || len(ui.player.Queue) < currentIndex {
		return
	}

	// if the deleted item was the first one, and the player is loaded
	// remove the track. Removing the track auto starts the next one
	if currentIndex == 0 {
		if isSongLoaded, err := ui.player.IsSongLoaded(); err != nil {
			ui.connection.Logger.Printf("handleDeleteFromQueue: IsSongLoaded -- %s", err.Error())
			return
		} else if isSongLoaded {
			ui.player.Stop()
		}
		return
	}

	// remove the item from the queue
	if len(ui.player.Queue) > 1 {
		ui.player.Queue = append(queue[:currentIndex], queue[currentIndex+1:]...)
	} else {
		ui.player.Queue = make([]QueueItem, 0)
	}

	updateQueueList(ui.player, ui.queueList, ui.starIdList)
}

func (ui *Ui) handleAddRandomSongs() {
	ui.addRandomSongsToQueue()
	updateQueueList(ui.player, ui.queueList, ui.starIdList)
}

func (ui *Ui) handleToggleStar() {
	currentIndex := ui.queueList.GetCurrentItem()
	queue := ui.player.Queue

	if currentIndex == -1 || len(ui.player.Queue) < currentIndex {
		return
	}

	var entity = queue[currentIndex]

	// If the song is already in the star list, remove it
	_, remove := ui.starIdList[entity.Id]

	// resp, _ := ui.connection.ToggleStar(entity.Id, remove)
	ui.connection.ToggleStar(entity.Id, ui.starIdList)

	if remove {
		delete(ui.starIdList, entity.Id)
	} else {
		ui.starIdList[entity.Id] = struct{}{}
	}

	var text = queueListTextFormat(ui.player.Queue[currentIndex], ui.starIdList)
	updateQueueListItem(ui.queueList, currentIndex, text)
	// Update the entity list to reflect any changes
	ui.connection.Logger.Printf("entity test %v", ui.currentDirectory)
	if ui.currentDirectory != nil {
		ui.handleEntitySelected(ui.currentDirectory.Id)
	}
}

func (ui *Ui) handleAddEntityToQueue() {
	currentIndex := ui.entityList.GetCurrentItem()
	if currentIndex+1 < ui.entityList.GetItemCount() {
		ui.entityList.SetCurrentItem(currentIndex + 1)
	}

	// if we have a parent directory subtract 1 to account for the [..]
	// which would be index 0 in that case with index 1 being the first entity
	if ui.currentDirectory.Parent != "" {
		currentIndex--
	}

	if currentIndex == -1 || len(ui.currentDirectory.Entities) <= currentIndex {
		return
	}

	entity := ui.currentDirectory.Entities[currentIndex]

	if entity.IsDirectory {
		ui.addDirectoryToQueue(&entity)
	} else {
		ui.addSongToQueue(&entity)
	}

	updateQueueList(ui.player, ui.queueList, ui.starIdList)
}

func (ui *Ui) handleToggleEntityStar() {
	currentIndex := ui.entityList.GetCurrentItem()

	var entity = ui.currentDirectory.Entities[currentIndex-1]

	// If the song is already in the star list, remove it
	_, remove := ui.starIdList[entity.Id]

	ui.connection.ToggleStar(entity.Id, ui.starIdList)

	if remove {
		delete(ui.starIdList, entity.Id)
	} else {
		ui.starIdList[entity.Id] = struct{}{}
	}

	var text = entityListTextFormat(entity, ui.starIdList)
	updateEntityListItem(ui.entityList, currentIndex, text)
	updateQueueList(ui.player, ui.queueList, ui.starIdList)
}

func entityListTextFormat(queueItem subsonic.SubsonicEntity, starredItems map[string]struct{}) string {
	var star = ""
	_, hasStar := starredItems[queueItem.Id]
	if hasStar {
		star = " [red]♥"
	}
	return queueItem.Title + star
}

// Just update the text of a specific row
func updateEntityListItem(entityList *tview.List, id int, text string) {
	entityList.SetItemText(id, text, "")
}

func (ui *Ui) handleAddPlaylistSongToQueue() {
	playlistIndex := ui.playlistList.GetCurrentItem()
	entityIndex := ui.selectedPlaylist.GetCurrentItem()
	if entityIndex+1 < ui.selectedPlaylist.GetItemCount() {
		ui.selectedPlaylist.SetCurrentItem(entityIndex + 1)
	}

	// TODO add some bounds checking here
	if playlistIndex == -1 || entityIndex == -1 {
		return
	}

	entity := ui.playlists[playlistIndex].Entries[entityIndex]
	ui.addSongToQueue(&entity)

	updateQueueList(ui.player, ui.queueList, ui.starIdList)
}

func (ui *Ui) handleAddPlaylistToQueue() {
	currentIndex := ui.playlistList.GetCurrentItem()
	if currentIndex+1 < ui.playlistList.GetItemCount() {
		ui.playlistList.SetCurrentItem(currentIndex + 1)
	}

	playlist := ui.playlists[currentIndex]

	for _, entity := range playlist.Entries {
		ui.addSongToQueue(&entity)
	}

	updateQueueList(ui.player, ui.queueList, ui.starIdList)
}

func (ui *Ui) handleAddSongToPlaylist(playlist *subsonic.SubsonicPlaylist) {
	currentIndex := ui.entityList.GetCurrentItem()

	// if we have a parent directory subtract 1 to account for the [..]
	// which would be index 0 in that case with index 1 being the first entity
	if ui.currentDirectory.Parent != "" {
		currentIndex--
	}

	if currentIndex == -1 || len(ui.currentDirectory.Entities) < currentIndex {
		return
	}

	entity := ui.currentDirectory.Entities[currentIndex]

	if !entity.IsDirectory {
		ui.connection.AddSongToPlaylist(string(playlist.Id), entity.Id)
	}
	// update the playlists
	response, err := ui.connection.GetPlaylists()
	if err != nil {
		ui.connection.Logger.Printf("handleAddSongToPlaylist: GetPlaylists -- %s", err.Error())
	}
	ui.playlists = response.Playlists.Playlists

	ui.playlistList.Clear()
	ui.addToPlaylistList.Clear()

	for _, playlist := range ui.playlists {
		ui.playlistList.AddItem(playlist.Name, "", 0, nil)
		ui.addToPlaylistList.AddItem(playlist.Name, "", 0, nil)
	}

	if currentIndex+1 < ui.entityList.GetItemCount() {
		ui.entityList.SetCurrentItem(currentIndex + 1)
	}
}

func (ui *Ui) addRandomSongsToQueue() {
	response, err := ui.connection.GetRandomSongs()
	if err != nil {
		ui.connection.Logger.Printf("addRandomSongsToQueue %s", err.Error())
	}
	for _, e := range response.RandomSongs.Song {
		ui.addSongToQueue(&e)
	}
}

func (ui *Ui) addStarredToList() {
	response, err := ui.connection.GetStarred()
	if err != nil {
		ui.connection.Logger.Printf("addStarredToList %s", err.Error())
	}
	for _, e := range response.Starred.Song {
		// We're storing empty struct as values as we only want the indexes
		// It's faster having direct index access instead of looping through array values
		ui.starIdList[e.Id] = struct{}{}
	}
}

func (ui *Ui) addDirectoryToQueue(entity *subsonic.SubsonicEntity) {
	response, err := ui.connection.GetMusicDirectory(entity.Id)
	if err != nil {
		ui.connection.Logger.Printf("addDirectoryToQueue: GetMusicDirectory %s -- %s", entity.Id, err.Error())
		return
	}

	sort.Sort(response.Directory.Entities)
	for _, e := range response.Directory.Entities {
		if e.IsDirectory {
			ui.addDirectoryToQueue(&e)
		} else {
			ui.addSongToQueue(&e)
		}
	}
}

func (ui *Ui) search() {
	name, _ := ui.pages.GetFrontPage()
	if name != "browser" {
		return
	}
	ui.searchField.SetText("")
	ui.app.SetFocus(ui.searchField)
}

func (ui *Ui) searchNext() {
	str := ui.searchField.GetText()
	idxs := ui.artistList.FindItems(str, "", false, true)
	if len(idxs) == 0 {
		return
	}
	curIdx := ui.artistList.GetCurrentItem()
	for _, nidx := range idxs {
		if nidx > curIdx {
			ui.artistList.SetCurrentItem(nidx)
			return
		}
	}
	ui.artistList.SetCurrentItem(idxs[0])
}

func (ui *Ui) searchPrev() {
	str := ui.searchField.GetText()
	idxs := ui.artistList.FindItems(str, "", false, true)
	if len(idxs) == 0 {
		return
	}
	curIdx := ui.artistList.GetCurrentItem()
	for nidx := len(idxs) - 1; nidx >= 0; nidx-- {
		if idxs[nidx] < curIdx {
			ui.artistList.SetCurrentItem(idxs[nidx])
			return
		}
	}
	ui.artistList.SetCurrentItem(idxs[len(idxs)-1])
}

func (ui *Ui) addSongToQueue(entity *subsonic.SubsonicEntity) {
	uri := ui.connection.GetPlayUrl(entity)

	var artist string
	if ui.currentDirectory == nil {
		artist = entity.Artist
	} else {
		artist = stringOr(entity.Artist, ui.currentDirectory.Name)
	}

	var id = entity.Id

	queueItem := QueueItem{
		id,
		uri,
		entity.GetSongTitle(),
		artist,
		entity.Duration,
	}
	ui.player.Queue = append(ui.player.Queue, queueItem)
}

func (ui *Ui) newPlaylist(name string) {
	response, err := ui.connection.CreatePlaylist(name)
	if err != nil {
		ui.connection.Logger.Printf("newPlaylist: CreatePlaylist %s -- %s", name, err.Error())
		return
	}

	ui.playlists = append(ui.playlists, response.Playlist)

	ui.playlistList.AddItem(response.Playlist.Name, "", 0, nil)
	ui.addToPlaylistList.AddItem(response.Playlist.Name, "", 0, nil)
}

func (ui *Ui) deletePlaylist(index int) {
	if index == -1 || len(ui.playlists) < index {
		return
	}

	playlist := ui.playlists[index]

	if index == 0 {
		ui.playlistList.SetCurrentItem(1)
	}

	// Removes item with specified index
	ui.playlists = append(ui.playlists[:index], ui.playlists[index+1:]...)

	ui.playlistList.RemoveItem(index)
	ui.addToPlaylistList.RemoveItem(index)
	ui.connection.DeletePlaylist(string(playlist.Id))
}

func makeSongHandler(id string, uri string, title string, artist string, duration int, player *Player, queueList *tview.List, starIdList map[string]struct{}) func() {
	return func() {
		player.Play(id, uri, title, artist, duration)
		updateQueueList(player, queueList, starIdList)
	}
}

func (ui *Ui) makeEntityHandler(directoryId string) func() {
	return func() {
		ui.handleEntitySelected(directoryId)
	}
}