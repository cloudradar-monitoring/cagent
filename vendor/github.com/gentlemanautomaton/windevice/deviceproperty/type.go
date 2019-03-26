package deviceproperty

// Device property type masks.
const (
	BaseTypeMask     = 0x00000FFF // DEVPROP_MASK_TYPE
	TypeModifierMask = 0x0000F000 // DEVPROP_MASK_TYPEMOD
)

// Property type modifiers.
const (
	Array Type = 0x00001000 // DEVPROP_TYPEMOD_ARRAY
	List  Type = 0x00002000 // DEVPROP_TYPEMOD_LIST
)

// Device property base types.
const (
	Empty                    Type = 0x00000000 // DEVPROP_TYPE_EMPTY
	Null                     Type = 0x00000001 // DEVPROP_TYPE_NULL
	Int8                     Type = 0x00000002 // DEVPROP_TYPE_SBYTE
	Byte                     Type = 0x00000003 // DEVPROP_TYPE_BYTE
	Int16                    Type = 0x00000004 // DEVPROP_TYPE_INT16
	Uint16                   Type = 0x00000005 // DEVPROP_TYPE_UINT16
	Int32                    Type = 0x00000006 // DEVPROP_TYPE_INT32
	Uint32                   Type = 0x00000007 // DEVPROP_TYPE_UINT32
	Int64                    Type = 0x00000008 // DEVPROP_TYPE_INT64
	Uint64                   Type = 0x00000009 // DEVPROP_TYPE_UINT64
	Float                    Type = 0x0000000A // DEVPROP_TYPE_FLOAT
	Double                   Type = 0x0000000B // DEVPROP_TYPE_DOUBLE
	Decimal                  Type = 0x0000000C // DEVPROP_TYPE_DECIMAL
	GUID                     Type = 0x0000000D // DEVPROP_TYPE_GUID
	Currency                 Type = 0x0000000E // DEVPROP_TYPE_CURRENCY
	Date                     Type = 0x0000000F // DEVPROP_TYPE_DATE
	FileTime                 Type = 0x00000010 // DEVPROP_TYPE_FILETIME
	Bool                     Type = 0x00000011 // DEVPROP_TYPE_BOOLEAN
	String                   Type = 0x00000012 // DEVPROP_TYPE_STRING
	SecurityDescriptor       Type = 0x00000013 // DEVPROP_TYPE_SECURITY_DESCRIPTOR
	SecurityDescriptorString Type = 0x00000014 // DEVPROP_TYPE_SECURITY_DESCRIPTOR_STRING
	DevicePropertyKey        Type = 0x00000015 // DEVPROP_TYPE_DEVPROPKEY
	DevicePropertyType       Type = 0x00000016 // DEVPROP_TYPE_DEVPROPTYPE
	Error                    Type = 0x00000017 // DEVPROP_TYPE_ERROR
	Status                   Type = 0x00000018 // DEVPROP_TYPE_NTSTATUS
	StringIndirect           Type = 0x00000019 // DEVPROP_TYPE_STRING_INDIRECT
)

// Common property types.
const (
	Binary     Type = Byte | Array  // DEVPROP_TYPE_BINARY
	StringList Type = String | List // DEVPROP_TYPE_STRING_LIST
)

// Type is a device property type.
type Type uint32

// Base returns the base type of t.
func (t Type) Base() Type {
	return t & BaseTypeMask
}

// Modifier returns the type modifier of t.
func (t Type) Modifier() Type {
	return t & TypeModifierMask
}

// DataLength returns the length of data required to store a value of type t.
func (t Type) DataLength() int {
	return -1
}
