package common

import (
	"errors"
	"fmt"
)

type ErrorCollector struct {
	errs []error
}

// Add adds error to collection
func (c *ErrorCollector) Add(err error) {
	if err != nil {
		c.errs = append(c.errs, err)
	}
}

// AddNew creates new error and adds to collection
func (c *ErrorCollector) AddNew(text string) {
	c.Add(errors.New(text))
}

// AddNewf creates new error using specified format and adds to collection
func (c *ErrorCollector) AddNewf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	c.Add(err)
}

// HasErrors returns true if there were any errors collected
func (c *ErrorCollector) HasErrors() bool {
	return len(c.errs) > 0
}

// Combine returns all collected errors as a single error
func (c *ErrorCollector) Combine() error {
	if c.HasErrors() {
		return errors.New(c.String())
	}
	return nil
}

// String returns string representation of ErrorCollector (all errors concatenated)
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
