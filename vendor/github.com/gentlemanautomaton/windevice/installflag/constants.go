package installflag

// Windows device installation and update flags.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/newdev/nf-newdev-updatedriverforplugandplaydevicesw
const (
	Force          = 0x00000001 // INSTALLFLAG_FORCE
	ReadOnly       = 0x00000002 // INSTALLFLAG_READONLY
	NonInteractive = 0x00000004 // INSTALLFLAG_NONINTERACTIVE
)
