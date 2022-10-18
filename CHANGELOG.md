v 0.10
  - New D command to delete current bookmark list. #20
  - dist: generate checksum files. #36

v 0.9
  - New 'h' command to show some help. #19
  - Add Makefile targets to build binaries amd64 platforms. #24
  - New '-v' flag to show program version. #9
  - Add -host and -port flags to connect to MPD host. #14
  - Gracefully exit if can't connect to MPD on startup. #27
  - Add a short tutorial and asciinema content. #26
  - New c command to edit a single time range. #13
  - Reconnect to MPD does not work as expected. #5
  - Buggy playing if many ranges in song. #21
  - Play only the best parts after run command. #16
  - New 'r' command that sends the playlist to MPD. #12
  - New '-f' flag to load a song bookmarks file. #10
  - New 'Q' command to exit the program even with unsaved changes. #4
  - Do not quit and issue a warning if changes not saved. #3
  - Persist bookmark list to disk. #1
