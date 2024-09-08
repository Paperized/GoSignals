package signals

import (
	"errors"
	"fmt"
	"sync"
)

type Signal[T comparable] struct {
	value         T
	setValueMutex sync.Mutex
	rwValue       sync.RWMutex

	listeners         map[string]ListenerWrapper[T]
	belowDependencies []signalReceiver
}

func MakeSignal[T comparable](optionalValue ...T) *Signal[T] {
	var newValue T
	if len(optionalValue) > 0 {
		newValue = optionalValue[0]
	}

	return &Signal[T]{
		listeners:         map[string]ListenerWrapper[T]{},
		belowDependencies: []signalReceiver{},
		value:             newValue,
		setValueMutex:     sync.Mutex{},
		rwValue:           sync.RWMutex{},
	}
}

func (cs *Signal[T]) AddBelowDependency(sr signalReceiver) {
	cs.belowDependencies = append(cs.belowDependencies, sr)
}

func (cs *Signal[T]) SetFromValue(fn func(T) T) T {
	cs.setValueMutex.Lock()
	defer cs.setValueMutex.Unlock()

	return cs.internalSet(fn(cs.value))
}

func (cs *Signal[T]) Set(value T, forceUpdate ...bool) T {
	cs.setValueMutex.Lock()
	defer cs.setValueMutex.Unlock()

	return cs.internalSet(value, forceUpdate...)
}

func (cs *Signal[T]) internalSet(value T, forceUpdate ...bool) T {
	if cs.Get() == value && (len(forceUpdate) == 0 || !forceUpdate[0]) {
		return value
	}

	cs.rwValue.Lock()
	prevValue := cs.value
	cs.value = value
	cs.rwValue.Unlock()

	// update below dependencies
	for _, dep := range cs.belowDependencies {
		dep.DependencyChanged()
	}

	// Call listeners for this signal value
	var bs BaseSignal[T] = cs

	// wrap in goroutine
	wg := sync.WaitGroup{}

	for _, lsWrapper := range cs.listeners {
		if lsWrapper.isAsync {
			wg.Add(1)
			go func() {
				defer wg.Done()
				lsWrapper.listener(cs.value, &bs)
			}()
		}
	}

	for _, lsWrapper := range cs.listeners {
		if !lsWrapper.isAsync {
			lsWrapper.listener(cs.value, &bs)
		}
	}

	wg.Wait()

	// Then we trigger the below dependencies listeners
	for _, dep := range cs.belowDependencies {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dep.TriggerAsyncListeners()
		}()
	}

	for _, dep := range cs.belowDependencies {
		dep.TriggerListeners()
	}

	wg.Wait()
	return prevValue
}

func (cs *Signal[T]) Get() T {
	cs.rwValue.RLock()
	defer cs.rwValue.RUnlock()
	return cs.value
}

func (cs *Signal[T]) ListenByEvent(listener *ListenerEvent[T], id ...string) (string, error) {
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

func (cs *Signal[T]) Listen(listener func(T, *BaseSignal[T]), id ...string) (string, error) {
	if listener == nil {
		return "", errors.New("listener is null")
	}

	event := MakeEventListener(listener)
	return cs.ListenByEvent(event, id...)
}

func (cs *Signal[T]) ListenAsyncByEvent(listener *ListenerEvent[T], id ...string) (string, error) {
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

func (cs *Signal[T]) ListenAsync(listener func(T, *BaseSignal[T]), id ...string) (string, error) {
	if listener == nil {
		return "", errors.New("listener is null")
	}

	event := MakeEventListener(listener)
	return cs.ListenAsyncByEvent(event, id...)
}

func (cs *Signal[T]) Unlisten(listener *ListenerEvent[T]) {
	id := fmt.Sprintf("%v", listener)
	delete(cs.listeners, id)
}

func (cs *Signal[T]) UnlistenById(id string) {
	delete(cs.listeners, id)
}

func (cs *Signal[T]) UnlistenAll() {
	cs.listeners = map[string]ListenerWrapper[T]{}
}
