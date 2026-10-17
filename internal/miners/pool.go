package miners

import (
	// TODO: Importar los paquetes necesarios para hash y conversión numérica
	// (ej: "crypto/sha256", "encoding/binary")
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"poa-demo/internal/blockchain"
	"sort"
)

var (
	ErrEmptySupply = errors.New("Error : empty suplly, no tokens. ")
)

// NewDeterministicPool inicializa el pool y reparte los satoshis del supply
// entre los mineros de forma determinista usando hashing de sus direcciones.
func NewDeterministicPool(chain blockchain.Chain, addresses []blockchain.Address) (blockchain.Chain, map[blockchain.Address]*Node) {
	numNodos := uint32(len(addresses))
	supply := chain.Blocks[chain.Tip].Header.Supply
	tokensPorMinero := supply / numNodos

	// 1. Calcular el hash criptográfico para cada dirección y pasarlo a uint64
	pairs := make([]minerHashPair, len(addresses))
	for i, addr := range addresses {
		// Convertir la dirección (uint32) a bytes para poder hacerle el hash
		addrBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(addrBytes, uint32(addr))

		// a) Aplicar función SHA-256
		hashSum := sha256.Sum256(addrBytes)

		// b) Convertir los primeros 8 bytes del resultado a un entero uint64
		hashVal := binary.BigEndian.Uint64(hashSum[0:8])

		pairs[i] = minerHashPair{
			addr:    addr,
			hashVal: hashVal,
		}
	}

	// 2. Ordenar los mineros según su valor de hash numérico.
	// Esto garantiza que todos los nodos de la red ordenen a los mineros exactamente igual.
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].hashVal < pairs[j].hashVal
	})

	// 3. Asignar los rangos de SatoshiIDs de manera contigua sin huecos
	nodos := make(map[blockchain.Address]*Node)
	for i := uint32(0); i < numNodos; i++ {
		inicio := i * tokensPorMinero

		// El último minero absorbe cualquier residuo matemático por redondeo del supply
		var fin uint32
		if i == numNodos-1 {
			fin = supply - 1
		} else {
			fin = (i+1)*tokensPorMinero - 1
		}

		// Rellenar la lista de SatoshiIDs correspondientes a este minero
		satoshis := make(map[blockchain.SatoshiID]struct{})
		for s := inicio; s <= fin; s++ {
			satoshis[blockchain.SatoshiID(s)] = struct{}{}
			chain.Owners[blockchain.SatoshiID(s)] = addresses[i]
		}

		nodos[addresses[i]] = NewNode(addresses[i], satoshis)
	}

	return chain, nodos
}

// FindOwner busca en el pool qué minero es el propietario de un SatoshiID dado.
func FindOwner(nodes map[blockchain.Address]*Node, chain blockchain.Chain, satID blockchain.SatoshiID) *Node {
	owner, ok := chain.Owners[satID]
	if !ok {
		return nil
	}
	return nodes[owner]
}
