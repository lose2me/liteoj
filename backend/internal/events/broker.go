// Package events provides an in-process pub/sub broker. Publishers (judge
// workers, submission handler) fan out Event values to all live SSE
// subscribers. Slow consumers are silently dropped — after reconnect or a
// fresh page load the client will refetch current state, so a missed event
// is not catastrophic.
package events

import "sync"

// Event is the wire payload. Data is marshalled as JSON by the SSE handler.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type Broker struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: map[chan Event]struct{}{}}
}

// Subscribe returns a buffered channel the caller must drain. When done, pass
// it back to Unsubscribe. The broker will not close a channel it still owns,
// so losing the return reference leaks one goroutine's worth of buffer.
func (b *Broker) Subscribe() chan Event {
	ch := make(chan Event, 8)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
}

// Publish fans out to every subscriber. Non-blocking on full buffers — a
// subscriber that can't keep up forfeits an event rather than stalling the
// publisher (typically a judge worker) or other subscribers.
func (b *Broker) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- e:
		default:
		}
	}
}
