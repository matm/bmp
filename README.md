## Best Music Parts

`bmp` stands for *Best Music Parts*. It's a simple CLI tool working along with the [Music Player Daemon (MPD)](https://www.musicpd.org/).

### Purpose

When listening to music, there is sometimes a favourite moment in a song: a chorus, a guitar solo, an amazing variation, whatever. Something you love. A part of a song that you wish you could remember the position of so you could listen to it again at any time.

That's what `bmp` is all about: marking all the musical moments you like the most and playing them automatically. These time sliders are saved in an easy to edit, easy to share text file.

`bmp` tries to keep things as simple as possible and provide an enjoyable audio experience. It works on most systems, binaries are provided for Linux, Mac and Windows.

### Features

- Interactive shell
- From this shell
  - Mark one or many locations while listening to a song
  - Edit those locations
  - Send the song playlist to MPD and start playing your favorite parts


### Installation

Binaries for most amd64 platforms are built for every release. Please just [grab a binary version of the latest release](https://github.com/matm/bmp/releases).

However, if you have a working Go installation and want to build it, just run
```bash
$ go install github.com/matm/bmp/cmd/bmp@latest
```

### Usage

```bash
$ bmp -h
Usage of bmp:
  -f string
        bookmarks list file to load
  -host string
        MPD host address
  -port int
        MPD host TCP port (default 6600)
```

To connect to a MPD server, `bmp` reads the `$MPD_HOST` env variable by default. You can also use the `-host` flag to provide a MPD address, i.e. `bmp -host 192.169.1.10`. The default port `6600` will be used.

Run `bmp` to access the interactive shell:
```bash
$ bmp
>
```

or load an existing bookmark file you saved:
```bash
$ bmp -f myhits
Loaded 2 songs, 3 bookmarks
>
```

### Tutorial

Let's take a simple example. I just loaded a playlist of Metallica's [Black Album](https://www.youtube.com/watch?v=DtJzRErAJ3Q&list=PLokAorcvoBv9LAxeK6xwqn3rSEEMhGfGr)) that is ready to play.

From there, I want to

1. Listen to track 5, *Wherever I My Roam*
2. Mark the range 00:48 to 00:56 because I like this part
3. List to track 8, *Nothing Else Matters*
4. Mark the range 01:00 to 01:23 then 03:03 to 03:24
5. List the current bookmarked locations for a quick preview before saving the list
6. Save the list
7. Exit and run `bmp` with `-f` this time to load the saved bookmarks and start playing

Here we go:

[![asciicast](https://asciinema.org/a/3Jn1kVJ7MXORbhjqPRBaHajHt.svg)](https://asciinema.org/a/3Jn1kVJ7MXORbhjqPRBaHajHt)

### All Commands

Once in the interactive shell, you can run a couple of commands:

**Command**|**Action**
---|---
`h`|Show some help
`q`|Exit the program
`Q`|Force exit the program, even with unsaved changes
`i`|Show current song information
`[`|Bookmark start: mark the beginning of the time frame
`]`|Bookmark end: mark the end of the time frame. The time interval is added to the list of bookmarks for the current song
`d pos`|Delete bookmark entry at position `pos`
`c pos MM:SS-MM:SS`|Change bookmark entry at position `pos` and set new start and end time boundaries
`r`|Start the autoplay of the best parts
`s`|Stop the autoplay of the best parts
`f`|Forward seek +10s in current song
`b`|Backward seek -10s in current song
`t`|Toggle play/pause of current song
`p`|List of current bookmarked locations in the current song
`n`|Numbered list of current bookmarked locations in the current song
`w [best.txt]`|List bookmarks on standard output. This is the content that would be saved to disk. Takes an optional argument of the filename to write to. For example, `w best.txt` would write the list to `best.txt`

### Donations

If you use this tool and want to support me in its development, a donation would be greatly appreciated!

It's not about the amount at all: making a donation boosts the motivation to work on a project. Thank you very much if you can give anything.

Monero address: `88uoutKJS2w3FfkKyJFsNwKPHzaHfTAo6LyTmHSAoQHgCkCeR8FUG4hZ8oD4fnt8iP7i1Ty72V6CLMHi1yUzLCZKHU1pB7c`

![My monero address](res/qr-donate.png)
