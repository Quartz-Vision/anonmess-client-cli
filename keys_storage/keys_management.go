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
	IdKey         *KeyBuffer
	PayloadKey    *KeyBuffer
	prefix        string
	sharingPrefix string
}

func NewKeyPack(packId uuid.UUID, prefix string, sharingPrefix string, fill bool) (keyPack *KeyPack, err error) {
	keyPack = &KeyPack{
		prefix:        prefix,
		sharingPrefix: sharingPrefix,
	}

	if utils.UntilErrorPointer(
		&err,
		func() { keyPack.IdKey, err = NewKeyBuffer(keyPath(packId, prefix, PACK_PREFIX_ID_KEY)) },
		func() { keyPack.PayloadKey, err = NewKeyBuffer(keyPath(packId, prefix, PACK_PREFIX_PAYLOAD_KEY)) },
	) != nil {
		keyPack.Close()
		return nil, err
	}

	size, err := keyPack.IdKey.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 && fill {
		err = utils.UntilFirstError(
			func() error { return keyPack.IdKey.GenerateKey(settings.Config.KeysStartSizeB) },
			func() error { return keyPack.PayloadKey.GenerateKey(settings.Config.KeysStartSizeB) },
		)
		if err != nil {
			keyPack.Close()
		}
	}

	return keyPack, err
}

func (p *KeyPack) ExportShared(dest string) (err error) {
	return utils.UntilFirstError(
		func() error {
			file, err := filestorage.NewFile(filepath.Join(dest, keyFileName(p.sharingPrefix, PACK_PREFIX_ID_KEY)), true, 0o600)
			if err == nil {
				err = p.IdKey.PipeTo(file, settings.Config.KeysBufferSizeB)
			}
			return err
		},
		func() error {
			file, err := filestorage.NewFile(filepath.Join(dest, keyFileName(p.sharingPrefix, PACK_PREFIX_PAYLOAD_KEY)), true, 0o600)
			if err == nil {
				err = p.PayloadKey.PipeTo(file, settings.Config.KeysBufferSizeB)
			}
			return err
		},
	)
}

func (p *KeyPack) ImportShared(src string) (err error) {
	return utils.UntilFirstError(
		func() error {
			file, err := filestorage.NewFile(filepath.Join(src, keyFileName(p.prefix, PACK_PREFIX_ID_KEY)), false, 0o600)
			if err == nil {
				err = file.PipeTo(p.IdKey, settings.Config.KeysBufferSizeB)
			}
			return err
		},
		func() error {
			file, err := filestorage.NewFile(filepath.Join(src, keyFileName(p.prefix, PACK_PREFIX_PAYLOAD_KEY)), false, 0o600)
			if err == nil {
				err = file.PipeTo(p.PayloadKey, settings.Config.KeysBufferSizeB)
			}
			return err
		},
	)
}

func (p *KeyPack) Close() {
	safeClose(p.IdKey)
	safeClose(p.PayloadKey)
}

type KeyIOPack struct {
	InKeys  *KeyPack
	OutKeys *KeyPack
}

func NewKeyIOPack(packId uuid.UUID, fill bool) (keyIOPack *KeyIOPack, err error) {
	keyIOPack = &KeyIOPack{}

	utils.UntilErrorPointer(
		&err,
		func() { keyIOPack.InKeys, err = NewKeyPack(packId, PACK_PREFIX_IN, PACK_PREFIX_OUT, fill) },
		func() { keyIOPack.OutKeys, err = NewKeyPack(packId, PACK_PREFIX_OUT, PACK_PREFIX_IN, fill) },
	)

	if err != nil {
		keyIOPack.Close()
	}

	return keyIOPack, err
}

func (p *KeyIOPack) ExportShared(dest string) (err error) {
	return utils.UntilFirstError(
		func() error { return p.InKeys.ExportShared(dest) },
		func() error { return p.OutKeys.ExportShared(dest) },
	)
}

func (p *KeyIOPack) ImportShared(src string) (err error) {
	return utils.UntilFirstError(
		func() error { return p.InKeys.ImportShared(src) },
		func() error { return p.OutKeys.ImportShared(src) },
	)
}

func (obj *KeyIOPack) Close() {
	safeClose(obj.InKeys)
	safeClose(obj.OutKeys)
}

var Packs = map[uuid.UUID]*KeyIOPack{}

func ManageKeyPack(packId uuid.UUID, fill bool) (err error) {
	if _, ok := Packs[packId]; ok {
		return nil
	}

	if pack, err := NewKeyIOPack(packId, fill); err != nil {
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

	return pack.ExportShared(dest)
}

func ImportSharedKeys(src string) (packId uuid.UUID, err error) {
	packageFile, err := filestorage.NewFile(filepath.Join(src, "_package_"), false, 0o600)
	if err != nil {
		return packId, err
	}
	packId = uuid.UUID{}

	_, err = packageFile.ReadAt(packId[:], 0)
	if err != nil {
		return packId, err
	}
	packageFile.Close()

	if _, ok := Packs[packId]; ok {
		return packId, nil
	}

	if err = ManageKeyPack(packId, false); err != nil {
		return packId, err
	}
	err = Packs[packId].ImportShared(src)
	return packId, err
}

func TryDecodePackId(idKeyPos int64, encId []byte) (id uuid.UUID, ok bool) {
	idLen := int64(len(encId))
	tmpEncId := make([]byte, idLen)
	key := make([]byte, idLen)

	for id := range Packs {
		copy(tmpEncId, encId)

		_, err := Packs[id].InKeys.IdKey.ReadAt(key, idKeyPos)

		if err == nil && quartzSymmetric.Decode(tmpEncId, key) == nil && utils.AreSlicesEqual(tmpEncId, id[:]) {
			return id, true
		}
	}

	return id, false
}
