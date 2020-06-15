package cagent

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helperCreateCagent(t *testing.T) *Cagent {
	t.Helper()

	tmpFile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)

	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	cfg, err := HandleAllConfigSetup(tmpFilePath)
	assert.NoError(t, err)

	ca, err := New(cfg, tmpFilePath)
	assert.NoError(t, err)

	return ca
}

func TestCagentCollectMeasurements(t *testing.T) {
	ca := helperCreateCagent(t)
	defer ca.Shutdown()

	m, _ := ca.collectMeasurements(true)
	errorMsg, ok := m["message"]
	if !ok {
		errorMsg = ""
	}
	assert.Equal(t, 1, m["cagent.success"], "msg %s", errorMsg)
}
