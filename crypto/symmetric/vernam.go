package symmetric

func Encode(data []byte, key []byte) (err error) {
	if len(data) > len(key) {
		return ErrWrongKeyLength
	}

	xorSlices(data, key)
	return nil
}

func Decode(data []byte, key []byte) (err error) {
	if len(data) > len(key) {
		return ErrWrongKeyLength
	}

	xorSlices(data, key)
	return nil
}
