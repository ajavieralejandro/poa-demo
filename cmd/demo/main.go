package main

import (
	"fmt"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/consensus"
)

func main() {
	fmt.Println("=== Demo de PoA y Follow the Satoshi ===")

	// 1. Configuración inicial
	var supply uint32 = 1000
	dificultad := 3

	// 2. Creamos el bloque Génesis (Raíz de la cadena)
	genesis := blockchain.NewGenesisBlock()
	fmt.Printf("[Génesis] Creado exitosamente en Height: %d\n", genesis.Header.Height)

	// 3. Minamos el primer bloque hijo utilizando el Génesis como padre
	validadorMiner := blockchain.Address(1)
	nuevaHeader, err := consensus.Mine(genesis, validadorMiner, dificultad)
	if err != nil {
		panic(err)
	}

	// Armamos el bloque hijo usando la cabecera recién minada
	bloqueHijo := blockchain.Block{
		Header:     nuevaHeader,
		Signatures: [3]blockchain.Signature{}, // Aquí irían las firmas reales de los validadores
		Txs:        []blockchain.Tx{},
	}

	// Usamos la variable 'nuevaHeader' para imprimir su resumen (adiós error de compilación)
	fmt.Printf("[Bloque 1 Minado] %s\n", bloqueHijo.Summary())

	// 4. Ejecutamos el sorteo de Follow the Satoshi usando la cabecera del padre (Génesis)
	validadoresEsperados, err := consensus.Select(genesis.Header, supply)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[Sorteo PoA] Validadores elegidos para el bloque 1: %v\n", validadoresEsperados)

	// 5. Verificamos el bloque hijo contra su padre
	valido, err := consensus.Verify(bloqueHijo, genesis, dificultad, supply)
	if err != nil {
		fmt.Printf("[Verificación] Falló: %v\n", err)
	} else {
		fmt.Printf("[Verificación] ¿El bloque es válido? %v\n", valido)
	}

	fmt.Println("=== ¡Simulación finalizada con éxito! ===")
}
