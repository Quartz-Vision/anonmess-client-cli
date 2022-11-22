package events

import (
	"errors"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type Event struct {
	Type      EventType
	ChannelId uuid.UUID // not marshaled
	Data      any

	manager *EventManager // not marshaled
}

func (m *EventManager) Manage(e *Event) (managed *Event) {
	e.manager = m
	return e
}

var ErrNoSerializer = errors.New("there is no such serializer")
var ErrEncodeFailed = errors.New("event encoding failed")

// Implements encoding.BinaryMarshaler
func (e *Event) MarshalBinary() (data []byte, err error) {
	data = utils.Int64ToBytes(int64(e.Type))

	if serializer, ok := e.manager.eventSerializers[e.Type]; !ok {
		return data, ErrNoSerializer
	} else if serialized, err := serializer.encode(e); err != nil {
		return data, err
	} else {
		return append(data, serialized...), nil
	}
}

// Implements encoding.BinaryUnmarshaler
func (e *Event) UnmarshalBinary(data []byte) (err error) {
	eventType, sizeRead := utils.BytesToInt64(data)
	e.Type = EventType(eventType)

	if serializer, ok := e.manager.eventSerializers[e.Type]; !ok {
		return ErrNoSerializer
	} else {
		return serializer.decode(e, data[sizeRead:])
	}
}

// Checks whether the event has a serializer or not
func (e *Event) IsSerializable() (ok bool) {
	_, ok = e.manager.eventSerializers[e.Type]
	return ok
}
