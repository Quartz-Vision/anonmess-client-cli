package utils

import (
	"unsafe"
)

func XorSlices(a []byte, b []byte) {
	dataLen := uint(len(a))
	a = a[:len(b)]
	b = b[:dataLen]
	chunksCount := dataLen >> 3

	if chunksCount != 0 {
		aPointer := unsafe.Pointer(&a[0])
		bPointer := unsafe.Pointer(&b[0])

		for i := uint(0); i < chunksCount; i++ {
			*(*uint64)(unsafe.Add(aPointer, i<<3)) ^= *(*uint64)(unsafe.Add(bPointer, i<<3))
		}
	}

	for i := chunksCount << 3; i < dataLen; i++ {
		a[i] ^= b[i]
	}
}

func ProcessSlices(dest []byte, op func(a []byte, b []byte), a []byte, b []byte) (res []byte) {
	copy(dest, a)
	op(dest, b)
	return dest
}

// Doesn't check the lengths (takes the smaller)
func AreSlicesEqual(a []byte, b []byte) (ok bool) {
	dataLen := uint(len(a))
	a = a[:len(b)]
	b = b[:dataLen]
	chunksCount := dataLen >> 3

	if chunksCount != 0 {
		aPointer := unsafe.Pointer(&a[0])
		bPointer := unsafe.Pointer(&b[0])

		for i := uint(0); i < chunksCount; i++ {
			if *(*uint64)(unsafe.Add(aPointer, i<<3)) != *(*uint64)(unsafe.Add(bPointer, i<<3)) {
				return false
			}
		}
	}

	for i := chunksCount << 3; i < dataLen; i++ {
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
