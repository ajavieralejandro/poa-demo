package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"poa-demo/internal/api"
	"poa-demo/internal/config"
	"poa-demo/internal/engine"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	eng, err := engine.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	hub := api.NewHub()
	handler := api.New(cfg, eng, hub).Handler()
	httpServer := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go runMiner(ctx, eng, hub, cfg.BlockInterval)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("API en http://localhost%s (dificultad %d, supply %d, intervalo %s)", cfg.ListenAddr(), cfg.Difficulty, cfg.InitialSupply, cfg.BlockInterval)
		log.Printf("stream SSE en http://localhost%s/api/v1/stream", cfg.ListenAddr())
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}

// runMiner repite la ronda de la demo y, tras cada verificación exitosa,
// emite el bloque a todos los clientes conectados al stream.
func runMiner(ctx context.Context, eng *engine.Engine, hub *api.Hub, interval time.Duration) {
	for {
		if ctx.Err() != nil {
			return
		}

		block, err := eng.MineNext()
		if err != nil {
			log.Printf("minado: %v", err)
			if !wait(ctx, time.Second) {
				return
			}
			continue
		}

		payload, err := json.Marshal(api.NewBlockView(block))
		if err != nil {
			log.Printf("serializar bloque %d: %v", block.Header.Height, err)
			continue
		}
		hub.Publish(payload)
		log.Printf("bloque %d verificado y emitido", block.Header.Height)

		if !wait(ctx, interval) {
			return
		}
	}
}

func wait(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
