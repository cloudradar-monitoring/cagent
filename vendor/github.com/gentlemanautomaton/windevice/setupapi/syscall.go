package setupapi

import "golang.org/x/sys/windows"

var modsetupapi = windows.NewLazySystemDLL("setupapi.dll")
