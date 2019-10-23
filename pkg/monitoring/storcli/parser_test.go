package storcli

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helperLoadAndParseTestData(t *testing.T, fileName string) *controllersResult {
	path := filepath.Join("testdata", fileName)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := tryParseCmdOutput(&b)
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Controllers)

	return result
}

func TestTryParseCmdOutput(t *testing.T) {
	t.Run("all-good-output", func(t *testing.T) {
		output := helperLoadAndParseTestData(t, "output_allgood.json")
		measurements, alerts, warnings, err := getReportData(&output.Controllers[0].ResponseData)
		assert.NoError(t, err)
		assert.Empty(t, alerts)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
	})

	t.Run("all-good-output-ubuntu", func(t *testing.T) {
		output := helperLoadAndParseTestData(t, "output_allgood_ubuntu.json")
		measurements, alerts, warnings, err := getReportData(&output.Controllers[0].ResponseData)
		assert.NoError(t, err)
		assert.Empty(t, alerts)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
	})

	t.Run("non-optimal-output", func(t *testing.T) {
		output := helperLoadAndParseTestData(t, "output_nonoptimal.json")
		measurements, alerts, warnings, err := getReportData(&output.Controllers[0].ResponseData)
		assert.NoError(t, err)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
		assert.NotEmpty(t, alerts)
	})

	t.Run("virtual-drive-bad-output", func(t *testing.T) {
		output := helperLoadAndParseTestData(t, "output_vdbad.json")
		measurements, alerts, warnings, err := getReportData(&output.Controllers[0].ResponseData)
		assert.NoError(t, err)
		assert.Empty(t, warnings)
		assert.NotNil(t, measurements["Status"])
		assert.NotEmpty(t, alerts)
	})

	t.Run("hard-drive-bad-output", func(t *testing.T) {
		output := helperLoadAndParseTestData(t, "output_hdbad.json")
		measurements, alerts, warnings, err := getReportData(&output.Controllers[0].ResponseData)
		assert.NoError(t, err)
		assert.Empty(t, alerts)
		assert.NotEmpty(t, warnings)
		assert.NotNil(t, measurements["Status"])
	})
}
