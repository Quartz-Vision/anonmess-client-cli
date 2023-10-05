package anoncastsdk

import (
	"errors"
	"fmt"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/Quartz-Vision/goslice"
	"github.com/google/uuid"
)

type DataPackage struct {
	ChannelId uuid.UUID
	Payload   []byte
}

var ErrNoKeyPack = errors.New("no key pack for the channel")

func (p *DataPackage) MarshalBinary() (data []byte, err error) {
	channelId := make([]byte, utils.UUID_SIZE)
	copy(channelId, p.ChannelId[:])

	payload := make([]byte, len(p.Payload))
	copy(payload, p.Payload)

	var (
		keyPack              *keysstorage.KeyPack
		ok                   bool
		idKeyPos             int64
		payloadKeyPos        int64
		payloadKeyPosEnc     []byte
		payloadKeyPosEncSize int
		payloadSize          int
		payloadSizeEnc       []byte
		payloadSizeEncSize   int
		idKey                []byte
		payloadKey           []byte
	)

	utils.UntilErrorPointer(
		&err,
		// Key Pack
		func() {
			keyPack, ok = keysstorage.GetKeyPack(p.ChannelId)
			if !ok {
				err = ErrNoKeyPack
			}
		},
		// Event and Payload Key
		func() {
			payloadSize = len(payload)
			payloadSizeEnc = utils.Int64ToBytes(int64(payloadSize))
			payloadSizeEncSize = len(payloadSizeEnc)
		},
		func() {
			payloadKeyPos, err = keyPack.KeyPositions[keysstorage.PACK_PAYLOAD_KEY].Take(
				int64(payloadSize + payloadSizeEncSize),
			)
			payloadKeyPosEnc = utils.Int64ToBytes(payloadKeyPos)
			payloadKeyPosEncSize = len(payloadKeyPosEnc)
		},
		func() {
			payloadKey = make([]byte, payloadSize+payloadSizeEncSize)
			_, err = keyPack.Keys[keysstorage.PACK_OUT+keysstorage.PACK_PAYLOAD_KEY].ReadAt(payloadKey, payloadKeyPos)
		},
		// ID Key
		func() {
			idKeyPos, err = keyPack.KeyPositions[keysstorage.PACK_ID_KEY].Take(utils.UUID_SIZE + int64(payloadKeyPosEncSize))
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

	packageSize := len(idKeyPosEnc) + utils.UUID_SIZE + payloadKeyPosEncSize + payloadSizeEncSize + payloadSize
	packageSizeEnc := utils.Int64ToBytes(int64(packageSize))

	encodedData := make([]byte, len(packageSizeEnc)+packageSize)

	fmt.Printf("<<< Id key frag: %v\n", idKey[utils.UUID_SIZE:utils.UUID_SIZE+utils.INT_MAX_SIZE])
	fmt.Printf("<<< Id key pos: %v\n", idKeyPos)
	fmt.Printf("<<< Payload key frag: %v\n", payloadKey[:utils.INT_MAX_SIZE])
	fmt.Printf("<<< Payload key pos: %v\n", payloadKeyPos)
	fmt.Printf("<<< Event size: %v\n", payloadSize)
	fmt.Printf("<<< Event dec frag: %v\n", payload[:utils.INT_MAX_SIZE])

	goslice.Xor(channelId, idKey[:utils.UUID_SIZE])
	goslice.Xor(payloadKeyPosEnc, idKey[utils.UUID_SIZE:])
	goslice.Xor(payloadSizeEnc, payloadKey[:payloadSizeEncSize])
	goslice.Xor(payload, payloadKey[payloadSizeEncSize:])

	fmt.Printf("<<< Payload key pos enc frag: %v\n", payloadKeyPosEnc)
	fmt.Printf("<<< Event enc frag: %v\n", payload[:utils.INT_MAX_SIZE])
	fmt.Printf("<<< Event payload key frag: %v\n", payloadKey[utils.INT_MAX_SIZE:utils.INT_MAX_SIZE*2])

	goslice.Join(
		encodedData,
		packageSizeEnc,
		idKeyPosEnc,
		channelId,
		payloadKeyPosEnc,
		payloadSizeEnc,
		payload,
	)

	return encodedData, nil
}

var ErrKeyPackIdDecodeFailed = errors.New("no such key pack")

func (p *DataPackage) UnmarshalBinary(data []byte) (err error) {
	_, packageSizeLen := utils.BytesToInt64(data)
	data = data[packageSizeLen:]

	idKeyPos, idKeyPosLen := utils.BytesToInt64(data[:utils.INT_MAX_SIZE])
	data = data[idKeyPosLen:]

	fmt.Printf(">>> Id key pos: %v\n", idKeyPos)

	var ok bool
	// Don't use utils for the error here since we need maximal speed
	if p.ChannelId, ok = keysstorage.TryDecodePackId(idKeyPos, data[:utils.UUID_SIZE]); !ok {
		return ErrKeyPackIdDecodeFailed
	}
	data = data[utils.UUID_SIZE:]

	var (
		keyPack          *keysstorage.KeyPack
		payloadKeyPos    int64
		payloadKeyPosLen int
		payloadSize      int64
		payloadSizeLen   int
		idKey            []byte
		payloadKey       []byte
		tmpNum           = make([]byte, utils.INT_MAX_SIZE)
	)

	return utils.UntilErrorPointer(
		&err,
		// Key Pack
		func() {
			keyPack, ok = keysstorage.GetKeyPack(p.ChannelId)
			if !ok {
				err = ErrNoKeyPack
			}
		},
		// ID and Payload Key
		func() {
			idKey = make([]byte, utils.INT_MAX_SIZE)
			_, err = keyPack.Keys[keysstorage.PACK_IN+keysstorage.PACK_ID_KEY].ReadAt(idKey, idKeyPos+utils.UUID_SIZE)

			fmt.Printf(">>> Id key frag: %v\n", idKey[:utils.INT_MAX_SIZE])
		},
		func() {
			fmt.Printf(">>> Payload key pos enc frag: %v\n", data[:utils.INT_MAX_SIZE])

			goslice.SetResult(tmpNum, goslice.Xor, data[:utils.INT_MAX_SIZE], idKey)
			payloadKeyPos, payloadKeyPosLen = utils.BytesToInt64(tmpNum)
			data = data[payloadKeyPosLen:]

			fmt.Printf(">>> Payload key pos: %v\n", payloadKeyPos)

			payloadKey = make([]byte, utils.INT_MAX_SIZE)
			_, err = keyPack.Keys[keysstorage.PACK_IN+keysstorage.PACK_PAYLOAD_KEY].ReadAt(payloadKey, payloadKeyPos)
		},
		func() {
			fmt.Printf(">>> Payload key frag: %v\n", payloadKey[:utils.INT_MAX_SIZE])

			goslice.SetResult(tmpNum, goslice.Xor, data[:utils.INT_MAX_SIZE], payloadKey)
			payloadSize, payloadSizeLen = utils.BytesToInt64(tmpNum)
			data = data[payloadSizeLen:]

			fmt.Printf(">>> Payload size: %v\n", payloadSize)

			payloadKey = make([]byte, payloadSize)
			_, err = keyPack.Keys[keysstorage.PACK_IN+keysstorage.PACK_PAYLOAD_KEY].ReadAt(payloadKey, payloadKeyPos+int64(payloadSizeLen))
		},
		func() {
			fmt.Printf(">>> Event enc frag: %v\n", data[:utils.INT_MAX_SIZE])
			fmt.Printf(">>> Event payload key frag: %v\n", payloadKey[:utils.INT_MAX_SIZE])

			goslice.SetResult(p.Payload, goslice.Xor, data[:payloadSize], payloadKey[:payloadSize])

			fmt.Printf(">>> Event dec frag: %v\n", data[:utils.INT_MAX_SIZE])
		},
	)
}
