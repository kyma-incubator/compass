package pkg

import "sync"

// Adapters contains pairing adapters configuration mapping between integration system and adapter URL
type Adapters struct {
	mutex   sync.RWMutex
	Mapping map[string]string
}

// NewAdapters return empty Adapters struct
func NewAdapters() *Adapters {
	return &Adapters{}
}

// Get return pairing adapter mapping configuration
func (a *Adapters) Get() map[string]string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.Mapping
}

// Update updates pairing adapter mapping with the given value
func (a *Adapters) Update(mapping map[string]string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.Mapping = mapping
}
