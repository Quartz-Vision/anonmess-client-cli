package keysstorage

import (
	"path/filepath"
	quartzSymmetric "quartzvision/anonmess-client-cli/crypto/symmetric"
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type KeyPack struct {
	idKey      *KeyBuffer
	payloadKey *KeyBuffer
}

func NewKeyPack(packId uuid.UUID, prefix string) (keyPack *KeyPack, err error) {
	keyPack = &KeyPack{}

	if utils.UntilErrorPointer(
		&err,
		func() { keyPack.idKey, err = NewKeyBuffer(keyPath(packId, prefix, PACK_PREFIX_ID_KEY)) },
		func() { keyPack.payloadKey, err = NewKeyBuffer(keyPath(packId, prefix, PACK_PREFIX_PAYLOAD_KEY)) },
	) != nil {
		keyPack.Close()
		return nil, err
	}

	size, err := keyPack.idKey.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		err = utils.UntilFirstError(
			func() error { return keyPack.idKey.GenerateKey(settings.Config.KeysStartSizeB) },
			func() error { return keyPack.payloadKey.GenerateKey(settings.Config.KeysStartSizeB) },
		)
	}

	return keyPack, err
}

func (obj *KeyPack) Close() {
	safeClose(obj.idKey)
	safeClose(obj.payloadKey)
}

type KeyIOPack struct {
	InKeys  *KeyPack
	OutKeys *KeyPack
}

func NewKeyIOPack(packId uuid.UUID) (keyIOPack *KeyIOPack, err error) {
	keyIOPack = &KeyIOPack{}

	utils.UntilErrorPointer(
		&err,
		func() { keyIOPack.InKeys, err = NewKeyPack(packId, PACK_PREFIX_IN) },
		func() { keyIOPack.OutKeys, err = NewKeyPack(packId, PACK_PREFIX_OUT) },
	)

	if err != nil {
		keyIOPack.Close()
	}

	return keyIOPack, err
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

	return nil
}

func UnmanageKeyPack(packId uuid.UUID) {
	if pack, ok := Packs[packId]; ok {
		safeClose(pack)
		delete(Packs, packId)
	}
}

func ExportSharedKeys(packId uuid.UUID, dest string) (err error) {
	pack := Packs[packId]
	packageFile, err := filestorage.NewFile(filepath.Join(dest, "_package_"), true, 0o600)
	if err != nil {
		return err
	}

	_, err = packageFile.Write(packId[:])
	if err != nil {
		return err
	}
	packageFile.Close()

	pack.InKeys.idKey.SaveTo(filepath.Join(dest, PACK_PREFIX_OUT+"-"+PACK_PREFIX_ID_KEY))
	pack.InKeys.payloadKey.SaveTo(filepath.Join(dest, PACK_PREFIX_OUT+"-"+PACK_PREFIX_PAYLOAD_KEY))
	pack.OutKeys.idKey.SaveTo(filepath.Join(dest, PACK_PREFIX_IN+"-"+PACK_PREFIX_ID_KEY))
	pack.OutKeys.payloadKey.SaveTo(filepath.Join(dest, PACK_PREFIX_IN+"-"+PACK_PREFIX_PAYLOAD_KEY))

	return
}

func TryDecodePackId(idKeyPos int64, encId []byte) (id uuid.UUID, ok bool) {
	idLen := int64(len(encId))
	tmpEncId := make([]byte, idLen)
	key := make([]byte, idLen)

	for id := range Packs {
		copy(tmpEncId, encId)

		_, err := Packs[id].InKeys.idKey.ReadAt(key, idKeyPos)

		if err == nil && quartzSymmetric.Decode(tmpEncId, key) == nil && utils.AreSlicesEqual(tmpEncId, id[:]) {
			return id, true
		}
	}

	return id, false
}
