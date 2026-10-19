package api

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/config"
	"poa-demo/internal/engine"
)

func TestChainAndStream(t *testing.T) {
	cfg := config.Config{
		Port:           "8080",
		Difficulty:     0,
		InitialSupply:  1000,
		AllowedOrigins: []string{"http://localhost:5173"},
		BlockInterval:  time.Second,
	}
	eng, err := engine.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	hub := NewHub()
	server := httptest.NewServer(New(cfg, eng, hub).Handler())
	t.Cleanup(server.Close)

	chainReq, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/chain", nil)
	if err != nil {
		t.Fatal(err)
	}
	chainReq.Header.Set("Origin", "http://localhost:5173")
	chainResp, err := http.DefaultClient.Do(chainReq)
	if err != nil {
		t.Fatal(err)
	}
	defer chainResp.Body.Close()
	if chainResp.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("CORS = %q", chainResp.Header.Get("Access-Control-Allow-Origin"))
	}

	var snapshot ChainView
	if err := json.NewDecoder(chainResp.Body).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Blocks) != 1 || snapshot.Blocks[0].Header.Height != 0 {
		t.Fatalf("génesis = %+v", snapshot.Blocks)
	}
	if snapshot.Supply != 1000 || snapshot.Difficulty != 0 {
		t.Fatalf("parámetros = supply %d dificultad %d", snapshot.Supply, snapshot.Difficulty)
	}

	streamReq, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	streamResp, err := http.DefaultClient.Do(streamReq)
	if err != nil {
		t.Fatal(err)
	}
	defer streamResp.Body.Close()
	if streamResp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("content-type = %s", streamResp.Header.Get("Content-Type"))
	}

	reader := bufio.NewReader(streamResp.Body)
	intro, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(intro, "connected") {
		t.Fatalf("apertura = %q", intro)
	}

	var mined blockchain.Block
	for attempt := 0; attempt < 20; attempt++ {
		mined, err = eng.MineNext()
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(NewBlockView(mined))
	if err != nil {
		t.Fatal(err)
	}
	hub.Publish(payload)

	var data string
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.HasPrefix(line, "data: ") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			break
		}
	}

	var view BlockView
	if err := json.Unmarshal([]byte(data), &view); err != nil {
		t.Fatal(err)
	}
	if view.Header.Height != mined.Header.Height || !view.Verified || view.Hash == "" {
		t.Fatalf("evento = %+v", view)
	}
	if len(view.Signatures) != 3 {
		t.Fatalf("firmas = %d", len(view.Signatures))
	}
}
