# STMPS (Subsonic Terminal Music Player S)

A terminal client for *sonic music servers. Inspired by ncmpcpp and musickube.

## This fork contains all my PRs

I'm now working with [@spezifisch's](https://github.com/spezifisch/stmps) fork.
I've submitted a number of patches; this fork contains all of those, and I plan
to keep this in sync with any of @spezifisch's changes.

## Features

* browse by folder
* queue songs and albums
* create and play playlists
* favorites
* volume control
* server-side scrobbling (e.g. on Navidrome, gonic)
* [MPRIS2](https://mpris2.readthedocs.io/en/latest/) control

## Screenshots

These are using [Navidrome's demo server](https://demo.navidrome.org/) ([config file](./stmp-navidromedemo.toml)).

Queue:

![Queue View](./docs/screenshots/queue.png)

Browser:

![Browser View](./docs/screenshots/browser.png)

## Dependencies

[mpv](https://mpv.io):

* Linux (Debian/Ubuntu): `apt install libmpv libmpv-dev`
* MacOS (Homebrew): `brew install mpv` (not the cask)

Go build dependencies

* Go 1.19+
* [tview](https://github.com/rivo/tview)
* [go-mpv](https://github.com/spezifisch/go-mpv)

### OSX path setup

On OSX if you installed mpv with brew you may need to set the following paths
(these differ between Intel and ARM Macs):

```shell
export C_INCLUDE_PATH="$(brew --prefix)/include:$C_INCLUDE_PATH"
export LIBRARY_PATH="$(brew --prefix)/lib:$LIBRARY_PATH"
```

## Compiling

stmp should compile normally with `go build`. Cgo is needed for linking with libmpv.

## Configuration

stmp looks for a config file called `stmp.toml` in either `$HOME/.config/stmp`
or the directory in which the executable is placed.

### Example configuration

```toml
[auth]
username = 'admin'
password = 'password'
plaintext = true  # Use 'legacy' unsalted password auth. (default: false)

[server]
host = 'https://your-subsonic-host.tld'
scrobble = true   # Use Subsonic scrobbling for last.fm/ListenBrainz (default: false)
```

## Usage

* Q - quit
* 1 - folder view
* 2 - queue view
* 3 - playlist view
* 4 - log (errors, etc) view
* Escape/Return - close modal if open

### Playback

These are accessible in every view.

* p - play/pause
* P - stop
* &gt; - next song
* -/= volume down/volume up
* ,/. seek -10/+10 seconds
* r - add 50 random songs to the queue

### Browser

* Enter - play song (clears current queue)
* a - add album or song to queue
* y - toggle star on song/album
* A - add song to playlist
* R - refresh the list (if in artist directory, only refreshes that artist)
* / - Search artists
* n - Continue search forward
* N - Continue search backwards

### Queue

* d/Delete - remove currently selected song from the queue
* D - remove all songs from queue
* y - toggle star on song

### Playlist

* n - new playlist
* d - delete playlist
* a - add playlist or song to queue

## Credits

* This is a fork of [STMP](https://github.com/wildeyedskies/stmp), see
[AUTHORS](./AUTHORS). I decided to rename my fork as its codebase has diverged
quite a bit.
