package utils

import "unsafe"

func XorSlices(a []byte, b []byte) {
	data_len := uint(len(a))
	chunks_count := data_len >> 3

	if chunks_count != 0 {
		aPointer := unsafe.Pointer(&a[0])
		bPointer := unsafe.Pointer(&b[0])

		for i := uint(0); i < chunks_count; i++ {
			*(*uint64)(unsafe.Add(aPointer, i<<3)) ^= *(*uint64)(unsafe.Add(bPointer, i<<3))
		}
	}

	for i := chunks_count << 3; i < data_len; i++ {
		a[i] ^= b[i]
	}
}

func AreSlicesEqual(a []byte, b []byte) (ok bool) {
	data_len := uint(len(a))
	chunks_count := data_len >> 3

	if chunks_count != 0 {
		aPointer := unsafe.Pointer(&a[0])
		bPointer := unsafe.Pointer(&b[0])

		for i := uint(0); i < chunks_count; i++ {
			if *(*uint64)(unsafe.Add(aPointer, i<<3)) != *(*uint64)(unsafe.Add(bPointer, i<<3)) {
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

// Joins all slices and puts them into dst.
// Returns position right after the end of the last slice put
func JoinSlices(dst []byte, slices ...[]byte) (pos int) {
	pos = 0
	for i := range slices {
		pos += copy(dst[pos:], slices[i])
	}

	return pos
}
