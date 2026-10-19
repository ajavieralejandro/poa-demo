package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/config"
)

// ChainSource es la vista de solo lectura que el servidor expone por HTTP.
type ChainSource interface {
	Snapshot() (blocks []blockchain.Block, supply uint32, difficulty int)
}

// Server publica la cadena y el stream de bloques verificados.
type Server struct {
	cfg      config.Config
	chain    ChainSource
	hub      *Hub
	origins  map[string]struct{}
	allowAll bool
}

// New construye el servidor con la configuración, la cadena y el hub de streaming.
func New(cfg config.Config, chain ChainSource, hub *Hub) *Server {
	origins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	allowAll := false
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" {
			allowAll = true
			continue
		}
		origins[origin] = struct{}{}
	}
	return &Server{
		cfg:      cfg,
		chain:    chain,
		hub:      hub,
		origins:  origins,
		allowAll: allowAll,
	}
}

// Handler registra los endpoints de la API v1 y aplica CORS.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/chain", s.handleChain)
	mux.HandleFunc("GET /api/v1/stream", s.handleStream)
	return s.withCORS(mux)
}

func (s *Server) handleChain(w http.ResponseWriter, _ *http.Request) {
	view := NewChainView(s.chain.Snapshot())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		http.Error(w, "no se pudo serializar la cadena", http.StatusInternalServerError)
	}
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming no disponible", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	events, unsubscribe := s.hub.Subscribe()
	defer unsubscribe()

	if _, err := fmt.Fprintf(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case payload, open := <-events:
			if !open {
				return
			}
			if _, err := fmt.Fprintf(w, "event: block\ndata: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin, ok := s.allowedOrigin(r.Header.Get("Origin")); ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) allowedOrigin(origin string) (string, bool) {
	if origin == "" {
		return "", false
	}
	if s.allowAll {
		return origin, true
	}
	if _, ok := s.origins[origin]; ok {
		return origin, true
	}
	return "", false
}
