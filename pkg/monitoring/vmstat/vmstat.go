package vmstat

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Provider interface {
	Run() error
	Shutdown() error

	Name() string
	IsAvailable() error
	GetMeasurements() (map[string]interface{}, error)
}

type provEntry struct {
	prov Provider
	wg   sync.WaitGroup
	run  sync.Once
}

type vmstat struct {
	pr   map[string]*provEntry
	lock sync.Mutex
}

var providers *vmstat

var (
	ErrAlreadyExists = errors.New("vmstat: provider already registered")
	ErrNotAvailable  = errors.New("vmstat: provider not available")
	ErrCheck         = errors.New("vmstat: check provider availability")
	ErrNotRegistered = errors.New("vmstat: provider not registered")
)

func init() {
	providers = &vmstat{
		pr: make(map[string]*provEntry),
	}
}

func RegisterVMProvider(p Provider) error {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if _, ok := providers.pr[p.Name()]; ok {
		return fmt.Errorf("%s: %s", ErrAlreadyExists.Error(), p.Name())
	}

	providers.pr[p.Name()] = &provEntry{prov: p}

	return nil
}

func Acquire(name string) (Provider, error) {
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

	return nil, ErrNotRegistered
}

func Release(p Provider) error {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	if entry, ok := providers.pr[p.Name()]; ok {
		entry.wg.Done()
		return nil
	}

	return ErrNotRegistered
}

func IterateRegistered(f func(string, Provider) bool) {
	providers.lock.Lock()
	defer providers.lock.Unlock()

	for name, p := range providers.pr {
		if !f(name, p.prov) {
			break
		}
	}
}
