package vmstat

import (
	"fmt"
	"sync"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
)

type provEntry struct {
	prov types.Provider
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

func RegisterVMProvider(p types.Provider) error {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if _, ok := providers.pr[p.Name()]; ok {
		return fmt.Errorf("%s: %s", types.ErrAlreadyExists.Error(), p.Name())
	}

	providers.pr[p.Name()] = &provEntry{prov: p}

	return nil
}

func Acquire(name string) (types.Provider, error) {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if entry, ok := providers.pr[name]; ok {
		var err error

		entry.run.Do(func() {
			if err = entry.prov.IsAvailable(); err != nil {
				err = types.ErrNotAvailable
				return
			}

			if err = entry.prov.Run(); err != nil {
				err = fmt.Errorf("vmstat: start vm provider \"%s\": %s", name, err.Error())
			}
		})

		if err != nil {
			entry.run = sync.Once{}
			return nil, err
		}
		entry.wg.Add(1)
		return entry.prov, nil
	}

	return nil, types.ErrNotRegistered
}

func Release(p types.Provider) error {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if entry, ok := providers.pr[p.Name()]; ok {
		entry.wg.Done()
		return nil
	}

	return types.ErrNotRegistered
}

func IterateRegistered(f func(string, types.Provider) bool) {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	for name, p := range providers.pr {
		if !f(name, p.prov) {
			break
		}
	}
}
