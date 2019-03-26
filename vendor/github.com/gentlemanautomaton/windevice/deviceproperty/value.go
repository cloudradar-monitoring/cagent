package deviceproperty

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gentlemanautomaton/winguid"
	"golang.org/x/sys/windows"
)

// Value is a device property value. It can safely be copied by value.
type Value struct {
	t      Type
	length uint32   // number of bytes in array
	array  [20]byte // data array for small values
	slice  []byte   // data slice for large values
}

// NewValue returns a device property value for the given data.
func NewValue(t Type, data []byte) Value {
	v := Value{t: t}
	length := len(data)
	switch {
	case length <= len(v.array):
		// Copy data to the inline byte array and record its length
		v.length = uint32(length)
		copy(v.array[:length], data)
	case length != cap(data):
		// Allocate a slice with the correct capacity for the data
		v.slice = append(data[:0:0], data[:length]...)
	default:
		// Store the data as is
		v.slice = data
	}
	return v
}

// Type returns the type of the value.
func (v Value) Type() Type {
	return v.t
}

// Bytes returns the bytes of the value.
func (v Value) Bytes() []byte {
	if v.slice == nil {
		return v.array[:v.length]
	}
	return v.slice
}

// String returns a string representation of the value.
func (v Value) String() string {
	switch v.t.Base() {
	case Empty:
		return ""
	case Null:
		return ""
	case Int8:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Int8()))
		case Array:
			return fmt.Sprintf("%v", v.Int8List())
		}
	case Byte:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Byte()))
		case Array:
			return fmt.Sprintf("%#x", v.Bytes())
		}
	case Int16:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Int16()))
		case Array:
			return fmt.Sprintf("%v", v.Int16List())
		}
	case Uint16:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Uint16()))
		case Array:
			return fmt.Sprintf("%v", v.Uint16List())
		}
	case Int32:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Int32()))
		case Array:
			return fmt.Sprintf("%v", v.Int32List())
		}
	case Uint32:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Uint32()))
		case Array:
			return fmt.Sprintf("%v", v.Uint32List())
		}
	case Int64:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Int64()))
		case Array:
			return fmt.Sprintf("%v", v.Int64List())
		}
	case Uint64:
		switch v.t.Modifier() {
		case 0:
			return strconv.Itoa(int(v.Uint64()))
		case Array:
			return fmt.Sprintf("%v", v.Uint64List())
		}
	case Float:
		switch v.t.Modifier() {
		case 0:
			return fmt.Sprintf("%v", v.Float32())
		case Array:
			return fmt.Sprintf("%v", v.Float32List())
		}
	case Double:
		switch v.t.Modifier() {
		case 0:
			return fmt.Sprintf("%v", v.Float64())
		case Array:
			return fmt.Sprintf("%v", v.Float64List())
		}
	case Decimal:
	case GUID:
		switch v.t.Modifier() {
		case 0:
			return fmt.Sprintf("%v", winguid.String(v.GUID()))
		case Array:
			guids := v.GUIDList()
			s := make([]string, 0, len(guids))
			for _, guid := range guids {
				s = append(s, winguid.String(guid))
			}
			return fmt.Sprintf("%v", s)
		}
	case Currency:
	case Date:
	case FileTime:
		switch v.t.Modifier() {
		case 0:
			return fmt.Sprintf("%v", v.Time())
		case Array:
			return fmt.Sprintf("%v", v.TimeList())
		}
	case Bool:
		switch v.t.Modifier() {
		case 0:
			return fmt.Sprintf("%v", v.Bool())
		case Array:
			return fmt.Sprintf("%v", v.BoolList())
		}
	case String:
		switch v.t.Modifier() {
		case 0:
			return utf16BytesToString(v.Bytes())
		case List:
			return fmt.Sprintf("%v", utf16BytesToSplitString(v.Bytes()))
		}
	case SecurityDescriptor:
	case SecurityDescriptorString:
	case DevicePropertyKey:
	case DevicePropertyType:
	case Error:
	case Status:
	case StringIndirect:
	}
	return ""
}

// Int8 returns v as an int8.
func (v Value) Int8() int8 {
	return int8(v.array[0])
}

// Byte interprets v as byte.
func (v Value) Byte() byte {
	return v.array[0]
}

// Int16 interprets v as int16.
func (v Value) Int16() int16 {
	return int16(binary.LittleEndian.Uint16(v.array[:2]))
}

// Uint16 interprets v as uint16.
func (v Value) Uint16() uint16 {
	return binary.LittleEndian.Uint16(v.array[:2])
}

// Int32 interprets v as int32.
func (v Value) Int32() int32 {
	return int32(binary.LittleEndian.Uint32(v.array[:4]))
}

