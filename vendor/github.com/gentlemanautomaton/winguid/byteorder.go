package winguid

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// ByteOrder defines a byte order when converting to and from bytes.
type ByteOrder interface {
	GUID([]byte) windows.GUID
	//PutGUID([]byte, windows.GUID)
}

// BigEndian is the big-endian implementation of ByteOrder.
var BigEndian bigEndian

type bigEndian struct{}

func (bigEndian) GUID(b []byte) windows.GUID {
	return windows.GUID{
		Data1: makeUint32(b[0], b[1], b[2], b[3]),
		Data2: makeUint16(b[4], b[5]),
		Data3: makeUint16(b[6], b[7]),
		Data4: makeByte64(b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]),
	}
}

func (bigEndian) PutGUID([]byte, windows.GUID) {
	// TODO: Write this
}

// LittleEndian is the little-endian implementation of ByteOrder.
var LittleEndian littleEndian

type littleEndian struct{}

func (littleEndian) GUID(b []byte) windows.GUID {
	return windows.GUID{
		Data1: makeUint32(b[3], b[2], b[1], b[0]),
		Data2: makeUint16(b[5], b[4]),
		Data3: makeUint16(b[7], b[6]),
		Data4: makeByte64(b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]),
	}
}

func (littleEndian) PutGUID([]byte, windows.GUID) {
	// TODO: Write this
}

// NativeEndian is the implementation of ByteOrder that always matches the
// endianness of the local system.
var NativeEndian nativeEndian

type nativeEndian struct{}

func (nativeEndian) GUID(b []byte) windows.GUID {
	return windows.GUID{
		Data1: *(*uint32)(unsafe.Pointer(&b[0])),
		Data2: *(*uint16)(unsafe.Pointer(&b[4])),
		Data3: *(*uint16)(unsafe.Pointer(&b[6])),
		Data4: makeByte64(b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]),
	}
}

func (nativeEndian) PutGUID(b []byte, guid windows.GUID) {
	// TODO: Write this
}

func makeUint32(b1, b2, b3, b4 byte) uint32 {
	return uint32(b1)<<24 | uint32(b2)<<16 | uint32(b3)<<8 | uint32(b4)
}

func makeUint16(b1, b2 byte) uint16 {
	return uint16(b1)<<8 | uint16(b2)
}

func makeByte64(b1, b2, b3, b4, b5, b6, b7, b8 byte) [8]byte {
	return [8]byte{b1, b2, b3, b4, b5, b6, b7, b8}
}
