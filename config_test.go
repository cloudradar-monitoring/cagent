package cagent

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/troian/toml"
)

func TestNewMinimumConfig(t *testing.T) {
	envURL := "http://foo.bar"
	envUser := "foo"
	envPass := "bar"

	os.Setenv("CAGENT_HUB_URL", envURL)
	os.Setenv("CAGENT_HUB_USER", envUser)
	os.Setenv("CAGENT_HUB_PASSWORD", envPass)

	// Unset in the end for cleanup
	defer os.Unsetenv("CAGENT_HUB_URL")
	defer os.Unsetenv("CAGENT_HUB_USER")
	defer os.Unsetenv("CAGENT_HUB_PASSWORD")

	mvc := NewMinimumConfig()

	assert.Equal(t, envURL, mvc.HubURL, "HubURL should be set from env")
	assert.Equal(t, envUser, mvc.HubUser, "HubUser should be set from env")
	assert.Equal(t, envPass, mvc.HubPassword, "HubPassword should be set from env")
	assert.Equal(t, LogLevelError, mvc.LogLevel, "default log level should be error")
}

func TestTryUpdateConfigFromFile(t *testing.T) {
	config := Config{
		PidFile:           "fooPID",
		Interval:          1.5,
		HeartbeatInterval: 12.5,
		HubGzip:           false,
		FSMetrics:         []string{"a"},
	}

	const sampleConfig = `
pid = "/pid"
interval = 1.0
heartbeat = 10.0
hub_gzip = true
fs_metrics = ['a', 'b']
`

	tmpFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	err = ioutil.WriteFile(tmpFile.Name(), []byte(sampleConfig), 0755)
	assert.Nil(t, err)

	err = TryUpdateConfigFromFile(&config, tmpFile.Name())
	assert.Nil(t, err)

	assert.Equal(t, "/pid", config.PidFile)
	assert.Equal(t, 1.0, config.Interval)
	assert.Equal(t, 10.0, config.HeartbeatInterval)
	assert.Equal(t, true, config.HubGzip)
	assert.Equal(t, []string{"a", "b"}, config.FSMetrics)
}

func TestGenerateDefaultConfigFile(t *testing.T) {
	mvc := &MinValuableConfig{
		LogLevel: "debug",
		HubUser:  "bar",
	}

	tmpFile, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	err = GenerateDefaultConfigFile(mvc, tmpFile.Name())
	assert.Nil(t, err)

	loadedMVC := &MinValuableConfig{}
	_, err = toml.DecodeReader(tmpFile, loadedMVC)
	assert.Nil(t, err)

	if !assert.ObjectsAreEqual(*mvc, *loadedMVC) {
		t.Errorf("expected %+v, got %+v", *mvc, *loadedMVC)
	}
}

func TestHandleAllConfigSetup(t *testing.T) {
	t.Run("config-file-does-exist", func(t *testing.T) {
		const sampleConfig = `
pid = "/pid"
interval = 100.0
heartbeat = 10.0
hub_gzip = false
fs_metrics = ['a', 'b']
`

		tmpFile, err := ioutil.TempFile("", "")
		assert.Nil(t, err)
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(sampleConfig), 0755)
		assert.Nil(t, err)

		config, err := HandleAllConfigSetup(tmpFile.Name())
		assert.Nil(t, err)

		assert.Equal(t, "/pid", config.PidFile)
		assert.Equal(t, 100.0, config.Interval)
		assert.Equal(t, 10.0, config.HeartbeatInterval)
		assert.Equal(t, false, config.HubGzip)
		assert.Equal(t, []string{"a", "b"}, config.FSMetrics)
	})

	t.Run("config-file-does-not-exist", func(t *testing.T) {
		// Create a temp file to get a file path we can use for temp
		// config generation. But delete it so we can actually write our
		// config file under the path.
		tmpFile, err := ioutil.TempFile("", "")
		assert.Nil(t, err)
		configFilePath := tmpFile.Name()
		err = os.Remove(tmpFile.Name())
		assert.Nil(t, err)

		_, err = HandleAllConfigSetup(configFilePath)
		assert.Nil(t, err)

		_, err = os.Stat(configFilePath)
		assert.Nil(t, err)

		mvc := NewMinimumConfig()
		loadedMVC := &MinValuableConfig{}
		_, err = toml.DecodeFile(configFilePath, loadedMVC)
		assert.Nil(t, err)

		if !assert.ObjectsAreEqual(*mvc, *loadedMVC) {
			t.Errorf("expected %+v, got %+v", *mvc, *loadedMVC)
		}
	})

	t.Run("invalid-interval-value-specified", func(t *testing.T) {
		const sampleConfig = `
pid = "/pid"
interval = 29.9
heartbeat = 15.0
hub_gzip = false
fs_metrics = ['a', 'b']
`

		tmpFile, err := ioutil.TempFile("", "")
		assert.Nil(t, err)
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(sampleConfig), 0755)
		assert.Nil(t, err)

		_, err = HandleAllConfigSetup(tmpFile.Name())
		assert.Error(t, err)

	})

	t.Run("invalid-heartbeat-value-specified", func(t *testing.T) {
		const sampleConfig = `
pid = "/pid"
interval = 50.0
heartbeat = 0.0
hub_gzip = false
fs_metrics = ['a', 'b']
`

		tmpFile, err := ioutil.TempFile("", "")
		assert.Nil(t, err)
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(sampleConfig), 0755)
		assert.Nil(t, err)

		_, err = HandleAllConfigSetup(tmpFile.Name())
		assert.Error(t, err)

	})
}

