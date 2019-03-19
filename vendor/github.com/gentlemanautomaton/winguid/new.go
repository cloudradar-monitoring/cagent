package winguid

import "golang.org/x/sys/windows"

const emptyGUID = "{00000000-0000-0000-0000-000000000000}"

// New converts the given string into a windows.GUID struct that is
// compliant with the Windows API.
//
// The supplied string may be in any of these formats:
//
//  XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
//  XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
//  {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
//
// The supplied string is expected to be in hexadecimal notation with all fields
// in big-endian byte order. Note that some systems may supply mixed-endian or
// little-endian hexadecimal representations.
//
// The conversion of the supplied string is not case-sensitive.
//
// If the conversion fails an empty GUID will be returned.
func New(guid string) windows.GUID {
	g, _ := TryNew(guid)
	return g
}

// TryNew performs the same operation as New, but it returns false if the
// conversion fails.
func TryNew(guid string) (g windows.GUID, ok bool) {
	d := []byte(guid)
	var d1, d2, d3, d4a, d4b []byte

	switch len(d) {
	case 38:
		if d[0] != '{' || d[37] != '}' {
			return windows.GUID{}, false
		}
		d = d[1:37]
		fallthrough
	case 36:
		if d[8] != '-' || d[13] != '-' || d[18] != '-' || d[23] != '-' {
			return windows.GUID{}, false
		}
		d1 = d[0:8]
		d2 = d[9:13]
		d3 = d[14:18]
		d4a = d[19:23]
		d4b = d[24:36]
	case 32:
		d1 = d[0:8]
		d2 = d[8:12]
		d3 = d[12:16]
		d4a = d[16:20]
		d4b = d[20:32]
	default:
		return windows.GUID{}, false
	}

	var ok1, ok2, ok3, ok4 bool
	g.Data1, ok1 = decodeHexUint32(d1)
	g.Data2, ok2 = decodeHexUint16(d2)
	g.Data3, ok3 = decodeHexUint16(d3)
	g.Data4, ok4 = decodeHexByte64(d4a, d4b)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return windows.GUID{}, false
	}

	return g, true
}
