package main

import (
	"fmt"
	"log"

	"poa-demo/internal/config"
	"poa-demo/internal/engine"
)

func main() {
	fmt.Println("=== Demo de PoA y Follow the Satoshi (Stress Test) ===")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	eng, err := engine.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	blocks, _, _ := eng.Snapshot()
	fmt.Printf("[Génesis] Creado en Height: %d (supply %d, dificultad %d)\n", blocks[0].Header.Height, cfg.InitialSupply, cfg.Difficulty)

	totalBloques := 100
	for i := 1; i <= totalBloques; i++ {
		nextBlock, err := eng.MineNext()
		if err != nil {
			panic(fmt.Sprintf("Error minando el bloque %d: %v", i, err))
		}

		if i%10 == 0 || i == 1 {
			fmt.Printf("-> [Bloque %d] Minado y verificado con éxito (Hash: %x...)\n", nextBlock.Header.Height, nextBlock.Header.Hash())
		}
	}

	fmt.Printf("\n=== ¡Los %d bloques se minaron y verificaron de punta a punta sin errores! ===\n", totalBloques)
}
