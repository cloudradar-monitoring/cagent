package installstate

import "fmt"

// Windows device installation states.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/ddi/content/wdm/ne-wdm-_device_install_state
const (
	Installed      = 0 // InstallStateInstalled
	NeedsReinstall = 1 // InstallStateNeedsReinstall
	FailedInstall  = 2 // InstallStateFailedInstall
	FinishInstall  = 3 // InstallStateFinishInstall
)

// State represents an installation state
type State uint32

// String returns a string representation of the installation state.
func (state State) String() string {
	switch state {
	case Installed:
		return "Installed"
	case NeedsReinstall:
		return "NeedsReinstall"
	case FailedInstall:
		return "FailedInstall"
	case FinishInstall:
		return "FinishInstall"
	default:
		return fmt.Sprintf("UnknownState %d", state)
	}
}
