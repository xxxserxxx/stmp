package main

const helpPlayback = `
p     play/pause
P     stop
>     next song
-/=   volume down/volume up
,/.   seek -10/+10 seconds
r     add 50 random songs to queue
`

const helpPageBrowser = `
artist tab
  R     refresh the list
  /     Search artists
  a     Add all artist songs to queue
  n     Continue search forward
  N     Continue search backwards
song tab
  ENTER play song (clears current queue)
  a     add album or song to queue
  A     add song to playlist
  y     toggle star on song/album
  R     refresh the list
ESC   Close search
`

const helpPageQueue = `
d/DEL remove currently selected song from the queue
D     remove all songs from queue
y     toggle star on song
`

const helpPagePlaylists = `
n     new playlist
d     delete playlist
a     add playlist or song to queue
`
