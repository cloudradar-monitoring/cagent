package installflag

// Format maps flags to their string representations.
type Format map[Value]string

// FormatGo maps values to Go-style constant strings.
var FormatGo = Format{
	Force:          "Force",
	ReadOnly:       "ReadOnly",
	NonInteractive: "NonInteractive",
}
