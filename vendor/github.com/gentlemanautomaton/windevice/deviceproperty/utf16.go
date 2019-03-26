package deviceproperty

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

func utf16BytesToString(s []byte) string {
	p := (*[0xffff]uint16)(unsafe.Pointer(&s[0]))
	return syscall.UTF16ToString(p[:len(s)/2])
}

func utf16BytesToSplitString(s []byte) []string {
	p := (*[0xffff]uint16)(unsafe.Pointer(&s[0]))
	return utf16ToSplitString(p[:len(s)/2])
}

// utf16ToSplitString splits a set of null-separated utf16 characters and
// returns a slice of substrings between those separators.
func utf16ToSplitString(s []uint16) []string {
	var values []string
	cut := 0
	for i, v := range s {
		if v == 0 {
			if i-cut > 0 {
				values = append(values, string(utf16.Decode(s[cut:i])))
			}
			cut = i + 1
		}
	}
	if cut < len(s) {
		values = append(values, string(utf16.Decode(s[cut:])))
	}
	return values
}
