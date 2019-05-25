// +build windows

package common

func IsCommandAvailable(name string) bool {
	return false
}

func WrapCommandToAdmin(args ...string) string {
	var cmd string
	for _, s := range args {
		if cmd != "" {
			cmd += " "
		}

		cmd += s
	}

	return cmd
}
