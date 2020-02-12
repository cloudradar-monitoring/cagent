// +build !windows

package updates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryParseMajorVersion(t *testing.T) {
	const invalidVal = -1
	var testMap = map[string]int{
		"":                              invalidVal,
		"-1":                            invalidVal,
		"16.10":                         16,
		"v16.10":                        16,
		"v.13":                          invalidVal,
		"16.10-rc23":                    16,
		"12321312321312313212313123123": invalidVal,
		"rawhide":                       invalidVal,
		"5.4.3-arch1-1":                 5,
	}

	for input, expected := range testMap {
		assert.Equal(t, expected, tryParseMajorVersion(input))
	}
}
