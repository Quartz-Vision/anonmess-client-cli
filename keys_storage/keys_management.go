package keysstorage

import (
	"errors"
	quartzSymmetric "quartzvision/anonmess-client-cli/crypto/symmetric"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

var packs = map[uuid.UUID]*KeyPack{}
var ErrPackageExists = errors.New("this package already exists")

func ManageKeyPack(packId uuid.UUID) (pack *KeyPack, err error) {
	pack, ok := packs[packId]
	if !ok {
		pack, err = newKeyPack(packId)
		if err == nil {
			packs[packId] = pack
		}
	} else {
		err = ErrPackageExists
	}

	return pack, err
}

func UnmanageKeyPack(packId uuid.UUID) {
	if pack, ok := packs[packId]; ok {
		safeClose(pack)
		delete(packs, packId)
	}
}

func ManageSharedKeyPack(src string) (pack *KeyPack, err error) {
	var packId = uuid.UUID{}
	var ok bool

	utils.UntilErrorPointer(
		&err,
		func() { packId, err = getSharedPackId(src) },
		func() {
			if pack, ok = packs[packId]; ok {
				err = ErrPackageExists
			}
		},
		func() { pack, err = importSharedKeyPack(packId, src) },
		func() { packs[packId] = pack },
	)

	return pack, err
}

func GetKeyPack(packId uuid.UUID) (pack *KeyPack, ok bool) {
	pack, ok = packs[packId]
	return pack, ok
}

func TryDecodePackId(idKeyPos int64, encId []byte) (id uuid.UUID, ok bool) {
	idLen := int64(len(encId))
	tmpEncId := make([]byte, idLen)
	key := make([]byte, idLen)

	for id := range packs {
		copy(tmpEncId, encId)

		_, err := packs[id].Keys[PACK_IN+PACK_ID_KEY].ReadAt(key, idKeyPos)

		if err == nil && quartzSymmetric.Decode(tmpEncId, key) == nil && utils.AreSlicesEqual(tmpEncId, id[:]) {
			return id, true
		}
	}

	return id, false
}