func TestVirtualNetworkInterfacesExcludedByDefault(t *testing.T) {
	cfg := NewConfig()

	toBeExcluded := []string{"vnet", "vnet0", "vnet2", "virbr", "virbr0", "virbr11", "vmnet", "vmnet1"}
	for _, ifName := range toBeExcluded {
		excluded := false
		for _, reString := range cfg.NetInterfaceExcludeRegex {
			re, err := regexp.Compile(reString)
			assert.NoError(t, err)
			if re.Match([]byte(ifName)) {
				excluded = true
				break
			}
		}
		assert.True(t, excluded, ifName+" excluded")
	}

	notToBeExcluded := []string{"docker", "docker0", "eth0", "eno0", "eth", "eno", "br-426c8e07f670", "wlo1", "lo"}
	for _, ifName := range notToBeExcluded {
		excluded := false
		for _, reString := range cfg.NetInterfaceExcludeRegex {
			re, err := regexp.Compile(reString)
			assert.NoError(t, err)
			if re.Match([]byte(ifName)) {
				excluded = true
				break
			}
		}
		assert.False(t, excluded, ifName+" excluded")
	}
}

func TestGetParsedNetworkInterfaceMaxSpeed(t *testing.T) {
	cfg := NewConfig()

	t.Run("default-value-empty", func(t *testing.T) {
		val, err := cfg.GetParsedNetInterfaceMaxSpeed()
		assert.NoError(t, err)
		assert.Zero(t, val)
	})

	type expectedOut struct {
		val   uint64
		isErr bool
	}
	var m = map[string]expectedOut{
		"": {0, false},

		"0":    {0, true},
		"0.0":  {0, true},
		"0M":   {0, true},
		"0.0M": {0, true},
		"1":    {0, true},
		"1.1":  {0, true},
		"1D":   {0, true},
		"1m":   {0, true},
		"1.1D": {0, true},
		"M":    {0, true},
		"D":    {0, true},
		"MM":   {0, true},
		"10MM": {0, true},
		"-5M":  {0, true},

		"1K":    {1000, false},
		"1M":    {1000 * 1000, false},
		"1G":    {1000 * 1000 * 1000, false},
		"125M":  {1000 * 1000 * 125, false},
		"12.5M": {1000 * 1000 * 12.5, false},
		"25G":   {1000 * 1000 * 1000 * 25, false},
	}

	for strVal, expected := range m {
		cfg.NetInterfaceMaxSpeed = strVal
		val, err := cfg.GetParsedNetInterfaceMaxSpeed()
		if expected.isErr {
			assert.Error(t, err, "for input %s. %v", strVal, err)
		} else {
			assert.NoError(t, err, "for input %s. %v", strVal, err)
		}

		assert.Equal(t, expected.val, val, "for input %s: %d vs %d", strVal, expected.val, val)
	}
}
