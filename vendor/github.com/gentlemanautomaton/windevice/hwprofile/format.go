package hwprofile

// Format maps hardware profile scope flags to their string representations.
type Format map[Scope]string

// FormatGo maps hardware profile scope flags to Go-style constant strings.
var FormatGo = Format{
	Global:         "Global",
	ConfigSpecific: "ConfigSpecific",
	ConfigGeneral:  "ConfigGeneral",
}
