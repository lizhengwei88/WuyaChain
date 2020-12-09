package event

import (
	"reflect"
	"sync"
)

// EventManager is an interface which defines the event manager behaviors
// Note that it is thread safe
type EventManager struct {
	lock      sync.RWMutex
	listeners []eventListener
}


// NewEventManager creates a new instance of event manager
func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make([]eventListener, 0),
	}
}

// Fire triggers all listeners with the specified event and removes listeners which run only once.
func (h *EventManager) Fire(e Event) {
	h.lock.Lock()
	defer h.lock.Unlock()

	for _, l := range h.listeners {
		if l.IsAsyncListener {
			go l.Callable(e)
		} else {
			l.Callable(e)
		}
	}

	h.removeOnceListener()
}

// AddListener registers a listener.
// If there is already a same listener (same method pointer), we will not add it
func (h *EventManager) AddListener(callback EventHandleMethod) {
	listener := eventListener{
		Callable: callback,
	}

	h.addEventListener(listener)
}

// removeOnceListener removes all listeners which run only once
func (h *EventManager) removeOnceListener() {
	listeners := make([]eventListener, 0, len(h.listeners))
	for _, l := range h.listeners {
		if !l.IsOnceListener {
			listeners = append(listeners, l)
		}
	}

	h.listeners = listeners
}

// AddAsyncListener registers a listener which runs async
func (h *EventManager) AddAsyncListener(callback EventHandleMethod) {
	listener := eventListener{
		Callable:        callback,
		IsAsyncListener: true,
	}

	h.addEventListener(listener)
}

// addEventListener registers a event listener.
// If there is already a same listener (same method pointer), we will not add it
func (h *EventManager) addEventListener(listener eventListener) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if index := h.find(listener.Callable); index != -1 {
		return
	}

	h.listeners = append(h.listeners, listener)
}


// find finds listener existing in the manager
// returns -1 if not found, otherwise the index of the listener
func (h *EventManager) find(callback EventHandleMethod) int {
	p := reflect.ValueOf(callback).Pointer()

	for i, l := range h.listeners {
		lp := reflect.ValueOf(l.Callable).Pointer()
		if lp == p {
			return i
		}
	}

	return -1
}