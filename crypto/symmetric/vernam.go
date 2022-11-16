package symmetric

import sliceutils "quartzvision/anonmess-client-cli/slice_utils"

func Encode(data []byte, key []byte) (err error) {
	if len(data) > len(key) {
		return ErrWrongKeyLength
	}

	sliceutils.Xor(data, key)
	return nil
}

func Decode(data []byte, key []byte) (err error) {
	if len(data) > len(key) {
		return ErrWrongKeyLength
	}

	sliceutils.Xor(data, key)
	return nil
}
