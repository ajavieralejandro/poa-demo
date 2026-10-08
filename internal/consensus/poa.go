// Package consensus reserva la minería y la elección de tres satoshis.
//
// La primera versión corre en un proceso y extiende una cadena lineal.
// Elegir un satoshi no depende de una lista de candidatos ni de quién esté
// conectado: el dueño se consulta después en el estado confirmado.
// Bifurcaciones, reorganización y red quedan fuera de este paquete.
package consensus

import (
	"errors"

	"poa-demo/internal/blockchain"
)

var (
	// ErrNotImplemented lo devuelven minería y selección.
	// Ninguna de las dos acepta un bloque ni devuelve un candidato válido.
	ErrNotImplemented = errors.New("consensus: not implemented")

	ErrEmptySupply = errors.New("consensus: no hay satoshis para elegir")
)

// Mine busca un nonce para la cabecera hija de tip.
// Cuando exista, devolverá solo la cabecera. No escribe en la cadena.
func Mine(tip blockchain.Block, miner blockchain.Address, bits int) (blockchain.Header, error) {
	return blockchain.Header{}, ErrNotImplemented
}

// Select deriva tres índices a partir de la cabecera, su padre y la posición.
// supply es la cantidad de satoshis existentes, con identificadores 0..supply-1.
// El mismo dueño puede ocupar más de un cupo. Un dueño desconectado sigue
// siendo elegible. El módulo tiene que descartar el resto del rango cuando
// supply no es potencia de dos, para no favorecer índices bajos.
func Select(h blockchain.Header, supply uint32) ([3]blockchain.SatoshiID, error) {
	if supply == 0 {
		return [3]blockchain.SatoshiID{}, ErrEmptySupply
	}
	return [3]blockchain.SatoshiID{}, ErrNotImplemented
}
