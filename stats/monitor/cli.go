/*
	NOTE: The implementation in this file is loosely based on
	https://github.com/verdverm/httopd/
*/

package main

import (
	"time"

	"github.com/nsf/termbox-go"
)

var quit = false

func startCli() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	termbox.SetInputMode(termbox.InputEsc)
	redraw()

	// capture and process events from the CLI
	eventChan := make(chan termbox.Event, 16)
	go handleEvents(eventChan)
	go func() {
		for {
			ev := termbox.PollEvent()
			eventChan <- ev
		}
	}()

	timer := time.Tick(time.Millisecond * 100)
	for {
		select {
		case <-timer:
			redraw()
		}
	}
}

func handleEvents(eventChan chan termbox.Event) {
	for {
		ev := <-eventChan
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlQ, termbox.KeyEsc:
				goto endfunc
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
endfunc:
	quit = true
}
