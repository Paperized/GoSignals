package signals

import "sync"

type ComputedValue[T any] struct {
	value T
	dirty bool
	mapFn func() T

	rwValue sync.RWMutex
}

func MakeComputedValue[T any](mapFn func() T, dependsOn ...signalSender) *ComputedValue[T] {
	if mapFn == nil || len(dependsOn) == 0 {
		return nil
	}

	cmpValue := &ComputedValue[T]{
		mapFn:   mapFn,
		dirty:   true,
		rwValue: sync.RWMutex{},
	}

	for _, dep := range dependsOn {
		dep.AddBelowDependency(cmpValue)
	}

	return cmpValue
}

func (cs *ComputedValue[T]) DependencyChanged() {
	cs.rwValue.Lock()
	defer cs.rwValue.Unlock()

	cs.dirty = true
}

// not really used since it's only a computed value and don't trigger anything
func (cs *ComputedValue[T]) TriggerListeners()      {}
func (cs *ComputedValue[T]) TriggerAsyncListeners() {}

func (cs *ComputedValue[T]) Get() T {
	cs.rwValue.RLock()

	if !cs.dirty {
		defer cs.rwValue.RUnlock()
		return cs.value
	}

	// would be better to have an upgradable rwlock
	cs.rwValue.RUnlock()

	cs.rwValue.Lock()
	defer cs.rwValue.Unlock()

	cs.value = cs.mapFn()
	cs.dirty = false
	return cs.value
}
