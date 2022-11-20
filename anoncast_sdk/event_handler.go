package anoncastsdk

import "quartzvision/anonmess-client-cli/lists/squeue"

type EventHandlerFn func(e *Event)

var EventHandlers map[EventType]EventHandlerFn = map[EventType]EventHandlerFn{}
var EventsToSend = squeue.New()

func EmitEvent(e *Event) {
	EventHandlers[e.Type](e)
}

func SendEvent(e *Event) {
	EventsToSend.Push(e)
}