// Uint32 interprets v as uint32.
func (v Value) Uint32() uint32 {
	return binary.LittleEndian.Uint32(v.array[:4])
}

// Int64 interprets v as int64.
func (v Value) Int64() int64 {
	return int64(binary.LittleEndian.Uint64(v.array[:8]))
}

// Uint64 interprets v as uint64.
func (v Value) Uint64() uint64 {
	return binary.LittleEndian.Uint64(v.array[:8])
}

// Float32 interprets v as float32.
func (v Value) Float32() float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(v.array[:4]))
}

// Float64 interprets v as float64.
func (v Value) Float64() float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(v.array[:8]))
}

// GUID interprets v as windows.GUID.
func (v Value) GUID() windows.GUID {
	return winguid.NativeEndian.GUID(v.array[:16])
}

// Time interprets v as time.Time.
func (v Value) Time() time.Time {
	filetime := windows.Filetime{
		LowDateTime:  binary.LittleEndian.Uint32(v.array[0:4]),
		HighDateTime: binary.LittleEndian.Uint32(v.array[4:8]),
	}
	return time.Unix(0, filetime.Nanoseconds())
}

// Bool interprets v as bool.
func (v Value) Bool() bool {
	return v.array[0] != 0
}

// Int8List interprets v as []int8.
func (v Value) Int8List() []int8 {
	data := v.Bytes()
	list := make([]int8, len(data))
	for i := range data {
		list[i] = int8(data[i])
	}
	return list
}

// Int16List interprets v as []int16.
func (v Value) Int16List() []int16 {
	data := v.Bytes()
	count := len(data) / 2
	list := make([]int16, count)
	for i := 0; i < count; i++ {
		list[i] = int16(binary.LittleEndian.Uint16(data[i*2:]))
	}
	return list
}

// Uint16List interprets v as []uint16.
func (v Value) Uint16List() []uint16 {
	data := v.Bytes()
	count := len(data) / 2
	list := make([]uint16, count)
	for i := 0; i < count; i++ {
		list[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return list
}

// Int32List interprets v as []int32.
func (v Value) Int32List() []int32 {
	data := v.Bytes()
	count := len(data) / 4
	list := make([]int32, count)
	for i := 0; i < count; i++ {
		list[i] = int32(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return list
}

// Uint32List interprets v as []uint32.
func (v Value) Uint32List() []uint32 {
	data := v.Bytes()
	count := len(data) / 4
	list := make([]uint32, count)
	for i := 0; i < count; i++ {
		list[i] = binary.LittleEndian.Uint32(data[i*4:])
	}
	return list
}

// Int64List interprets v as []int64.
func (v Value) Int64List() []int64 {
	data := v.Bytes()
	count := len(data) / 8
	list := make([]int64, count)
	for i := 0; i < count; i++ {
		list[i] = int64(binary.LittleEndian.Uint64(data[i*8:]))
	}
	return list
}

// Uint64List interprets v as []uint64.
func (v Value) Uint64List() []uint64 {
	data := v.Bytes()
	count := len(data) / 8
	list := make([]uint64, count)
	for i := 0; i < count; i++ {
		list[i] = binary.LittleEndian.Uint64(data[i*8:])
	}
	return list
}

// Float32List interprets v as float32.
func (v Value) Float32List() []float32 {
	data := v.Bytes()
	count := len(data) / 4
	list := make([]float32, count)
	for i := 0; i < count; i++ {
		list[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return list
}

// Float64List interprets v as []float64.
func (v Value) Float64List() []float64 {
	data := v.Bytes()
	count := len(data) / 8
	list := make([]float64, count)
	for i := 0; i < count; i++ {
		list[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[i*8:]))
	}
	return list
}

// GUIDList interprets v as []windows.GUID.
func (v Value) GUIDList() []windows.GUID {
	data := v.Bytes()
	count := len(data) / 16
	list := make([]windows.GUID, count)
	for i := 0; i < count; i++ {
		list[i] = winguid.NativeEndian.GUID(data[i*16:])
	}
	return list
}

// TimeList interprets v as []time.Time.
func (v Value) TimeList() []time.Time {
	data := v.Bytes()
	count := len(data) / 8
	list := make([]time.Time, count)
	for i := 0; i < count; i++ {
		offset := i * 8
		filetime := windows.Filetime{
			LowDateTime:  binary.LittleEndian.Uint32(data[offset : offset+4]),
			HighDateTime: binary.LittleEndian.Uint32(data[offset+4 : offset+8]),
		}
		list[i] = time.Unix(0, filetime.Nanoseconds())
	}
	return list
}

// BoolList interprets v as []bool.
func (v Value) BoolList() []bool {
	data := v.Bytes()
	list := make([]bool, len(data))
	for i := range data {
		list[i] = data[i] != 0
	}
	return list
}
