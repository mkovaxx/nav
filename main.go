package main

import term "github.com/nsf/termbox-go"

func main() {
	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	event_loop()
}

func event_loop() {
	for exit := false; !exit; {
		event := term.PollEvent()
		switch event.Type {
		case term.EventKey:
			if event.Key == term.KeyEsc {
				exit = true
			}
		case term.EventResize:
		case term.EventMouse:
		case term.EventError:
		case term.EventInterrupt:
		case term.EventRaw:
		case term.EventNone:
		}
	}
}
