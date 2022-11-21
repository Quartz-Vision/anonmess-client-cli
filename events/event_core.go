package events

import (
	"errors"
	"hash/crc32"
	"hash/crc64"
	"quartzvision/anonmess-client-cli/utils"
)

type EventType uint64

var eventTypeCrcTable = crc64.MakeTable(crc32.IEEE)

func CreateEventType(name string) EventType {
	return EventType(crc64.Checksum([]byte(name), eventTypeCrcTable))
}

type Event struct {
	Type EventType
	Data any
}

type EventMarshalFn func(e *Event) (dst []byte, err error)
type EventUnmarshalFn func(e *Event, src []byte) (err error)
type eventSerializer struct {
	encode EventMarshalFn
	decode EventUnmarshalFn
}

var ErrNoSerializer = errors.New("there is no such serializer")
var ErrEncodeFailed = errors.New("event encoding failed")

// var ErrDecodeFailed = errors.New("data decoding failed")

var eventSerializers map[EventType]*eventSerializer = map[EventType]*eventSerializer{}

func SetEventCoders(etype EventType, encoder EventMarshalFn, decoder EventUnmarshalFn) {
	eventSerializers[etype] = &eventSerializer{
		encode: encoder,
		decode: decoder,
	}
}

// Implements encoding.BinaryMarshaler
func (e *Event) MarshalBinary() (data []byte, err error) {
	data = utils.Int64ToBytes(int64(e.Type))

	if serializer, ok := eventSerializers[e.Type]; !ok {
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

	if serializer, ok := eventSerializers[e.Type]; !ok {
		return ErrNoSerializer
	} else {
		return serializer.decode(e, data[sizeRead:])
	}
}
