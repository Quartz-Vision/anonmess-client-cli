package keystorage

import (
	"quartzvision/anonmess-client-cli/utils"

	"github.com/Quartz-Vision/gofile"

	"github.com/google/uuid"
)

// KeyPosition takes care for accessing and storing current key position, that can be used for encoding/decoding

type KeyPosition struct {
	file gofile.File
}

func NewKeyPosition(packId uuid.UUID, keyPrefix string, packPrefix string) (pos *KeyPosition, err error) {
	var file gofile.File

	return pos, utils.UntilErrorPointer(
		&err,
		func() {
			file, err = gofile.NewFile(keyPath(packId, packPrefix, keyPrefix, PACK_PREFIX_POS), 0o600)
		},
		func() { _, err = file.Write(utils.Int64ToBytes(0)) },
		func() {
			pos = &KeyPosition{
				file: file,
			}
		},
	)
}

// Takes the size of data you want to de/encode and returns the position,
// from which you can use the reserved key part
func (p *KeyPosition) Take(dataSize int64) (pos int64, err error) {
	return pos, p.file.TReadWrite(func(txn gofile.Editable) (err error) {
		data := make([]byte, utils.INT_MAX_SIZE)
		_, err = txn.ReadAt(data, 0)
		if err != nil {
			return err
		}
		pos, _ = utils.BytesToInt64(data)
		nextPos := pos + dataSize
		_, err = txn.WriteAt(utils.Int64ToBytes(nextPos), 0)
		return err
	})
}
