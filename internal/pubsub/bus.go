package pubsub

import (
	"context"
	"sync"
	"time"

	"sharing-vision-backend/internal/model"
)

type EventType string

const (
	EventCreated EventType = "created"
	EventUpdated EventType = "updated"
	EventDeleted EventType = "deleted"
)

type Event struct {
	Type     EventType `json:"type"`
	Post     *model.Post
	Occurred time.Time `json:"occurred_at"`
}

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[int]chan Event
	nextID      int
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: map[int]chan Event{},
	}
}

func (b *EventBus) Subscribe(ctx context.Context) <-chan Event {
	ch := make(chan Event, 16)

	b.mu.Lock()
	b.nextID++
	id := b.nextID
	b.subscribers[id] = ch
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.unsubscribe(id)
	}()

	return ch
}

func (b *EventBus) unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[id]; ok {
		delete(b.subscribers, id)
		close(ch)
	}
}

func (b *EventBus) Publish(_ context.Context, event Event) {
	if event.Occurred.IsZero() {
		event.Occurred = time.Now()
	}

	b.mu.RLock()
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
	b.mu.RUnlock()
}
