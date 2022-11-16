package sliceutils

import "unsafe"

func Xor(a []byte, b []byte) {
	data_len := uint(len(a))
	chunks_count := data_len >> 3

	if chunks_count != 0 {
		aPointer := unsafe.Pointer(&a[0])
		bPointer := unsafe.Pointer(&b[0])

		for i := uint(0); i < chunks_count; i++ {
			*(*int64)(unsafe.Add(aPointer, i<<3)) ^= *(*int64)(unsafe.Add(bPointer, i<<3))
		}
	}

	for i := chunks_count << 3; i < data_len; i++ {
		a[i] ^= b[i]
	}
}

func IsEqual(a []byte, b []byte) (ok bool) {
	data_len := uint(len(a))
	chunks_count := data_len >> 3

	if chunks_count != 0 {
		aPointer := unsafe.Pointer(&a[0])
		bPointer := unsafe.Pointer(&b[0])

		for i := uint(0); i < chunks_count; i++ {
			if *(*int64)(unsafe.Add(aPointer, i<<3)) != *(*int64)(unsafe.Add(bPointer, i<<3)) {
				return false
			}
		}
	}

	for i := chunks_count << 3; i < data_len; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
