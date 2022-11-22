package anoncastsdk

import (
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type dataPackage struct {
	channelId uuid.UUID
	event     *events.Event

	client *Client
}

func (p *dataPackage) MarshalBinary() (data []byte, err error) {
	eventEnc, err := p.event.MarshalBinary()
	if err != nil {
		return data, err
	}
	eventSize := len(eventEnc)
	eventSizeEnc := utils.Int64ToBytes(int64(eventSize))

	packageSize := utils.UUID_SIZE + len(eventSizeEnc) + eventSize
	packageSizeEnc := utils.Int64ToBytes(int64(packageSize))

	encodedData := make([]byte, len(packageSizeEnc)+packageSize)

	utils.JoinSlices(encodedData, packageSizeEnc, p.channelId[:], eventSizeEnc, eventEnc)

	return encodedData, nil
}

func (p *dataPackage) UnmarshalBinary(data []byte) (err error) {
	_, packageSizeLen := utils.BytesToInt64(data)

	copy(p.channelId[:], data[packageSizeLen:packageSizeLen+utils.UUID_SIZE])

	eventSize, eventSizeLen := utils.BytesToInt64(data[packageSizeLen+utils.UUID_SIZE:])

	skip := packageSizeLen + utils.UUID_SIZE + eventSizeLen
	return p.event.UnmarshalBinary(data[skip : skip+int(eventSize)])
}
