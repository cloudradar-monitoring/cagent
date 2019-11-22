package csender

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

const MaxKeyLength = 100
const MaxValueLength = 500
const HubTimeout = 30 * time.Second

var keyBadRE = regexp.MustCompile(`[\s\t]`)

func (cs *Csender) SetSuccess(success bool) error {
	var s string
	if success {
		s = "1"
	} else {
		s = "0"
	}
	return cs.AddKeyValue("success=" + s)
}

func (cs *Csender) SetAlert(alert string) error {
	return cs.AddKeyValue("alert=" + alert)
}

func (cs *Csender) SetWarning(warning string) error {
	return cs.AddKeyValue("warning=" + warning)
}

func (cs *Csender) AddMultipleKeyValue(kv []string) error {
	err := validateKey(cs.CheckName)
	if err != nil {
		return fmt.Errorf("invalid key: %s, got \"%s\"", err.Error(), cs.CheckName)
	}

	if cs.result == nil {
		cs.result = make(common.MeasurementsMap)
	}

	for _, v := range kv {
		err := cs.AddKeyValue(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *Csender) AddKeyValue(kv string) error {
	parts := strings.Split(kv, "=")
	if len(parts) < 2 {
		return fmt.Errorf("failed to parse key=value: %s", kv)
	}
	key := strings.TrimSpace(parts[0])

	// will check the concat'ed key to validate the maximum key size
	if err := validateKey(cs.CheckName + "." + key); err != nil {
		return fmt.Errorf("invalid check name: %s, got \"%s\"", err.Error(), key)
	}

	if cs.result == nil {
		cs.result = make(common.MeasurementsMap)
	}

	if _, exists := cs.result[cs.CheckName+"."+key]; exists {
		return fmt.Errorf("key '%s' duplicated", key)
	}

	value := strings.TrimSpace(parts[1])

	if len(value) > MaxValueLength {
		return fmt.Errorf("invalid value: length is longer than maximum %d", MaxValueLength)
	}

	valueParsed, err := strconv.ParseFloat(value, 64)
	if err == nil {
		cs.result[cs.CheckName+"."+key] = valueParsed
	} else {
		cs.result[cs.CheckName+"."+key] = value
	}

	return nil
}

func validateKey(key string) error {
	if keyBadRE.MatchString(key) {
		return errors.New("can't contain space")
	}

	if len(key) > MaxKeyLength {
		return fmt.Errorf("length is longer than maximum %d", MaxKeyLength)
	}

	if key[0:1] == "." || key[len(key)-1:] == "." {
		return errors.New("starts or ends with a dot")
	}
	if strings.Contains(key, "..") {
		return errors.New("has more than 1 dot in a row")
	}

	return nil
}
