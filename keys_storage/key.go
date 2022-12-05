package keysstorage

import (
	"path/filepath"
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type Key struct {
	*KeyBuffer
	prefix            string
	packPrefix        string
	packSharingPrefix string
}

func NewKey(packId uuid.UUID, keyPrefix string, packPrefix string, packSharingPrefix string) (key *Key, err error) {
	kb, err := NewKeyBuffer(keyPath(packId, packPrefix, keyPrefix))
	if err != nil {
		return nil, err
	}

	return &Key{
		KeyBuffer:         kb,
		prefix:            keyPrefix,
		packPrefix:        packPrefix,
		packSharingPrefix: packSharingPrefix,
	}, nil
}

func (k *Key) ExportShared(dest string) (err error) {
	var file filestorage.File

	return utils.UntilErrorPointer(
		&err,
		func() {
			file, err = filestorage.NewFile(filepath.Join(dest, keyFileName(k.packSharingPrefix, k.prefix)), 0o600)
		},
		func() { err = file.Trunc() },
		func() { err = k.PipeTo(file, settings.Config.KeysBufferSizeB) },
		func() { file.Close() },
	)
}
func (k *Key) ImportShared(src string) (err error) {
	var file filestorage.File

	return utils.UntilErrorPointer(
		&err,
		func() {
			file, err = filestorage.NewFile(filepath.Join(src, keyFileName(k.packPrefix, k.prefix)), 0o600)
		},
		func() { err = file.PipeTo(k, settings.Config.KeysBufferSizeB) },
		func() { file.Close() },
	)
}
