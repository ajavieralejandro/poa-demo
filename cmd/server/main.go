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

	go eng.Run(ctx, func(notice engine.Notice) {
		switch notice.Kind {
		case engine.NoticeBlock:
			payload, err := json.Marshal(api.BlockEvent{
				BlockView:  api.NewBlockView(notice.Block),
				Generation: notice.Generation,
			})
			if err != nil {
				log.Printf("serializar bloque %d: %v", notice.Block.Header.Height, err)
				return
			}
			hub.PublishEvent("block", payload)
			log.Printf("bloque %d verificado y emitido", notice.Block.Header.Height)
		case engine.NoticeStatus:
			payload, err := json.Marshal(api.StatusFromView(notice.State))
			if err != nil {
				log.Printf("serializar estado: %v", err)
				return
			}
			hub.PublishEvent("status", payload)
			log.Printf("límite de la demo: generación %d, %d bloques", notice.State.Generation, notice.State.BlockCount)
		}
	})

	errCh := make(chan error, 1)
	go func() {
		log.Printf("API en http://localhost%s (dificultad %d, supply %d, máximo %d bloques, intervalo %s)", cfg.ListenAddr(), cfg.Difficulty, cfg.InitialSupply, cfg.MaxBlocks, cfg.BlockInterval)
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
