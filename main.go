package main

import term "github.com/nsf/termbox-go"

func main() {
	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	var st state
	for st.input() {
		st.render()
	}
}

type state struct {
	path   [][]rune
	buffer []rune
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
			case term.KeyBackspace, term.KeyBackspace2, term.KeyDelete:
				if len(st.buffer) > 0 {
					st.buffer = st.buffer[0 : len(st.buffer)-1]
				} else {
					if len(st.path) > 0 {
						st.buffer = st.path[len(st.path)-1]
						st.path = st.path[0 : len(st.path)-1]
					}
				}
			}
		case '/':
			st.path = append(st.path, st.buffer)
			st.buffer = []rune{}
		default:
			st.buffer = append(st.buffer, event.Ch)
		}
	case term.EventResize:
	case term.EventMouse:
	case term.EventError:
	case term.EventInterrupt:
	case term.EventRaw:
	case term.EventNone:
	}
	return true
}

func (st state) render() {
	term.Clear(term.ColorDefault, term.ColorDefault)
	var i int
	for _, component := range st.path {
		for _, ch := range component {
			term.SetCell(i, 0, ch, term.ColorDefault, term.ColorDefault)
			i++
		}
		term.SetCell(i, 0, '/', term.ColorDefault, term.ColorDefault)
		i++
	}
	for _, ch := range st.buffer {
		term.SetCell(i, 0, ch, term.ColorDefault, term.ColorDefault)
		i++
	}
	term.SetCursor(i, 0)
	term.Flush()
}
