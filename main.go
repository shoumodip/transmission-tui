package main

import (
	"fmt"
	"github.com/shoumodip/screen-go"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	err    error
	items  []string
	cursor int
}

func clientNew() (Client, error) {
	client := Client{cursor: 1}

	if exec.Command("pgrep", "-x", "transmission-da").Run() != nil {
		if err := exec.Command("transmission-daemon").Run(); err != nil {
			return client, err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return client, nil
}

func (client *Client) SelectPrev() {
	if client.cursor > 1 {
		client.cursor--
	}
}

func (client *Client) SelectNext() {
	if client.cursor+1 < len(client.items) {
		client.cursor++
	}
}

func (client *Client) SelectFirst() {
	client.cursor = 1
}

func (client *Client) SelectLast() {
	client.cursor = len(client.items) - 1
}

func (client *Client) Update() error {
	output, err := exec.Command("transmission-remote", "-l").Output()
	if err != nil {
		return err
	}

	client.items = strings.Split(string(output), "\n")
	client.items = client.items[:len(client.items)-2]
	return nil
}

func (client *Client) Append() error {
	link, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
	if err != nil {
		return err
	}

	return exec.Command("transmission-remote", "-a", string(link)).Run()
}

func (client *Client) Remove() error {
	if len(client.items) < 2 {
		return nil
	}

	n := strings.Split(strings.TrimLeft(client.items[client.cursor], " "), " ")[0]
	return exec.Command("transmission-remote", "-t", n, "-r").Run()
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func sendMessage(scr screen.Screen, message string) {
	scr.MoveCursor(0, scr.Height-1)
	fmt.Fprint(scr, message)
	scr.Flush()
}

func main() {
	client, err := clientNew()
	handleError(err)
	handleError(client.Update())

	scr, err := screen.New()
	handleError(err)
	scr.HideCursor()

	defer scr.Reset()
	for {
		scr.Clear()
		for i, item := range client.items {
			if i == client.cursor {
				scr.Apply(screen.STYLE_REVERSE)
			}

			fmt.Fprint(scr, item, "\r\n")

			if i == client.cursor {
				scr.Apply(screen.STYLE_NONE)
			}
		}

		if client.err != nil {
			scr.MoveCursor(0, scr.Height-1)
			scr.Apply(screen.COLOR_RED, screen.STYLE_BOLD)
			fmt.Fprint(scr, client.err)
			scr.Apply(screen.STYLE_NONE)
			client.err = nil
		}
		scr.Flush()

		key, err := scr.Input()
		handleError(err)

		switch key {
		case 'q':
			return

		case 'r':
			sendMessage(scr, "Refreshing...")
			handleError(client.Update())

		case 'j':
			client.SelectNext()

		case 'k':
			client.SelectPrev()

		case 'g':
			client.SelectFirst()

		case 'G':
			client.SelectLast()

		case 'x':
			client.err = client.Remove()
			handleError(client.Update())

		case 'a':
			sendMessage(scr, "Adding Torrent...")
			if client.err = client.Append(); client.err == nil {
				handleError(client.Update())
				client.SelectLast()
			}
		}
	}
}
