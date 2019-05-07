package smart

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSmartctlVersionValid(t *testing.T) {
	ver := "smartctl 7.0 2018-12-30 r4883 [Darwin 18.2.0 x86_64] (local build)"

	parsedVer, err := smartctlIsSupportedVersion(ver)
	assert.NoError(t, err)
	assert.Equal(t, "7.0", parsedVer)
}

func TestSmartctlVersionInvalid1(t *testing.T) {
	ver := ""

	parsedVer, err := smartctlIsSupportedVersion(ver)
	assert.Error(t, err)
	assert.Equal(t, "", parsedVer)
}

func TestSmartctlVersionInvalid2(t *testing.T) {
	ver := "smartctl 6.5 2018-12-30 r4883 [Darwin 18.2.0 x86_64] (local build)"

	parsedVer, err := smartctlIsSupportedVersion(ver)
	assert.Error(t, err)
	assert.EqualError(t, errors.New("smart: unsupported smartctl version. expected minimum [7.0], actual [6.5]"), err.Error())
	assert.Equal(t, "", parsedVer)
}
