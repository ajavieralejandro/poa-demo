package miners

import (
	"poa-demo/internal/blockchain"
)

type Address uint32
type SatoshiID uint32

type Miner struct {
	Addr     Address
	Satoshis []SatoshiID
	Blocks   []blockchain.Block
}

// Pool centraliza la gestión de todos los nodos mineros de la red.
type Pool struct {
	Miners []Miner
}

// Estructura auxiliar interna para ordenar los mineros por su hash de forma determinista
type minerHashPair struct {
	addr    Address
	hashVal uint64
}
