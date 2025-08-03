package mqtt

import (
	"sync"
)

// EventEmitter é a estrutura principal que gerencia eventos e listeners
type EventEmitter struct {
	mu     sync.RWMutex
	events map[string][]chan interface{}
	once   map[string][]chan interface{}
	wg     sync.WaitGroup
	closed bool
}

// NewEventEmitter cria uma nova instância do EventEmitter
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		events: make(map[string][]chan interface{}),
		once:   make(map[string][]chan interface{}),
	}
}

// On adiciona um listener para um evento específico
func (ee *EventEmitter) On(event string, listener func(interface{})) {
	ch := make(chan interface{}, 8) // buffer maior para evitar bloqueio

	ee.mu.Lock()
	ee.events[event] = append(ee.events[event], ch)
	ee.mu.Unlock()

	ee.wg.Add(1)
	go func() {
		defer ee.wg.Done()
		for data := range ch {
			if ee.closed {
				return
			}
			listener(data)
		}
	}()
}

// Once adiciona um listener que será executado apenas uma vez
func (ee *EventEmitter) Once(event string, listener func(interface{})) {
	ch := make(chan interface{}, 1)

	ee.mu.Lock()
	ee.once[event] = append(ee.once[event], ch)
	ee.mu.Unlock()

	ee.wg.Add(1)
	go func() {
		defer ee.wg.Done()
		if data, ok := <-ch; ok && !ee.closed {
			listener(data)
		}
	}()
}

// Emit dispara um evento com os dados fornecidos
func (ee *EventEmitter) Emit(event string, data interface{}) {
	ee.mu.RLock()
	listeners := append([]chan interface{}(nil), ee.events[event]...)
	onceListeners := append([]chan interface{}(nil), ee.once[event]...)
	ee.mu.RUnlock()

	for _, ch := range listeners {
		select {
		case ch <- data:
		default:
		}
	}

	for _, ch := range onceListeners {
		select {
		case ch <- data:
		default:
		}
		close(ch)
	}

	if len(onceListeners) > 0 {
		ee.mu.Lock()
		delete(ee.once, event)
		ee.mu.Unlock()
	}
}

// RemoveListener remove um listener específico de um evento
func (ee *EventEmitter) RemoveListener(event string, _ func(interface{})) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	if listeners, ok := ee.events[event]; ok {
		for i, ch := range listeners {
			close(ch)
			ee.events[event] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// RemoveAllListeners remove todos os listeners de um evento (ou de todos os eventos)
func (ee *EventEmitter) RemoveAllListeners(events ...string) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	if len(events) == 0 {
		for _, listeners := range ee.events {
			for _, ch := range listeners {
				close(ch)
			}
		}
		ee.events = make(map[string][]chan interface{})

		for _, listeners := range ee.once {
			for _, ch := range listeners {
				close(ch)
			}
		}
		ee.once = make(map[string][]chan interface{})
	} else {
		for _, event := range events {
			if listeners, ok := ee.events[event]; ok {
				for _, ch := range listeners {
					close(ch)
				}
				delete(ee.events, event)
			}
			if listeners, ok := ee.once[event]; ok {
				for _, ch := range listeners {
					close(ch)
				}
				delete(ee.once, event)
			}
		}
	}
}

// Close libera todos os recursos do EventEmitter
func (ee *EventEmitter) Close() {
	ee.mu.Lock()
	ee.closed = true
	ee.RemoveAllListeners()
	ee.mu.Unlock()
	ee.wg.Wait()
}
