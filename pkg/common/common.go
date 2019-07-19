package common

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	Timeout    = 3 * time.Second
	ErrTimeout = errors.New("invoker: command timed out")
)

// Invoker executes command in context and gathers stdout/stderr output into slice
type Invoker interface {
	CommandWithContext(context.Context, string, ...string) ([]byte, error)
}

type Invoke struct{}

var _ Invoker = (*Invoke)(nil)

func (i Invoke) CommandWithContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, arg...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return buf.Bytes(), err
	}

	if err := cmd.Wait(); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

// RunCommandWithContext convenience wrapper to CommandWithContext
func RunCommandWithContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	var invoke Invoke

	return invoke.CommandWithContext(ctx, name, arg...)
}

// RunCommandInBackground convenience wrapper to RunCommandWithContext
func RunCommandInBackground(name string, arg ...string) ([]byte, error) {
	return RunCommandWithContext(context.Background(), name, arg...)
}

func MergeStringMaps(mapA, mapB map[string]interface{}) map[string]interface{} {
	for k, v := range mapB {
		mapA[k] = v
	}
	return mapA
}

func RoundToTwoDecimalPlaces(v float64) float64 {
	return math.Round(v*100) / 100
}

func FloatToIntRoundUP(f float64) int {
	return int(f + 0.5)
}

// GetEnv retrieves the environment variable key. If it does not exist it returns the default.
func GetEnv(key string, dfault string, combineWith ...string) string {
	value := os.Getenv(key)
	if value == "" {
		value = dfault
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
// from github.com/shriou/gopsutil/internal/common.go
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLines reads contents from file and splits them by new line.
// from github.com/shriou/gopsutil/internal/common.go
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

// StrInSlice returns true if search string found in slice
func StrInSlice(search string, slice []string) bool {
	for _, str := range slice {
		if str == search {
			return true
		}
	}
	return false
}
