package main

import (
	"fmt"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/consensus"
	"poa-demo/internal/miners"
)

func main() {
	fmt.Println("=== Demo de PoA y Follow the Satoshi (Stress Test) ===")

	// 1. Configuración inicial
	var supply uint32 = 1000
	dificultad := 24

	// 2. Creamos el bloque Génesis (Raíz de la cadena)
	genesis := blockchain.NewGenesisBlock(supply)
	chain := blockchain.Chain{
		Blocks: make(map[blockchain.Hash]blockchain.Block),
		Tip:    genesis.Header.Hash(),
		Owners: make(map[blockchain.SatoshiID]blockchain.Address),
	}
	chain.Blocks[chain.Tip] = genesis
	addresses := make([]blockchain.Address, 100)
	for i := range 100 {
		addresses[i] = blockchain.Address(i + 1)
	}
	chain, nodos := miners.NewDeterministicPool(chain, addresses)
	fmt.Printf("[Génesis] Creado en Height: %d\n", genesis.Header.Height)

	// 3. Bucle para simular y verificar 100 bloques encadenados
	totalBloques := 100
	for i := 1; i <= totalBloques; i++ {
		// Minamos el bloque hijo utilizando el bloque actual como padre (tip)
		nextBlock, err := nodos[blockchain.Address(i)].Mine(chain, dificultad)
		if err != nil {
			panic(fmt.Sprintf("Error minando el bloque %d: %v", i, err))
		}
		lucky, err := consensus.Select(nextBlock.Header, chain)
		if err != nil {
			panic(fmt.Sprintf("Error seleccionando validadores: %s", err))
		}
		for i, satId := range lucky {
			nodo := miners.FindOwner(nodos, chain, satId)
			if nodo == nil {
				panic(fmt.Sprintf("SatID %d no tiene dueño", satId))
			}
			if i == 2 {
				nextBlock, err = nodo.ProposeBlock(nextBlock, chain, nil, dificultad)
			} else {
				nextBlock, err = nodo.SignBlock(nextBlock, chain)
			}
			if err != nil {
				panic(fmt.Sprintf("Error durante la firma del bloque: %s", err))
			}
		}

		// Verificamos el bloque hijo contra su padre inmediato
		if err = consensus.Verify(nextBlock, chain, dificultad); err != nil {
			panic(fmt.Sprintf("Falló la verificación en el bloque %d: %v", i, err))
		}

		// Avanzamos el puntero de la cadena para la siguiente iteración
		chain.Blocks[nextBlock.Header.Hash()] = nextBlock
		chain.Tip = nextBlock.Header.Hash()

		// Imprimimos un progreso cada 10 bloques para no saturar la terminal
		if i%10 == 0 || i == 1 {
			fmt.Printf("-> [Bloque %d] Minado y verificado con éxito (Hash: %x...)\n", i, nextBlock.Header.Hash())
		}
	}

	fmt.Printf("\n=== ¡Los %d bloques se minaron y verificaron de punta a punta sin errores! ===\n", totalBloques)
}
