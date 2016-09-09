package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	term "github.com/nsf/termbox-go"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	err := term.Init()
	check(err)
	defer term.Close()

	pwd, err := os.Getwd()
	check(err)
	var st state
	st.setPath(pwd)
	st.render()
	for st.input() {
		st.render()
	}

	f, err := os.Create("/tmp/nav-path")
	check(err)
	defer f.Close()
	f.WriteString(st.getPath(false))
}

type state struct {
	path   []component
	buffer []rune
}

type component struct {
	entries   []entry
	matches   []int
	selection int
	width     int
	err       error
}

type entry struct {
	name      []rune
	isMatched bool
}

func (c *component) next() {
	c.selection++
	if c.selection >= len(c.matches) {
		c.selection = 0
	}
}

func (c *component) prev() {
	c.selection--
	if c.selection < 0 {
		c.selection = len(c.matches) - 1
	}
}

func (c component) isValid() bool {
	return c.selection >= 0 && c.selection < len(c.matches)
}

func (c component) getSelected() *entry {
	if c.selection < 0 {
		return nil
	}
	return &c.entries[c.matches[c.selection]]
}

func (c *component) list(path string) {
	c.entries = nil
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		c.err = err
		c.width = len([]rune(err.Error()))
		return
	}
	for _, info := range infos {
		e := entry{
			name: []rune(info.Name()),
		}
		c.entries = append(c.entries, e)
	}
	c.selection = -1
	if len(c.matches) > 0 {
		c.selection = 0
	}
	// determine display width
	WIDTH_MIN := 8
	WIDTH_MAX := 25
	c.width = WIDTH_MIN
	for i := range c.entries {
		w := len(c.entries[i].name)
		if w <= WIDTH_MAX && w > c.width {
			c.width = w
		}
	}
}

func (c *component) filter(pattern []rune) {
	oldSelection := -1
	if c.selection < 0 {
	}
	if len(c.matches) > 0 {
		if c.selection < 0 {
			c.selection = 0
		}
		oldSelection = c.matches[c.selection]
	}
	c.matches = nil
	newSelection := -1
	for i := range c.entries {
		entry := &c.entries[i]
		entry.isMatched = strings.HasPrefix(string(entry.name), string(pattern))
		if entry.isMatched {
			if newSelection == -1 && oldSelection <= i {
				newSelection = len(c.matches)
			}
			c.matches = append(c.matches, i)
		}
	}
	if len(c.matches) > 0 && newSelection == -1 {
		newSelection = len(c.matches) - 1
	}
	c.selection = newSelection
}

func (c *component) commonPrefix() []rune {
	var cp []rune
	if len(c.matches) > 0 {
		cpLen := len(c.entries[c.matches[0]].name)
		for _, idx := range c.matches {
			name := c.entries[idx].name
			if len(name) > cpLen {
				cpLen = len(name)
			}
		}
		for i := 0; i < cpLen; i++ {
			ch := c.entries[c.matches[0]].name[i]
			for _, idx := range c.matches {
				name := c.entries[idx].name
				if ch != name[i] {
					return cp
				}
			}
			cp = append(cp, ch)
		}
	}
	return cp
}

func (st state) getPath(appendBuffer bool) string {
	if len(st.path) == 0 {
		return "/"
	}
	var buffer bytes.Buffer
	for _, component := range st.path[0 : len(st.path)-1] {
		// TODO: turn absolute path into one relative to the current dir
		// needed for better behavior w.r.t. access rights
		buffer.WriteRune('/')
		buffer.WriteString(string(component.getSelected().name))
	}
	if appendBuffer {
		buffer.WriteRune('/')
		buffer.WriteString(string(st.buffer))
	}
	return buffer.String()
}

func (st *state) setPath(path string) {
	absPath, err := filepath.Abs(path)
	check(err)
	parts := strings.Split(absPath, "/")
	cursor := "/"
	st.path = nil
	for i, part := range parts {
		if i > 0 {
			for j, entry := range st.path[i-1].entries {
				if string(entry.name) == part {
					st.path[i-1].selection = j
					break
				}
			}
		}
		cursor = filepath.Join(cursor, part)
		var c component
		c.list(cursor)
		c.filter(nil)
		st.path = append(st.path, c)
	}
}

func (st state) getCurrent() *component {
	return &st.path[len(st.path)-1]
}

func (st *state) push() {
	if st.getCurrent().isValid() {
		st.buffer = nil
		st.getCurrent().filter(nil)
		st.path = append(st.path, component{})
		st.getCurrent().list(st.getPath(false))
		st.getCurrent().filter(nil)
	}
}

func (st *state) pop() {
	if len(st.path) > 1 {
		st.path = st.path[0 : len(st.path)-1]
		st.buffer = nil
		st.getCurrent().filter(nil)
	}
}

func (st *state) insertChar(ch rune) {
	st.buffer = append(st.buffer, ch)
	st.getCurrent().filter(st.buffer)
}

func (st *state) deleteChar() {
	st.buffer = st.buffer[0 : len(st.buffer)-1]
	st.getCurrent().filter(st.buffer)
}

func (st *state) input() bool {
	event := term.PollEvent()
	switch event.Type {
	case term.EventKey:
		switch event.Ch {
		case 0:
			switch event.Key {
			case term.KeyEsc:
				return false
			case term.KeyArrowLeft:
				st.pop()
			case term.KeyArrowDown:
				if st.getCurrent().isValid() {
					st.getCurrent().next()
				}
			case term.KeyArrowUp:
				if st.getCurrent().isValid() {
					st.getCurrent().prev()
				}
			case term.KeyBackspace, term.KeyBackspace2, term.KeyDelete:
				if len(st.buffer) > 0 {
					st.deleteChar()
				} else {
					st.pop()
				}
			case term.KeyArrowRight:
				st.push()
			case term.KeyTab:
				if len(st.getCurrent().matches) == 1 {
					st.push()
				} else {
					st.buffer = st.getCurrent().commonPrefix()
					st.getCurrent().filter(st.buffer)
				}
			}
		default:
			st.insertChar(event.Ch)
		}
	case term.EventResize:
	case term.EventMouse:
	case term.EventError:
	}
	return true
}

func (st state) render() {
	term.Clear(term.ColorDefault, term.ColorDefault)
	// columns
	_, height := term.Size()
	baseX := 0
	for i, comp := range st.path {
		if comp.err != nil {
			fg, bg := term.ColorWhite, term.ColorRed
			msg := []rune(comp.err.Error())
			for y := 0; y < height-1; y++ {
				for x := 0; x < comp.width; x++ {
					ch := msg[x]
					term.SetCell(baseX+x, 1+y, ch, fg, bg)
				}
			}
		} else {
			for y := 0; y < height-1; y++ {
				var line []rune
				fg, bg := term.ColorBlack, term.ColorBlue
				if y < len(comp.entries) {
					line = comp.entries[y].name
					if !comp.entries[y].isMatched {
						fg = term.ColorMagenta
					}
				}
				if i&1 == 0 {
					bg = term.ColorGreen
				}
				if comp.selection >= 0 && y == comp.matches[comp.selection] {
					bg = term.ColorWhite
				}
				for x := 0; x < comp.width; x++ {
					ch := ' '
					if x < len(line) {
						ch = line[x]
					}
					term.SetCell(baseX+x, 1+y, ch, fg, bg)
				}
			}
		}
		baseX += comp.width
	}
	// path
	path := st.getPath(true)
	var x int
	for _, ch := range path {
		term.SetCell(x, 0, ch, term.ColorDefault, term.ColorDefault)
		x++
	}
	term.SetCursor(x, 0)
	term.Flush()
}
