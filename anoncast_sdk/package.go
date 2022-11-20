package anoncastsdk

import "github.com/google/uuid"

type Package struct {
	ChatId uuid.UUID
	Event  *Event
}

func (p *Package) MarshalBinary() (data []byte, err error) {
	encodedEvent, err := p.Event.MarshalBinary()
	if err != nil {
		return data, err
	}
	encodedEventLen := Int64ToBytes(int64(len(encodedEvent)))
	encodedPackageLen := Int64ToBytes(int64(len(encodedEventLen) + len(encodedEvent)))

	encodedData := make([]byte, len(encodedPackageLen)+len(encodedEventLen)+len(encodedEvent))

	copy(encodedData, encodedPackageLen)
	copy(encodedData[len(encodedPackageLen):], encodedEventLen)
	copy(encodedData[len(encodedPackageLen)+len(encodedEventLen):], encodedEvent)

	return encodedData, nil
}

func (p *Package) UnmarshalBinary(data []byte) (err error) {
	_, packageLenSize := BytesToInt64(data)
	_, eventLenSize := BytesToInt64(data[packageLenSize:])
	p.Event = &Event{}
	if err := p.Event.UnmarshalBinary(data[packageLenSize+eventLenSize:]); err != nil {
		return err
	}

	return nil
}
