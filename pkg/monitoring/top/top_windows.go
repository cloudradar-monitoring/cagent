// +build windows

package top

import "errors"

func (t *Top) GetProcesses() ([]*ProcessInfo, error) {
	return nil, errors.New("Not implemented")
}
