package consensus

import (
	"errors"

	"poa-demo/internal/blockchain"
)

var (
	// ErrNotImplemented se usa temporalmente en componentes pendientes.
	ErrNotImplemented = errors.New("consensus: not implemented")

	// ErrEmptySupply indica que no hay unidades monetarias para realizar el sorteo.
	ErrEmptySupply = errors.New("consensus: no hay satoshis para elegir")
)

// Mine busca un nonce para la cabecera hija de tip (Fase 1: Proof of Work).
// Cuando encuentra un hash válido que cumple la dificultad, devuelve la cabecera.
func Mine(tip blockchain.Block, miner blockchain.Address, bits int) (blockchain.Header, error) {
	// 1. Preparamos la cabecera base conectándola con el bloque anterior (tip)
	prevHash := tip.Header.Hash()

	h := blockchain.Header{
		Prev:   prevHash,
		Miner:  miner,
		Height: tip.Header.Height + 1,
		Nonce:  0,
	}

	// 2. Bucle de búsqueda (Proof of Work)
	for {
		// Calculamos el hash de la cabecera con el nonce actual
		currentHash := h.Hash()

		// Verificamos si cumple con la dificultad exigida
		if cumpleDificultad(currentHash, bits) {
			return h, nil
		}

		// Si no cumple, incrementamos el nonce y probamos de nuevo
		h.Nonce++
	}
}

// cumpleDificultad evalúa si el hash cumple con el objetivo de bytes en cero.
func cumpleDificultad(hash blockchain.Hash, bits int) bool {
	bytesRequeridos := bits / 8
	if bytesRequeridos > len(hash) {
		return false
	}

	for i := 0; i < bytesRequeridos; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}

// Select deriva tres índices de satoshis a partir de la cabecera minada (Fase 2: Follow the Satoshi).
// supply es la cantidad de satoshis existentes (0 .. supply-1).
func Select(h blockchain.Header, supply uint32) ([3]blockchain.SatoshiID, error) {
	if supply == 0 {
		return [3]blockchain.SatoshiID{}, ErrEmptySupply
	}

	var out [3]blockchain.SatoshiID

	// Iteramos 3 veces para obtener los 3 cupos de validadores (posiciones 0, 1 y 2)
	for pos := uint8(0); pos < 3; pos++ {
		// TODO: 
		// 1. Combinar de forma determinista la cabecera (h) o su hash con la variable 'pos'.
		// 2. Convertir ese resultado a un número para aplicar módulo 'supply'.
		// 3. Asignar el SatoshiID resultante a out[pos].
	}

	return out, nil
}