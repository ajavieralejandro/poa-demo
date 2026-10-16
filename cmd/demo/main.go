package main

import (
	"fmt"
	"poa-demo/internal/blockchain"
	"poa-demo/internal/consensus"
	// TODO: Importa tus paquetes internos (blockchain y consensus)
)

func main() {
	fmt.Println("=== Demo de PoA y Follow the Satoshi ===")

	// 1. Define tus variables de configuración (supply y dificultad)
	var supply uint32 := 1000
	var dificultad uint32 := 3



	// 2. Inicializa tu bloque Génesis
	var genesis = blockachain.newGenesisBlock()



	// 3. Llama a la función de minería (Mine)
	

	// 4. Ejecuta el sorteo de validadores (Select)


	// 5. Verifica el bloque (Verify) e imprime el resultado final

}