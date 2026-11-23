package events

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventUnidadRegistrada        EventType = "UnidadRegistrada"
	EventEstadoUnidadActualizado EventType = "EstadoUnidadActualizado"
	EventUnidadProximaACaducar   EventType = "UnidadProximaACaducar"
	EventEmergenciaReservada     EventType = "EmergenciaReservada"
	EventSolicitudCancelada      EventType = "SolicitudCancelada"
	EventStockCriticoDetectado   EventType = "StockCriticoDetectado"
)

type DomainEvent struct {
	ID        string         `json:"id"`
	Type      EventType      `json:"evento"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

type EventHandler func(ctx context.Context, event DomainEvent)

// IEventBus defines the outbound/inbound port for asynchronous domain event publishing.
type IEventBus interface {
	Publish(ctx context.Context, eventType EventType, payload map[string]any)
	Subscribe(eventType EventType, handler EventHandler)
	IsHealthy() bool
	Close()
}

type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]EventHandler
	queue       chan DomainEvent
	wg          sync.WaitGroup
	closed      atomic.Bool
}

func NewInMemoryEventBus(bufferSize int, workers int) *InMemoryEventBus {
	if bufferSize <= 0 {
		bufferSize = 512
	}
	if workers <= 0 {
		workers = 4
	}
	bus := &InMemoryEventBus{
		subscribers: make(map[EventType][]EventHandler),
		queue:       make(chan DomainEvent, bufferSize),
	}
	for i := 0; i < workers; i++ {
		bus.wg.Add(1)
		go bus.worker()
	}
	return bus
}

func (b *InMemoryEventBus) worker() {
	defer b.wg.Done()
	for ev := range b.queue {
		b.mu.RLock()
		handlers := append([]EventHandler(nil), b.subscribers[ev.Type]...)
		b.mu.RUnlock()

		for _, h := range handlers {
			func(handler EventHandler) {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic recovered in event handler", "event", ev.Type, "recover", r)
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				handler(ctx, ev)
			}(h)
		}
	}
}

func (b *InMemoryEventBus) Publish(_ context.Context, eventType EventType, payload map[string]any) {
	if b.closed.Load() {
		return
	}
	ev := DomainEvent{
		ID:        uuid.NewString(),
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
	select {
	case b.queue <- ev:
	default:
		slog.Warn("event bus buffer full, dropping event", "type", eventType)
	}
}

func (b *InMemoryEventBus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

func (b *InMemoryEventBus) IsHealthy() bool {
	return !b.closed.Load()
}

func (b *InMemoryEventBus) Close() {
	if b.closed.CompareAndSwap(false, true) {
		close(b.queue)
		b.wg.Wait()
	}
}
