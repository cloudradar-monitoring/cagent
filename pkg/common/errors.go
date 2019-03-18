package common

import (
	"errors"
	"fmt"
)

type ErrorCollector struct {
	errs []error
}

func (c *ErrorCollector) New(err error) {
	c.errs = append(c.errs, err)
}

func (c *ErrorCollector) Add(text string) {
	c.New(errors.New(text))
}

func (c *ErrorCollector) Addf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	c.New(err)
}

func (c *ErrorCollector) HasErrors() bool {
	return len(c.errs) > 0
}

func (c *ErrorCollector) Combine() error {
	if c.HasErrors() {
		return errors.New(c.String())
	}
	return nil
}

func (c *ErrorCollector) String() string {
	result := ""
	for i, err := range c.errs {
		result += err.Error()
		if i != len(c.errs)-1 {
			result += "; "
		}
	}
	return result
}
