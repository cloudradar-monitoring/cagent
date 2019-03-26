package infpath

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// Prepare takes a setup information file path and prepares it for use in
// various windows API system calls.
//
// If the path is relative it will be transformed into an absolute path.
//
// An error will be returned if the the path is too long or describes a
// non-existant or non-regular file.
func Prepare(path string) (string, error) {
	// Check for empty parameters
	if path == "" {
		return "", errors.New("an empty inf file path was provided")
	}

	// Windows does special things when given a relative path to an INF file,
	// so make sure the path is absolute
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("unable to construct an absolute inf file path: %v", err)
	}

	// Make sure the path length doesn't exceed MAX_PATH
	if len(path)+1 >= windows.MAX_PATH {
		return "", fmt.Errorf("path length exceeds the %d character limit specified by MAX_PATH: %s", windows.MAX_PATH, path)
	}

	// Make sure the INF file exists and is a regular file
	fi, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("unable to open inf file: %v", err)
	}
	if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("inf path is not a regular file: %s", path)
	}

	return path, nil
}
