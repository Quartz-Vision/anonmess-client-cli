package keysstorage

import (
	quartzSymmetric "quartzvision/anonmess-client-cli/crypto/symmetric"
)

var Keys = map[KeyId]*KeyBuffer{}

func ManageKey(keyId KeyId) (err error) {
	if _, ok := Keys[keyId]; ok {
		return
	}
	Keys[keyId], err = NewKeyBuffer(keyId)
	return err
}

func UnmanageKey(keyId KeyId) {
	if buf, ok := Keys[keyId]; ok {
		buf.Close()
		delete(Keys, keyId)
	}
}

func DecodeKeyID(keyPos int64, encId []byte) (id KeyId, ok bool) {
	idLen := int64(len(encId))

	for id, buf := range Keys {
		tmpEncId := make([]byte, idLen)
		copy(tmpEncId, encId)

		key, err := buf.GetKeySlice(keyPos, idLen)

		if err == nil && quartzSymmetric.Decode(tmpEncId, key) == nil && KeyId(tmpEncId) == id {
			return id, true
		}
	}

	return id, false
}
