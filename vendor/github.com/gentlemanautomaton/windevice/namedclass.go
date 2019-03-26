package windevice

import (
	"github.com/gentlemanautomaton/windevice/setupapi"
	"golang.org/x/sys/windows"
)

// NamedClass is a named device class.
type NamedClass struct {
	Name    string
	Members []windows.GUID
}

// NewNamedClass returns a NamedClass entry for the given class name.
func NewNamedClass(name string) (NamedClass, error) {
	members, err := setupapi.ClassGuidsFromNameEx(name, "")
	if err != nil {
		return NamedClass{}, err
	}
	return NamedClass{
		Name:    name,
		Members: members,
	}, nil
}
