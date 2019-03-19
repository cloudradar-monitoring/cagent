package setupapi

import (
	"encoding/binary"
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

func utf16BytesFromStrings(s []string) ([]byte, error) {
	var lines [][]uint16
	var length int

	// Build a set of utf16 lines
	for i := range s {
		line, err := syscall.UTF16FromString(s[i])
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
		length += len(line)
	}

	// Terminate with an empty line
	{
		line, err := syscall.UTF16FromString("")
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
		length += len(line)
	}

	// Make a buffer of appropriate length
	buffer := make([]byte, length*2)

	// Copy little-endian bytes into the buffer
	var pos int
	for _, line := range lines {
		for _, char := range line {
			var le [2]byte
			binary.LittleEndian.PutUint16(le[:], char)
			buffer[pos] = le[0]
			buffer[pos+1] = le[1]
			pos += 2
		}
	}

	return buffer, nil
}

func utf16BytesFromString(s string) ([]byte, error) {
	// Convert to utf16
	line, err := syscall.UTF16FromString(s)
	if err != nil {
		return nil, err
	}

	// Make a buffer of appropriate length
	buffer := make([]byte, len(line)*2)

	// Copy little-endian bytes into the buffer
	var pos int
	for _, char := range line {
		var le [2]byte
		binary.LittleEndian.PutUint16(le[:], char)
		buffer[pos] = le[0]
		buffer[pos+1] = le[1]
		pos += 2
	}

	return buffer, nil
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
