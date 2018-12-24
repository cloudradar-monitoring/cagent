package vmstat

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
)

type provEntry struct {
	prov vmstattypes.Provider
	wg   sync.WaitGroup
	run  sync.Once
}

type vmstat struct {
	pr   map[string]*provEntry
	lock sync.Mutex
}

var providers *vmstat

func init() {
	providers = &vmstat{
		pr: make(map[string]*provEntry),
	}
}

func RegisterVMProvider(p vmstattypes.Provider) error {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if _, ok := providers.pr[p.Name()]; ok {
		return fmt.Errorf("%s: %s", vmstattypes.ErrAlreadyExists.Error(), p.Name())
	}

	providers.pr[p.Name()] = &provEntry{prov: p}

	return nil
}

func Acquire(name string) (vmstattypes.Provider, error) {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	var err error

	if entry, ok := providers.pr[name]; ok {
		entry.wg.Add(1)
		entry.run.Do(func() {
			if err = entry.prov.Run(); err != nil {
				logrus.Errorf("[vmstat] unable to start \"%s\" provider: %s", name, err.Error())
			}
		})

		if err != nil {
			entry.run = sync.Once{}
			return nil, err
		}
		return entry.prov, nil
	}

	return nil, vmstattypes.ErrNotRegistered
}

func Release(p vmstattypes.Provider) error {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if entry, ok := providers.pr[p.Name()]; ok {
		entry.wg.Done()
		return nil
	}

	return vmstattypes.ErrNotRegistered
}

func IterateRegistered(f func(string, vmstattypes.Provider) bool) {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	for name, p := range providers.pr {
		if !f(name, p.prov) {
			break
		}
	}
}
