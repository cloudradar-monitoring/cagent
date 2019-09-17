package storcli

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helperLoadTestData(t *testing.T, fileName string) []byte {
	path := filepath.Join("testdata", fileName)
	result, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestTryParseCmdOutput(t *testing.T) {
	t.Run("all-good-output", func(t *testing.T) {
		output := helperLoadTestData(t, "output_allgood.json")
		measurements, alerts, warnings, err := tryParseCmdOutput(&output)
		assert.NoError(t, err)
		assert.Empty(t, alerts)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
	})

	t.Run("all-good-output-ubuntu", func(t *testing.T) {
		output := helperLoadTestData(t, "output_allgood_ubuntu.json")
		measurements, alerts, warnings, err := tryParseCmdOutput(&output)
		assert.NoError(t, err)
		assert.Empty(t, alerts)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
	})

	t.Run("non-optimal-output", func(t *testing.T) {
		output := helperLoadTestData(t, "output_nonoptimal.json")
		measurements, alerts, warnings, err := tryParseCmdOutput(&output)
		assert.NoError(t, err)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
		assert.NotEmpty(t, alerts)
	})

	t.Run("virtual-drive-bad-output", func(t *testing.T) {
		output := helperLoadTestData(t, "output_vdbad.json")
		measurements, alerts, warnings, err := tryParseCmdOutput(&output)
		assert.NoError(t, err)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
		assert.NotEmpty(t, alerts)
	})

	t.Run("hard-drive-bad-output", func(t *testing.T) {
		output := helperLoadTestData(t, "output_hdbad.json")
		measurements, alerts, warnings, err := tryParseCmdOutput(&output)
		assert.NoError(t, err)
		assert.Empty(t, alerts)
		assert.NotEmpty(t, warnings)
		assert.NotNil(t, measurements["Status"])
	})
}
