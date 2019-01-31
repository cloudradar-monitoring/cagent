package smart

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSmartctlVersionValid(t *testing.T) {
	ver := "smartctl 7.0 2018-12-30 r4883 [Darwin 18.2.0 x86_64] (local build)"

	err := smartctlIsSupportedVersion(ver)
	assert.NoError(t, err)
}

func TestSmartctlVersionInvalid1(t *testing.T) {
	ver := ""

	err := smartctlIsSupportedVersion(ver)
	assert.Error(t, err)
}

func TestSmartctlVersionInvalid2(t *testing.T) {
	ver := "smartctl 6.5 2018-12-30 r4883 [Darwin 18.2.0 x86_64] (local build)"

	err := smartctlIsSupportedVersion(ver)
	assert.Error(t, err)
	assert.EqualError(t, errors.New("smart: unsupported smartctl version. expected minimum [7.0], actual [6.5]"), err.Error())
}
