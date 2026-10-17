package main

import (
	"fmt"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/consensus"
)

func main() {
	fmt.Println("=== Demo de PoA y Follow the Satoshi (Stress Test) ===")

	// 1. Configuración inicial
	var supply uint32 = 1000
	dificultad := 3
	validadorMiner := blockchain.Address(1)

	// 2. Creamos el bloque Génesis (Raíz de la cadena)
	currentBlock := blockchain.NewGenesisBlock()
	fmt.Printf("[Génesis] Creado en Height: %d\n", currentBlock.Header.Height)

	// 3. Bucle para simular y verificar 100 bloques encadenados
	totalBloques := 100
	for i := 1; i <= totalBloques; i++ {
		// Minamos el bloque hijo utilizando el bloque actual como padre (tip)
		nextBlock, err := consensus.Mine(currentBlock, validadorMiner, dificultad, supply)
		if err != nil {
			panic(fmt.Sprintf("Error minando el bloque %d: %v", i, err))
		}

		// Verificamos el bloque hijo contra su padre inmediato
		valido, err := consensus.Verify(nextBlock, currentBlock, dificultad, supply)
		if err != nil || !valido {
			panic(fmt.Sprintf("Falló la verificación en el bloque %d: %v", i, err))
		}

		// Avanzamos el puntero de la cadena para la siguiente iteración
		currentBlock = nextBlock

		// Imprimimos un progreso cada 10 bloques para no saturar la terminal
		if i%10 == 0 || i == 1 {
			fmt.Printf("-> [Bloque %d] Minado y verificado con éxito (Hash: %x...)\n", i, nextBlock.Header.Hash())
		}
	}

	fmt.Printf("\n=== ¡Los %d bloques se minaron y verificaron de punta a punta sin errores! ===\n", totalBloques)
}
