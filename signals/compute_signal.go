package signals

import (
	"errors"
	"fmt"
	"sync"
)

type ComputedSignal[T comparable] struct {
	value             T
	lastRecalcChanged bool
	mapFn             func() T

	rwValue sync.RWMutex

	listeners         map[string]ListenerWrapper[T]
	belowDependencies []signalReceiver
}

func MakeComputedSignal[T comparable](mapFn func() T, dependsOn ...signalSender) *ComputedSignal[T] {
	if mapFn == nil || len(dependsOn) == 0 {
		return nil
	}

	cmpSignal := &ComputedSignal[T]{
		mapFn:             mapFn,
		lastRecalcChanged: true,
		rwValue:           sync.RWMutex{},
		listeners:         map[string]ListenerWrapper[T]{},
		belowDependencies: []signalReceiver{},
	}

	for _, dep := range dependsOn {
		dep.AddBelowDependency(cmpSignal)
	}

	return cmpSignal
}

func (cs *ComputedSignal[T]) AddBelowDependency(sr signalReceiver) {
	cs.belowDependencies = append(cs.belowDependencies, sr)
}

func (cs *ComputedSignal[T]) DependencyChanged() {
	newValue := cs.mapFn()

	if cs.Get() == newValue {
		return
	}

	cs.rwValue.Lock()
	cs.lastRecalcChanged = true
	cs.value = newValue
	cs.rwValue.Unlock()

	for _, dep := range cs.belowDependencies {
		dep.DependencyChanged()
	}
}

func (cs *ComputedSignal[T]) TriggerListeners() {
	value := cs.Get()
	var bs BaseSignal[T] = cs

	for _, lsWrapper := range cs.listeners {
		if !lsWrapper.isAsync {
			lsWrapper.listener(value, &bs)
		}
	}

	for _, dep := range cs.belowDependencies {
		dep.TriggerListeners()
	}
}

func (cs *ComputedSignal[T]) TriggerAsyncListeners() {
	value := cs.Get()
	var bs BaseSignal[T] = cs

	// we are already in a go routine started from the main signal
	for _, lsWrapper := range cs.listeners {
		if lsWrapper.isAsync {
			lsWrapper.listener(value, &bs)
		}
	}

	// idk if its worth having those in different go routines, too many threads might get spawned and less control
	for _, dep := range cs.belowDependencies {
		dep.TriggerAsyncListeners()
	}
}

func (cs *ComputedSignal[T]) Get() T {
	cs.rwValue.RLock()

	if cs.lastRecalcChanged {
		cs.rwValue.RUnlock()

		cs.rwValue.Lock()
		defer cs.rwValue.Unlock()
		cs.value = cs.mapFn()
		cs.lastRecalcChanged = false
		return cs.value
	}

	defer cs.rwValue.RUnlock()
	return cs.value
}

func (cs *ComputedSignal[T]) ListenByEvent(listener *ListenerEvent[T], id ...string) (string, error) {
	if listener == nil {
		return "", errors.New("listener is null")
	}

	var newId string
	if len(id) > 0 {
		newId = id[0]
	} else {
		newId = fmt.Sprintf("%v", listener)
	}

	cs.listeners[newId] = ListenerWrapper[T]{listener: *listener, isAsync: false}
	return newId, nil
}

func (cs *ComputedSignal[T]) Listen(listener func(T, *BaseSignal[T]), id ...string) (string, error) {
	if listener == nil {
		return "", errors.New("listener is null")
	}

	event := MakeEventListener(listener)
	return cs.ListenByEvent(event, id...)
}

func (cs *ComputedSignal[T]) ListenAsyncByEvent(listener *ListenerEvent[T], id ...string) (string, error) {
	if listener == nil {
		return "", errors.New("listener is null")
	}

	var newId string
	if len(id) > 0 {
		newId = id[0]
	} else {
		newId = fmt.Sprintf("%v", listener)
	}

	cs.listeners[newId] = ListenerWrapper[T]{listener: *listener, isAsync: true}
	return newId, nil
}

func (cs *ComputedSignal[T]) ListenAsync(listener func(T, *BaseSignal[T]), id ...string) (string, error) {
	if listener == nil {
		return "", errors.New("listener is null")
	}

	event := MakeEventListener(listener)
	return cs.ListenAsyncByEvent(event, id...)
}

func (cs *ComputedSignal[T]) Unlisten(listener *ListenerEvent[T]) {
	id := fmt.Sprintf("%v", listener)
	delete(cs.listeners, id)
}

func (cs *ComputedSignal[T]) UnlistenById(id string) {
	delete(cs.listeners, id)
}

func (cs *ComputedSignal[T]) UnlistenAll() {
	cs.listeners = map[string]ListenerWrapper[T]{}
}
