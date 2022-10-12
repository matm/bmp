## Best Song Parts

`bsp` stands for *Best Song Parts*. It's a simple CLI tool working along with the [Music Player Daemon (MPD)](https://www.musicpd.org/).

### Purpose

When listening to music, there is sometimes a favourite moment in a song: a chorus, a guitar solo, an amazing variation, whatever. Something you love. A part of a song that you wish you could remember the position of so you could listen to it again at any time.

That's what `bsp` is all about: marking all the musical moments you like the most and playing them automatically. These time sliders are saved in an easy to edit, easy to share text file.

`bsp` tries to keep things as simple as possible and provide an enjoyable audio experience. It works on most systems, binaries are provided for Linux, Mac and Windows.

### Features

- Interactive shell
- From this shell
  - Mark one or many locations while listening to a song
  - Edit those locations (deletion, easy time shifting etc.)
  - Send the song playlist to MPD and start playing your favorite parts

### Usage

Run `bsp` to access the interactive shell:
```bash
$ bsp
>
```

From there, you can run a couple of commands:

**Command**|**Action**
---|---
`q`|Exit the program
`Q`|Force exit the program, even with unsaved changes
`i`|Show current song information
`[`|Bookmark start: mark the beginning of the time frame
`]`|Bookmark end: mark the end of the time frame. The time interval is added to the list of bookmarks for the current song
`d pos`|Delete bookmark entry at position `pos`
`f`|Forward seek +10s in current song
`b`|Backward seek -10s in current song
`t`|Toggle play/pause of current song
`p`|List of current bookmarked locations in the current song
`n`|Numbered list of current bookmarked locations in the current song
`w [best.txt]`|List bookmarks on standard output. This is the content that would be saved to disk. Takes an optional argument of the filename to write to. For example, `w best.txt` would write the list to `best.txt`

### Donations

If you use this tool and want to support me in its development, a donation would be greatly appreciated!

It's not about the amount at all: making a donation boosts the motivation to work on a project. Thank you very much if you can give anything.