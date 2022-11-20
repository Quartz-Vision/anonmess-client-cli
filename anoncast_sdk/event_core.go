package anoncastsdk

import (
	"errors"
)

type EventType int8

type Event struct {
	Type EventType
	Data any
}

type EventMarshalFn func(e *Event) (dst []byte, err error)
type EventUnmarshalFn func(e *Event, src []byte) (err error)
type EventSerializer struct {
	Encode EventMarshalFn
	Decode EventUnmarshalFn
}

var ErrNoSerializer = errors.New("there is no such serializer")
var ErrEncodeFailed = errors.New("event encoding failed")

// var ErrDecodeFailed = errors.New("data decoding failed")

var EventSerializers map[EventType]EventSerializer = map[EventType]EventSerializer{}

func (e *Event) MarshalBinary() (data []byte, err error) {
	data = Int64ToBytes(int64(e.Type))

	if serializer, ok := EventSerializers[e.Type]; !ok {
		return data, ErrNoSerializer
	} else if serialized, err := serializer.Encode(e); err != nil {
		return data, err
	} else {
		return append(data, serialized...), nil
	}
}

func (e *Event) UnmarshalBinary(data []byte) (err error) {
	eventType, sizeRead := BytesToInt64(data)
	e.Type = EventType(eventType)

	if serializer, ok := EventSerializers[e.Type]; !ok {
		return ErrNoSerializer
	} else {
		return serializer.Decode(e, data[sizeRead:])
	}
}
