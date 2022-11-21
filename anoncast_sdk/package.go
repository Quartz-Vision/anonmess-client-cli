package anoncastsdk

import (
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

const UUID_SIZE = 16

type Package struct {
	ChannelId uuid.UUID
	Event     *events.Event
}

func (p *Package) MarshalBinary() (data []byte, err error) {
	eventEnc, err := p.Event.MarshalBinary()
	if err != nil {
		return data, err
	}
	eventSize := len(eventEnc)
	eventSizeEnc := utils.Int64ToBytes(int64(eventSize))

	packageSize := UUID_SIZE + len(eventSizeEnc) + eventSize
	packageSizeEnc := utils.Int64ToBytes(int64(packageSize))

	encodedData := make([]byte, len(packageSizeEnc)+packageSize)

	utils.JoinSlices(encodedData, packageSizeEnc, p.ChannelId[:], eventSizeEnc, eventEnc)

	return encodedData, nil
}

func (p *Package) UnmarshalBinary(data []byte) (err error) {
	_, packageSizeLen := utils.BytesToInt64(data)

	copy(p.ChannelId[:], data[packageSizeLen:packageSizeLen+UUID_SIZE])

	eventSize, eventSizeLen := utils.BytesToInt64(data[packageSizeLen+UUID_SIZE:])

	p.Event = &events.Event{}
	skip := packageSizeLen + UUID_SIZE + eventSizeLen
	if err := p.Event.UnmarshalBinary(data[skip : skip+int(eventSize)]); err != nil {
		return err
	}

	return nil
}
