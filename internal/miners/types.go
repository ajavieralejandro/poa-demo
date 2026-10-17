package miners

import "poa-demo/internal/blockchain"

type Node struct {
	Addr     blockchain.Address
	Satoshis map[blockchain.SatoshiID]struct{}
}

// Pool centraliza la gestión de todos los nodos mineros de la red.
type Pool struct {
	Miners []*Node
}

// Estructura auxiliar interna para ordenar los mineros por su hash de forma determinista
type minerHashPair struct {
	addr    blockchain.Address
	hashVal uint64
}
