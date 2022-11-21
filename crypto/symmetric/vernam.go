package symmetric

import utils "quartzvision/anonmess-client-cli/utils"

func Encode(data []byte, key []byte) (err error) {
	if len(data) > len(key) {
		return ErrWrongKeyLength
	}

	utils.XorSlices(data, key)
	return nil
}

func Decode(data []byte, key []byte) (err error) {
	if len(data) > len(key) {
		return ErrWrongKeyLength
	}

	utils.XorSlices(data, key)
	return nil
}
