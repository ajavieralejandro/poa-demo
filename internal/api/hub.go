package api

import "sync"

// Hub reparte cada bloque ya serializado a los clientes SSE conectados.
// Un cliente lento no detiene el minado: si su búfer está lleno, se omite el evento.
type Hub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
}

// NewHub crea un distribuidor vacío.
func NewHub() *Hub {
	return &Hub{clients: make(map[chan []byte]struct{})}
}

// Subscribe registra un cliente. La función devuelta lo da de baja y cierra el canal.
func (h *Hub) Subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, 8)
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

// Publish envía una copia del payload a cada cliente.
func (h *Hub) Publish(payload []byte) {
	if len(payload) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		message := append([]byte(nil), payload...)
		select {
		case ch <- message:
		default:
		}
	}
}
