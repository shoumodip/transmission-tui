# Transmission TUI
Simple TUI for [Transmission](https://transmissionbt.com/)

## Quick Start
Install [Xclip](https://github.com/astrand/xclip) and [Transmission](https://transmissionbt.com/)

```console
$ go build
$ ./transmission-tui
```

## Usage
| Key          | Description                                          |
| ------------ | ---------------------------------------------------- |
| <kbd>q</kbd> | Quit                                                 |
| <kbd>r</kbd> | Refresh                                              |
| <kbd>j</kbd> | Move the cursor down                                 |
| <kbd>k</kbd> | Move the cursor up                                   |
| <kbd>g</kbd> | Move the cursor to the top                           |
| <kbd>G</kbd> | Move the cursor to the bottom                        |
| <kbd>a</kbd> | Add Torrent from link in clipboard                   |
| <kbd>x</kbd> | Remove the torrent under the cursor                  |
