package anoncastsdk

const (
	EVENT_MESSAGE EventType = iota
)

func initEvents() {
	initMessageEvent(EVENT_MESSAGE)
}
