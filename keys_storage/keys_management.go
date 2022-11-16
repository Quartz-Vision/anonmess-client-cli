package keysstorage

import (
	quartzSymmetric "quartzvision/anonmess-client-cli/crypto/symmetric"
	sliceutils "quartzvision/anonmess-client-cli/slice_utils"

	"github.com/google/uuid"
)

type KeyPack struct {
	IdKey      *KeyBuffer
	PayloadKey *KeyBuffer
}

func NewKeyPack(packId uuid.UUID, prefix string) (keyPack *KeyPack, err error) {
	keyPack = &KeyPack{}

	if buf, err := NewKeyBuffer(keyPath(packId, prefix, PACK_PREFIX_ID_KEY)); err != nil {
		return nil, err
	} else {
		keyPack.IdKey = buf
	}

	if buf, err := NewKeyBuffer(keyPath(packId, prefix, PACK_PREFIX_PAYLOAD_KEY)); err != nil {
		return nil, err
	} else {
		keyPack.PayloadKey = buf
	}

	return keyPack, nil
}

func (obj *KeyPack) Close() {
	safeClose(obj.IdKey)
	safeClose(obj.PayloadKey)
}

type KeyIOPack struct {
	InKeys  *KeyPack
	OutKeys *KeyPack
}

func NewKeyIOPack(packId uuid.UUID) (keyIOPack *KeyIOPack, err error) {
	keyIOPack = &KeyIOPack{}

	if pack, err := NewKeyPack(packId, PACK_PREFIX_IN); err != nil {
		return nil, err
	} else {
		keyIOPack.InKeys = pack
	}

	if pack, err := NewKeyPack(packId, PACK_PREFIX_OUT); err != nil {
		return nil, err
	} else {
		keyIOPack.OutKeys = pack
	}

	return keyIOPack, nil
}

func (obj *KeyIOPack) Close() {
	safeClose(obj.InKeys)
	safeClose(obj.OutKeys)
}

var Packs = map[uuid.UUID]*KeyIOPack{}

func ManageKeyPack(packId uuid.UUID) (err error) {
	if _, ok := Packs[packId]; ok {
		return nil
	}

	if pack, err := NewKeyIOPack(packId); err != nil {
		return err
	} else {
		Packs[packId] = pack
	}

	return err
}

func UnmanageKeyPack(packId uuid.UUID) {
	if pack, ok := Packs[packId]; ok {
		safeClose(pack)
		delete(Packs, packId)
	}
}

func TryDecodePackId(idKeyPos int64, encId []byte) (id uuid.UUID, ok bool) {
	idLen := int64(len(encId))

	for id, pack := range Packs {
		tmpEncId := make([]byte, idLen)
		copy(tmpEncId, encId)

		key, err := pack.InKeys.IdKey.GetKeySlice(idKeyPos, idLen)

		if err == nil && quartzSymmetric.Decode(tmpEncId, key) == nil && sliceutils.IsEqual(tmpEncId, id[:]) {
			return id, true
		}
	}

	return id, false
}
