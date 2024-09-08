package signals

type ListenerEvent[T comparable] func(T, *BaseSignal[T])

type ListenerWrapper[T comparable] struct {
	listener ListenerEvent[T]
	isAsync  bool
}

type signalReceiver interface {
	DependencyChanged()
	TriggerListeners()
	TriggerAsyncListeners()
}

type signalSender interface {
	AddBelowDependency(signalReceiver)
}

type BaseSignal[T comparable] interface {
	Get() T

	ListenByEvent(*ListenerEvent[T], ...string) (string, error)
	Listen(func(T, *BaseSignal[T]), ...string) (string, error)
	ListenAsyncByEvent(*ListenerEvent[T], ...string) (string, error)
	ListenAsync(func(T, *BaseSignal[T]), ...string) (string, error)

	Unlisten(*ListenerEvent[T])
	UnlistenById(string)
	UnlistenAll()
}

func MakeEventListener[T comparable](fn func(T, *BaseSignal[T])) *ListenerEvent[T] {
	listener := ListenerEvent[T](fn)
	return &listener
}
