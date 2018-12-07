package cagent

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
)

var (
	unameTimeout    = 5 * time.Second
	cpuInfoTimeout  = 10 * time.Second
	hostInfoTimeout = 10 * time.Second
)

type Invoker interface {
	Command(string, ...string) ([]byte, error)
	CommandWithContext(context.Context, string, ...string) ([]byte, error)
}

type Invoke struct{}

func (i Invoke) Command(name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), unameTimeout)
	defer cancel()
	return i.CommandWithContext(ctx, name, arg...)
}

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

var invoke Invoker = Invoke{}

func Uname() (string, error) {
	if runtime.GOOS == "windows" {
		info, err := host.Info()

		if err != nil {
			return "", err
		}
		return info.Platform + " " + info.PlatformVersion + " " + info.PlatformFamily, nil
	}
	uname, err := exec.LookPath("uname")
	if err != nil {
		return "", err
	}
	b, err := invoke.Command(uname, "-a")
	return string(b), err
}

func (ca *Cagent) HostInfoResults() (MeasurementsMap, error) {
	res := MeasurementsMap{}

	if len(ca.Config.SystemFields) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), hostInfoTimeout)
	defer cancel()

	info, err := host.InfoWithContext(ctx)
	errs := []string{}

	if err != nil {
		log.Errorf("[SYSTEM] Failed to read host info: %s", err.Error())
		errs = append(errs, err.Error())
	}

	for _, field := range ca.Config.SystemFields {
		switch strings.ToLower(field) {
		case "os_kernel":
			if info != nil {
				res[field] = info.OS
			} else {
				res[field] = nil
			}
		case "os_family":
			if info != nil {
				res[field] = info.PlatformFamily
			} else {
				res[field] = nil
			}
		case "uname":
			uname, err := Uname()
			if err != nil {
				log.Errorf("[SYSTEM] Failed to read host uname: %s", err.Error())
				errs = append(errs, err.Error())
				res[field] = nil
			} else {
				res[field] = uname
			}
		case "fqdn":
			res[field] = getFQDN()
		case "cpu_model":
			ctx, cancel := context.WithTimeout(context.Background(), cpuInfoTimeout)
			defer cancel()
			cpuInfo, err := cpu.InfoWithContext(ctx)
			if err != nil {
				log.Errorf("[SYSTEM] Failed to read cpu info: %s", err.Error())
				errs = append(errs, err.Error())
				res[field] = nil
				continue
			}

			res[field] = cpuInfo[0].ModelName
		case "os_arch":
			res[field] = runtime.GOARCH
		case "memory_total_b":
			memStat, err := mem.VirtualMemory()
			if err != nil {
				log.Errorf("[SYSTEM] Failed to read mem info: %s", err.Error())
				errs = append(errs, err.Error())
				res[field] = nil
				continue
			}

			res[field] = memStat.Total
		}
	}

	if len(errs) == 0 {
		return res, nil
	}

	return res, errors.New("SYSTEM: " + strings.Join(errs, "; "))
}

func getFQDN() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, ".")
		}
	}
	return hostname
}
