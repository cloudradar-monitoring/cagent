// +build windows

package cagent

import (
	"os"
	"path/filepath"
)

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)

	DefaultCfgPath = filepath.Join(exPath, "./cagent.conf")
	defaultLogPath = filepath.Join(exPath, "./cagent.log")
}
