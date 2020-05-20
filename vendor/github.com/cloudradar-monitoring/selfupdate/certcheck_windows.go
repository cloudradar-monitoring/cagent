package selfupdate

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

var (
	msiDLL                              = windows.NewLazySystemDLL("msi.dll")
	procMsiGetFileSignatureInformationW = msiDLL.NewProc("MsiGetFileSignatureInformationW")

	crypt32DLL                     = windows.NewLazySystemDLL("crypt32.dll")
	procCertFreeCertificateContext = crypt32DLL.NewProc("CertFreeCertificateContext")
	procCertGetNameString          = crypt32DLL.NewProc("CertGetNameStringW")
)

const (
	CERT_NAME_SIMPLE_DISPLAY_TYPE = 4
)

func getMSICertificateDisplayName(packageFilePath string) (string, error) {
	err := procMsiGetFileSignatureInformationW.Find()
	if err != nil {
		return "", err
	}
	err = procCertFreeCertificateContext.Find()
	if err != nil {
		return "", err
	}
	err = procCertGetNameString.Find()
	if err != nil {
		return "", err
	}

	packageFilePathPtr, err := syscall.UTF16PtrFromString(packageFilePath)
	if err != nil {
		return "", err
	}

	var certContextPtr int
	r1, _, err := procMsiGetFileSignatureInformationW.Call(
		uintptr(unsafe.Pointer(packageFilePathPtr)),
		uintptr(0),
		uintptr(unsafe.Pointer(&certContextPtr)),
		0,
		0,
	)
	if r1 != 0 {
		return "", errors.Wrapf(err, "unable to get certificate from file. code 0x%x", r1)
	}

	certName := ""
	certQueryResult := make([]uint16, 1024)
	r1, _, err = procCertGetNameString.Call(
		uintptr(certContextPtr),
		uintptr(CERT_NAME_SIMPLE_DISPLAY_TYPE),
		uintptr(0),
		0,
		uintptr(unsafe.Pointer(&certQueryResult[0])),
		uintptr(len(certQueryResult)),
	)
	if r1 != 0 {
		certName = windows.UTF16ToString(certQueryResult)
	}

	r1, _, err = procCertFreeCertificateContext.Call(uintptr(certContextPtr))
	if r1 == 0 {
		return "", err
	}
	return certName, nil
}
