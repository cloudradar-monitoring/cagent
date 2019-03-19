package winguid

import "golang.org/x/sys/windows"

// String returns a string representation of guid with this pattern:
//
//  {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
func String(guid windows.GUID) string {
	var c [38]byte
	c[0] = '{'
	putUint32Hex(c[1:9], guid.Data1)
	c[9] = '-'
	putUint16Hex(c[10:14], guid.Data2)
	c[14] = '-'
	putUint16Hex(c[15:19], guid.Data3)
	c[19] = '-'
	putByteHex(c[20:24], guid.Data4[0:2])
	c[24] = '-'
	putByteHex(c[25:37], guid.Data4[2:8])
	c[37] = '}'
	return string(c[:])
}
