package common

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

const linuxRootCertsPath = "/etc/cagent/cacert.pem"

var ErrorCustomRootCertPoolNotImplementedForOS = fmt.Errorf("not implemented for current os")

func CustomRootCertPool() (*x509.CertPool, error) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return nil, ErrorCustomRootCertPoolNotImplementedForOS
	}

	if _, err := os.Stat(linuxRootCertsPath); err != nil {
		return nil, fmt.Errorf("root certs file not found")
	}

	certPool := x509.NewCertPool()

	b, err := ioutil.ReadFile(linuxRootCertsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cacert.pem: %s", err.Error())
	}

	ok := certPool.AppendCertsFromPEM(b)
	if ok {
		return certPool, nil
	}

	return nil, fmt.Errorf("failed to AppendCertsFromPEM")
}
