//+build !windows

package storcli

func (s *StorCLI) getCommandLine() []string {
	return []string{"sudo", s.binaryPath, "/call", "show", "all", "J"}
}
