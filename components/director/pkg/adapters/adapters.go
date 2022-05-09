package pkg

import "sync"

type Adapters struct {
	mutex   sync.RWMutex
	Mapping map[string]string
}

func NewAdapters() *Adapters {
	return &Adapters{}
}

func (a *Adapters) Get() map[string]string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.Mapping
}

func (a *Adapters) Update(mapping map[string]string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.Mapping = mapping
}
