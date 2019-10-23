//+build windows

package storcli

func (s *StorCLI) getCommandLine() []string {
	return []string{s.binaryPath, "/call", "show", "all", "J"}
}
