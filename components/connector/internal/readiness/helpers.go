package readiness

import "sync"

type atomicBool struct {
	isCacheLoaded bool
	m             *sync.Mutex
}

func NewAtomicBool(val bool) *atomicBool {
	return &atomicBool{
		isCacheLoaded: val,
		m:             &sync.Mutex{},
	}
}

func (ab *atomicBool) getValue() bool {
	ab.m.Lock()
	defer ab.m.Unlock()
	return ab.isCacheLoaded
}

func (ab *atomicBool) setValue(val bool) {
	ab.m.Lock()
	defer ab.m.Unlock()
	ab.isCacheLoaded = val
}
