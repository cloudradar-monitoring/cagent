package devicecreation

// Format maps flags to their string representations.
type Format map[Flags]string

// FormatGo maps flags to Go-style constant strings.
var FormatGo = Format{
	GenerateID:          "GenerateID",
	InheritClassDrivers: "InheritClassDrivers",
}
