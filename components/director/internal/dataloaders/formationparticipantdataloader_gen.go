// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package dataloader

import (
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// FormationParticipantDataloaderConfig captures the config to create a new FormationParticipantDataloader
type FormationParticipantDataloaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []ParamFormationParticipant) ([]graphql.FormationParticipant, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewFormationParticipantDataloader creates a new FormationParticipantDataloader given a fetch, wait, and maxBatch
func NewFormationParticipantDataloader(config FormationParticipantDataloaderConfig) *FormationParticipantDataloader {
	return &FormationParticipantDataloader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// FormationParticipantDataloader batches and caches requests
type FormationParticipantDataloader struct {
	// this method provides the data for the loader
	fetch func(keys []ParamFormationParticipant) ([]graphql.FormationParticipant, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[ParamFormationParticipant]graphql.FormationParticipant

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *formationParticipantDataloaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type formationParticipantDataloaderBatch struct {
	keys    []ParamFormationParticipant
	data    []graphql.FormationParticipant
	error   []error
	closing bool
	done    chan struct{}
}

// Load a FormationParticipant by key, batching and caching will be applied automatically
func (l *FormationParticipantDataloader) Load(key ParamFormationParticipant) (graphql.FormationParticipant, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a FormationParticipant.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *FormationParticipantDataloader) LoadThunk(key ParamFormationParticipant) func() (graphql.FormationParticipant, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() (graphql.FormationParticipant, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &formationParticipantDataloaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (graphql.FormationParticipant, error) {
		<-batch.done

		var data graphql.FormationParticipant
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *FormationParticipantDataloader) LoadAll(keys []ParamFormationParticipant) ([]graphql.FormationParticipant, []error) {
	results := make([]func() (graphql.FormationParticipant, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	formationParticipants := make([]graphql.FormationParticipant, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		formationParticipants[i], errors[i] = thunk()
	}
	return formationParticipants, errors
}

// LoadAllThunk returns a function that when called will block waiting for a FormationParticipants.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *FormationParticipantDataloader) LoadAllThunk(keys []ParamFormationParticipant) func() ([]graphql.FormationParticipant, []error) {
	results := make([]func() (graphql.FormationParticipant, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([]graphql.FormationParticipant, []error) {
		formationParticipants := make([]graphql.FormationParticipant, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			formationParticipants[i], errors[i] = thunk()
		}
		return formationParticipants, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *FormationParticipantDataloader) Prime(key ParamFormationParticipant, value graphql.FormationParticipant) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		l.unsafeSet(key, value)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *FormationParticipantDataloader) Clear(key ParamFormationParticipant) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *FormationParticipantDataloader) unsafeSet(key ParamFormationParticipant, value graphql.FormationParticipant) {
	if l.cache == nil {
		l.cache = map[ParamFormationParticipant]graphql.FormationParticipant{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *formationParticipantDataloaderBatch) keyIndex(l *FormationParticipantDataloader, key ParamFormationParticipant) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *formationParticipantDataloaderBatch) startTimer(l *FormationParticipantDataloader) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *formationParticipantDataloaderBatch) end(l *FormationParticipantDataloader) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
