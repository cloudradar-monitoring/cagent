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

var providers map[string]*provEntry
var lock sync.Mutex

var (
	ErrAlreadyExists = errors.New("vmstat: provider already registered")
	ErrNotAvailable  = errors.New("vmstat: provider not available")
	ErrCheck         = errors.New("vmstat: check provider availability")
	ErrNotRegistered = errors.New("vmstat: provider not registered")
)

func init() {
	providers = make(map[string]*provEntry)
}

func RegisterVMProvider(p Provider) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := providers[p.Name()]; ok {
		return fmt.Errorf("%s: %s", ErrAlreadyExists.Error(), p.Name())
	}

	providers[p.Name()] = &provEntry{prov: p}

	return nil
}

func Acquire(name string) (Provider, error) {
	lock.Lock()
	defer lock.Unlock()

	var err error

	if entry, ok := providers[name]; ok {
		entry.wg.Add(1)
		entry.run.Do(func() {
			if err = entry.prov.Run(); err != nil {
				logrus.WithError(err).Errorf("vmstat: unable to start \"%s\" provider", name)
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
	lock.Lock()
	defer lock.Unlock()

	if entry, ok := providers[p.Name()]; ok {
		entry.wg.Done()
		return nil
	}

	return ErrNotRegistered
}

func IterateRegistered(f func(string, Provider) bool) {
	lock.Lock()
	defer lock.Unlock()

	for name, p := range providers {
		if !f(name, p.prov) {
			break
		}
	}
}
