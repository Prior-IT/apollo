package main

import (
	"fmt"
)

type eventKind string

const (
	evFileChanged eventKind = "file_changed"
	evRestart     eventKind = "restart"
	evRefresh     eventKind = "refresh"
	evStop        eventKind = "stop"
	evCommandDone eventKind = "command_done"
	evError       eventKind = "error"
)

type event struct {
	kind    eventKind
	payload string
}

func (e event) String() string {
	return fmt.Sprintf("%v (%s)", e.kind, e.payload)
}

var (
	events    chan<- event
	listeners []chan event
)

func initializeEventSystem() {
	ev := make(chan event)
	go func() {
		for e := range ev {
			for _, l := range listeners {
				l <- e
			}
		}
	}()
	events = ev
}

// Trigger a function on every event of the specified kind.
// If f returns false, the handler will be removed. If it returns true, it will keep getting called
// for every new event of the specified kind.
// **Note:** This will not run in a goroutine, call `go onEvent` if you don't want to block the current
// thread until the event has been received (and the handler returned false).
func onEvent(kind eventKind, f func(ev event) bool) {
	list := newEventListener()
	for e := range list {
		if e.kind == kind && !f(e) {
			return
		}
	}
}

// Send a new FileChanged event
func newFileChangedEvent(file string) {
	events <- event{
		kind:    evFileChanged,
		payload: file,
	}
}

// Send a new Stop event
func newStopEvent() {
	events <- event{
		kind: evStop,
	}
}

// Send a new CommandDone event
func newCommandDoneEvent(name string) {
	events <- event{
		kind:    evCommandDone,
		payload: name,
	}
}

func newEventListener() <-chan event {
	list := make(chan event)
	listeners = append(listeners, list)
	return list
}
