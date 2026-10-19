package api

import "sync"

type streamMessage struct {
	event string
	data  []byte
}

// Hub reparte cada evento ya serializado a los clientes SSE conectados.
// Un cliente lento no detiene el minado: si su búfer está lleno, se omite el evento.
type Hub struct {
	mu      sync.Mutex
	clients map[chan streamMessage]struct{}
}

// NewHub crea un distribuidor vacío.
func NewHub() *Hub {
	return &Hub{clients: make(map[chan streamMessage]struct{})}
}

// Subscribe registra un cliente. La función devuelta lo da de baja y cierra el canal.
func (h *Hub) Subscribe() (<-chan streamMessage, func()) {
	ch := make(chan streamMessage, 8)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if _, ok := h.clients[ch]; !ok {
				return
			}
			delete(h.clients, ch)
			close(ch)
		})
	}
	return ch, unsubscribe
}

// Publish envía un evento block. PublishEvent permite status u otros nombres.
func (h *Hub) Publish(payload []byte) {
	h.PublishEvent("block", payload)
}

// PublishEvent envía una copia del payload a cada cliente con el nombre de evento indicado.
func (h *Hub) PublishEvent(event string, payload []byte) {
	if event == "" || len(payload) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		message := streamMessage{event: event, data: append([]byte(nil), payload...)}
		select {
		case ch <- message:
		default:
		}
	}
}
