package smart

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmartctlVersionValid(t *testing.T) {
	ver := "smartctl 7.0 2018-12-30 r4883 [Darwin 18.2.0 x86_64] (local build)"

	major, minor, err := smartctlParseVersion(ver)
	assert.NoError(t, err)
	assert.Equal(t, atLeastMajorVersion, major)
	assert.Equal(t, 0, minor)
}

func TestSmartctlVersionInvalid1(t *testing.T) {
	ver := ""

	major, minor, err := smartctlParseVersion(ver)
	assert.Error(t, err)
	assert.Equal(t, 0, major)
	assert.Equal(t, 0, minor)
}
