package anoncastsdk

import (
	"errors"
	"fmt"
	"quartzvision/anonmess-client-cli/events"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type dataPackage struct {
	channelId uuid.UUID
	event     *events.Event

	client *Client
}

var ErrNoKeyPack = errors.New("no key pack for the channel")

func (p *dataPackage) MarshalBinary() (data []byte, err error) {
	channelId := make([]byte, utils.UUID_SIZE)
	copy(channelId, p.channelId[:])

	var (
		keyPack              *keysstorage.KeyPack
		ok                   bool
		idKeyPos             int64
		payloadKeyPos        int64
		payloadKeyPosEnc     []byte
		payloadKeyPosEncSize int
		eventSize            int
		eventSizeEnc         []byte
		eventSizeEncSize     int
		idKey                []byte
		payloadKey           []byte
		eventEnc             []byte
	)

	utils.UntilErrorPointer(
		&err,
		// Key Pack
		func() {
			keyPack, ok = keysstorage.GetKeyPack(p.channelId)
			if !ok {
				err = ErrNoKeyPack
			}
		},
		// Event and Payload Key
		func() {
			eventEnc, err = p.event.MarshalBinary()
			eventSize = len(eventEnc)
			eventSizeEnc = utils.Int64ToBytes(int64(eventSize))
			eventSizeEncSize = len(eventSizeEnc)
		},
		func() {
			payloadKeyPos, err = keyPack.KeyPostions[keysstorage.PACK_PAYLOAD_KEY].Take(
				int64(eventSize + eventSizeEncSize),
			)
			payloadKeyPosEnc = utils.Int64ToBytes(payloadKeyPos)
			payloadKeyPosEncSize = len(payloadKeyPosEnc)
		},
		func() {
			payloadKey = make([]byte, eventSize+eventSizeEncSize)
			_, err = keyPack.Keys[keysstorage.PACK_OUT+keysstorage.PACK_PAYLOAD_KEY].ReadAt(payloadKey, payloadKeyPos)
		},
		// ID Key
		func() {
			idKeyPos, err = keyPack.KeyPostions[keysstorage.PACK_ID_KEY].Take(utils.UUID_SIZE + int64(payloadKeyPosEncSize))
		},
		func() {
			idKey = make([]byte, utils.UUID_SIZE+payloadKeyPosEncSize)
			_, err = keyPack.Keys[keysstorage.PACK_OUT+keysstorage.PACK_ID_KEY].ReadAt(idKey, idKeyPos)
		},
	)

	if err != nil {
		return nil, err
	}

	idKeyPosEnc := utils.Int64ToBytes(idKeyPos)

	packageSize := len(idKeyPosEnc) + utils.UUID_SIZE + payloadKeyPosEncSize + eventSizeEncSize + eventSize
	packageSizeEnc := utils.Int64ToBytes(int64(packageSize))

	encodedData := make([]byte, len(packageSizeEnc)+packageSize)

	fmt.Printf("<<< Id key frag: %v\n", idKey[utils.UUID_SIZE:utils.UUID_SIZE+utils.INT_MAX_SIZE])
	fmt.Printf("<<< Id key pos: %v\n", idKeyPos)
	fmt.Printf("<<< Payload key frag: %v\n", payloadKey[:utils.INT_MAX_SIZE])
	fmt.Printf("<<< Payload key pos: %v\n", payloadKeyPos)
	fmt.Printf("<<< Event size: %v\n", eventSize)
	fmt.Printf("<<< Event dec frag: %v\n", eventEnc[:utils.INT_MAX_SIZE])

	utils.XorSlices(channelId, idKey[:utils.UUID_SIZE])
	utils.XorSlices(payloadKeyPosEnc, idKey[utils.UUID_SIZE:])
	utils.XorSlices(eventSizeEnc, payloadKey[:eventSizeEncSize])
	utils.XorSlices(eventEnc, payloadKey[eventSizeEncSize:])

	fmt.Printf("<<< Payload key pos enc frag: %v\n", payloadKeyPosEnc)
	fmt.Printf("<<< Event enc frag: %v\n", eventEnc[:utils.INT_MAX_SIZE])
	fmt.Printf("<<< Event payload key frag: %v\n", payloadKey[utils.INT_MAX_SIZE:utils.INT_MAX_SIZE*2])

	utils.JoinSlices(
		encodedData,
		packageSizeEnc,
		idKeyPosEnc,
		channelId,
		payloadKeyPosEnc,
		eventSizeEnc,
		eventEnc,
	)

	return encodedData, nil
}

var ErrKeyPackIdDecodeFailed = errors.New("no such key pack")

func (p *dataPackage) UnmarshalBinary(data []byte) (err error) {
	_, packageSizeLen := utils.BytesToInt64(data)
	data = data[packageSizeLen:]

	idKeyPos, idKeyPosLen := utils.BytesToInt64(data[:utils.INT_MAX_SIZE])
	data = data[idKeyPosLen:]

	fmt.Printf(">>> Id key pos: %v\n", idKeyPos)

	var ok bool
	// Don't use utils for the error here since we need maximal speed
	if p.channelId, ok = keysstorage.TryDecodePackId(idKeyPos, data[:utils.UUID_SIZE]); !ok {
		return ErrKeyPackIdDecodeFailed
	}
	data = data[utils.UUID_SIZE:]

	var (
		keyPack          *keysstorage.KeyPack
		payloadKeyPos    int64
		payloadKeyPosLen int
		eventSize        int64
		eventSizeLen     int
		idKey            []byte
		payloadKey       []byte
		tmpNum           = make([]byte, utils.INT_MAX_SIZE)
	)

	return utils.UntilErrorPointer(
		&err,
		// Key Pack
		func() {
			keyPack, ok = keysstorage.GetKeyPack(p.channelId)
			if !ok {
				err = ErrNoKeyPack
			}
		},
		// Event and Payload Key
		func() {
			idKey = make([]byte, utils.INT_MAX_SIZE)
			_, err = keyPack.Keys[keysstorage.PACK_IN+keysstorage.PACK_ID_KEY].ReadAt(idKey, idKeyPos+utils.UUID_SIZE)

			fmt.Printf(">>> Id key frag: %v\n", idKey[:utils.INT_MAX_SIZE])
		},
		func() {
			fmt.Printf(">>> Payload key pos enc frag: %v\n", data[:utils.INT_MAX_SIZE])
			utils.ProcessSlices(tmpNum, utils.XorSlices, data[:utils.INT_MAX_SIZE], idKey)
			payloadKeyPos, payloadKeyPosLen = utils.BytesToInt64(tmpNum)
			data = data[payloadKeyPosLen:]

			fmt.Printf(">>> Payload key pos: %v\n", payloadKeyPos)

			payloadKey = make([]byte, utils.INT_MAX_SIZE)
			_, err = keyPack.Keys[keysstorage.PACK_IN+keysstorage.PACK_PAYLOAD_KEY].ReadAt(payloadKey, payloadKeyPos)
		},
		func() {
			fmt.Printf(">>> Payload key frag: %v\n", payloadKey[:utils.INT_MAX_SIZE])

			utils.ProcessSlices(tmpNum, utils.XorSlices, data[:utils.INT_MAX_SIZE], payloadKey)
			eventSize, eventSizeLen = utils.BytesToInt64(tmpNum)
			data = data[eventSizeLen:]

			fmt.Printf(">>> Event size: %v\n", eventSize)

			payloadKey = make([]byte, eventSize)
			_, err = keyPack.Keys[keysstorage.PACK_IN+keysstorage.PACK_PAYLOAD_KEY].ReadAt(payloadKey, payloadKeyPos+int64(eventSizeLen))
		},
		func() {
			fmt.Printf(">>> Event enc frag: %v\n", data[:utils.INT_MAX_SIZE])
			fmt.Printf(">>> Event payload key frag: %v\n", payloadKey[:utils.INT_MAX_SIZE])
			utils.XorSlices(data[:eventSize], payloadKey[:eventSize])
			fmt.Printf(">>> Event dec frag: %v\n", data[:utils.INT_MAX_SIZE])
			err = p.event.UnmarshalBinary(data[:eventSize])
		},
	)
}
