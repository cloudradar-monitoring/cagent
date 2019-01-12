// +build !windows

package main

import (
	"github.com/cloudradar-monitoring/cagent"
)

// this dumb func exists only for cross-platform compiling, because it was mentioned in the main.go(which is compiling for all platforms)
func windowsShowUI(ca *cagent.Cagent) {

}
