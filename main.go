package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	gc "github.com/vit1251/go-ncursesw"
)

type Client struct {
	err   error
	title string
	items []string

	anchor int
	cursor int

	height int
	window *gc.Window
}

func clientNew() (Client, error) {
	c := Client{}

	if exec.Command("pgrep", "-x", "transmission-da").Run() != nil {
		fmt.Println("Starting transmission daemon...")
		if err := exec.Command("transmission-daemon").Run(); err != nil {
			return c, err
		}
		time.Sleep(100 * time.Millisecond)
	}

	window, err := gc.Init()
	if err != nil {
		return c, err
	}

	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)
	gc.SetEscDelay(0)
	window.Keypad(true)
	window.Timeout(2000) // Refresh every 2 seconds

	gc.StartColor()
	gc.UseDefaultColors()

	gc.InitPair(COLOR_ERROR, gc.C_RED, -1)
	gc.InitPair(COLOR_TITLE, gc.C_BLUE, -1)

	c.window = window
	err = c.Update()
	return c, err
}

func (c *Client) SelectPrev() {
	if c.cursor > 0 {
		c.cursor--
	}
}

func (c *Client) SelectNext() {
	if c.cursor+1 < len(c.items) {
		c.cursor++
	}
}

func (c *Client) SelectFirst() {
	c.cursor = 0
}

func (c *Client) SelectLast() {
	c.cursor = max(0, len(c.items)-1)
}

func (c *Client) Update() error {
	output, err := exec.Command("transmission-remote", "-l").Output()
	if err != nil {
		return err
	}

	c.items = strings.Split(string(output), "\n")
	c.title = c.items[0]
	c.items = c.items[1 : len(c.items)-2]
	return nil
}

func (c *Client) Append() error {
	link, ok := c.Prompt("Link: ", "")
	if !ok {
		return nil
	}

	return exec.Command("transmission-remote", "-a", link).Run()
}

func (c *Client) Remove() error {
	if len(c.items) < 1 {
		return nil
	}

	n := strings.Split(strings.TrimLeft(c.items[c.cursor], " "), " ")[0]
	return exec.Command("transmission-remote", "-t", n, "-r").Run()
}

func (c *Client) Prompt(query string, init string) (string, bool) {
	gc.Cursor(1)
	defer gc.Cursor(0)

	input := NewLine(init)
	error := false

	for {
		c.height, _ = c.window.MaxYX()

		c.window.AttrOn(gc.A_BOLD)
		c.window.ColorOn(COLOR_TITLE)
		c.window.MovePrint(c.height-1, 0, query)
		c.window.AttrOff(gc.A_BOLD)
		c.window.ColorOff(COLOR_TITLE)
		c.window.ClearToEOL()

		if error {
			c.window.AttrOn(gc.A_BOLD)
			c.window.ColorOn(COLOR_ERROR)
		}

		c.window.Print(input.String())

		if error {
			c.window.AttrOff(gc.A_BOLD)
			c.window.ColorOff(COLOR_ERROR)
		}

		c.window.Move(c.height-1, len(query)+input.cursor)
		c.window.Refresh()

		ch := c.window.GetChar()
		switch ch {
		case 27:
			c.window.Timeout(10)
			ch := c.window.GetChar()
			c.window.Timeout(-1)

			switch ch {
			case 0:
				return "", false

			case 'f':
				input.NextWord()

			case 'b':
				input.PrevWord()

			case 'd':
				input.Delete((*Line).NextWord)

			case gc.KEY_BACKSPACE:
				input.Delete((*Line).PrevWord)
			}

		case 'c' & 0x1f:
			return "", false

		case 'f' & 0x1f:
			input.NextChar()

		case 'b' & 0x1f:
			input.PrevChar()

		case 'a' & 0x1f:
			input.Start()

		case 'e' & 0x1f:
			input.End()

		case 'd' & 0x1f:
			input.Delete((*Line).NextChar)

		case 'k' & 0x1f:
			input.Delete((*Line).End)

		case 'u' & 0x1f:
			input.Delete((*Line).Start)

		case gc.KEY_RETURN:
			return input.String(), true

		case gc.KEY_BACKSPACE:
			input.Delete((*Line).PrevChar)

		default:
			if strconv.IsPrint(rune(ch)) {
				input.Insert(byte(ch))
			}
		}
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

const (
	COLOR_ERROR = iota
	COLOR_TITLE
)

func main() {
	c, err := clientNew()
	handleError(err)
	defer gc.End()

	for {
		c.window.Erase()
		c.height, _ = c.window.MaxYX()
		rows := c.height - 2

		if c.cursor >= c.anchor+rows {
			c.anchor = c.cursor - rows + 1
		}

		if c.cursor < c.anchor {
			c.anchor = c.cursor
		}

		c.window.AttrOn(gc.A_BOLD)
		c.window.ColorOn(COLOR_TITLE)
		c.window.MovePrintln(0, 0, c.title)
		c.window.AttrOff(gc.A_BOLD)
		c.window.ColorOff(COLOR_TITLE)

		last := min(len(c.items), rows+c.anchor)
		for i := c.anchor; i < last; i++ {
			if i == c.cursor {
				c.window.AttrOn(gc.A_REVERSE)
			}

			c.window.Println(c.items[i])

			if i == c.cursor {
				c.window.AttrOff(gc.A_REVERSE)
			}
		}

		if c.err != nil {
			c.window.AttrOn(gc.A_BOLD)
			c.window.ColorOn(COLOR_ERROR)
			c.window.MovePrint(c.height-1, 0, c.err)
			c.window.AttrOff(gc.A_BOLD)
			c.window.ColorOff(COLOR_ERROR)
			c.err = nil
		}
		c.window.Refresh()

		key := c.window.GetChar()
		switch key {
		case 'q':
			return

		case 'r':
			handleError(c.Update())

		case 'j':
			c.SelectNext()

		case 'k':
			c.SelectPrev()

		case 'g':
			c.SelectFirst()

		case 'G':
			c.SelectLast()

		case 'x':
			c.err = c.Remove()
			handleError(c.Update())

		case 'a':
			if c.err = c.Append(); c.err == nil {
				handleError(c.Update())
				c.SelectLast()
			}

		case 0:
			c.Update()
		}
	}
}
